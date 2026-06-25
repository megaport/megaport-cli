package apply

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/commands/mve"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	defaultWaitTime = 10 * time.Minute
	statusError     = "error"

	// maxConfigFileSize caps how much of a config file is read into memory. Real
	// infra configs are kilobytes; this guards against pointing --file at a huge
	// or pathological file (e.g. /dev/zero) and exhausting memory.
	maxConfigFileSize = 10 * 1024 * 1024
)

// templateRe matches {{.type.name}} references in config values.
var templateRe = regexp.MustCompile(`\{\{\.(\w+)\.([^}]+)\}\}`)

// deleteCLICommand maps resource type to the CLI subcommand used to delete it.
var deleteCLICommand = map[string]string{
	"Port": "ports",
	"MCR":  "mcr",
	"MVE":  "mve",
	"VXC":  "vxc",
}

// createdResource records a resource that was successfully provisioned during an apply run.
type createdResource struct {
	resType string // "Port", "MCR", "MVE", or "VXC"
	name    string
	uid     string
}

// ApplyConfig is the entry point for `megaport-cli apply`.
func ApplyConfig(cmd *cobra.Command, _ []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	filePath, _ := cmd.Flags().GetString("file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	rollback, _ := cmd.Flags().GetBool("rollback-on-failure")

	if filePath == "" {
		output.PrintError("--file is required", noColor)
		return exitcodes.NewUsageError(fmt.Errorf("--file is required"))
	}

	cfg, err := parseConfigFile(filePath)
	if err != nil {
		output.PrintError("Failed to parse config file: %v", noColor, err)
		return err
	}

	// provisionTimeout is the per-resource provisioning budget; rollbackTimeout
	// reuses it for a fresh rollback context.
	provisionTimeout := utils.TimeoutFromCmd(cmd, defaultWaitTime)
	rollbackTimeout := provisionTimeout

	// No run-wide deadline: each resource's provisioning wait applies
	// provisionTimeout on its own context, so --timeout bounds each resource rather
	// than the whole run sharing one budget. Individual API calls stay bounded by
	// the SDK HTTP client timeout.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spinner := output.PrintLoggingInWithOutput(noColor, outputFormat)
	client, err := config.Login(ctx)
	if err != nil {
		spinner.Stop()
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	spinner.Stop()

	if dryRun {
		return validateAll(ctx, client, cfg, noColor, outputFormat)
	}

	total := len(cfg.Ports) + len(cfg.MCRs) + len(cfg.MVEs) + len(cfg.VXCs)
	if total == 0 {
		output.PrintInfo("Config file contains no resources to provision.", noColor)
		return nil
	}

	if !yes {
		output.PrintInfo("Resources to provision:", noColor)
		output.PrintInfo("  Ports: %d, MCRs: %d, MVEs: %d, VXCs: %d", noColor,
			len(cfg.Ports), len(cfg.MCRs), len(cfg.MVEs), len(cfg.VXCs))
		if !utils.ConfirmPrompt("Proceed with provisioning?", noColor) {
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	// uids["port"]["Sydney-Primary"] = "provisioned-uid"
	uids := map[string]map[string]string{
		"port": {},
		"mcr":  {},
		"mve":  {},
		"vxc":  {},
	}
	var results []ApplyResult
	var created []createdResource // tracks successfully provisioned resources for orphan reporting

	// 1. Ports
	for _, p := range cfg.Ports {
		req := &megaport.BuyPortRequest{
			Name:                  p.Name,
			LocationId:            p.LocationID,
			PortSpeed:             p.Speed,
			Term:                  p.Term,
			MarketPlaceVisibility: p.MarketplaceVisibility,
			DiversityZone:         p.DiversityZone,
			CostCentre:            p.CostCentre,
			ResourceTags:          p.ResourceTags,
			WaitForProvision:      false,
		}
		if err := validation.ValidatePortRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("validation failed for port %q: %w", p.Name, err))
		}
		validateSpinner := output.PrintResourceValidating("Port", noColor)
		if err := client.PortService.ValidatePortOrder(ctx, req); err != nil {
			validateSpinner.Stop()
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("server-side validation failed for port %q: %w", p.Name, err))
		}
		validateSpinner.Stop()

		createSpinner := output.PrintResourceCreating("Port", p.Name, noColor)
		var resp *megaport.BuyPortResponse
		err = utils.WithOrderOnceRetry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.PortService.BuyPort(ctx, req)
			return e
		})
		if err != nil {
			createSpinner.Stop()
			err = utils.WrapAPIError(err, "Port", p.Name)
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision port %q: %w", p.Name, err))
		}
		if resp == nil {
			createSpinner.Stop()
			err := fmt.Errorf("empty response from API")
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision port %q: %w", p.Name, err))
		}
		if len(resp.TechnicalServiceUIDs) == 0 {
			createSpinner.Stop()
			err := fmt.Errorf("API returned no UID")
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision port %q: %w", p.Name, err))
		}
		uid := resp.TechnicalServiceUIDs[0]
		uids["port"][p.Name] = uid
		// Track immediately: the order is placed and billing has started, even
		// though provisioning has not completed. If the wait below fails, the
		// resource must still be visible to rollback/orphan reporting.
		created = append(created, createdResource{resType: "Port", name: p.Name, uid: uid})
		if err := waitForProvision(ctx, provisionTimeout, "Port", p.Name, uid, func(ctx context.Context) (string, error) {
			port, e := client.PortService.GetPort(ctx, uid)
			if e != nil {
				return "", e
			}
			if port == nil {
				return "", fmt.Errorf("empty response from API")
			}
			return port.ProvisioningStatus, nil
		}); err != nil {
			createSpinner.Stop()
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision port %q: %w", p.Name, err))
		}
		createSpinner.Stop()
		results = append(results, ApplyResult{Type: "Port", Name: p.Name, UID: uid, Status: "provisioned"})
		output.PrintResourceCreated("Port", uid, noColor)
	}

	// 2. MCRs
	for _, m := range cfg.MCRs {
		req := &megaport.BuyMCRRequest{
			Name:             m.Name,
			LocationID:       m.LocationID,
			PortSpeed:        m.Speed,
			Term:             m.Term,
			MCRAsn:           m.ASN,
			DiversityZone:    m.DiversityZone,
			CostCentre:       m.CostCentre,
			ResourceTags:     m.ResourceTags,
			WaitForProvision: false,
		}
		if err := validation.ValidateMCRRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("validation failed for MCR %q: %w", m.Name, err))
		}
		validateSpinner := output.PrintResourceValidating("MCR", noColor)
		if err := client.MCRService.ValidateMCROrder(ctx, req); err != nil {
			validateSpinner.Stop()
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("server-side validation failed for MCR %q: %w", m.Name, err))
		}
		validateSpinner.Stop()

		createSpinner := output.PrintResourceCreating("MCR", m.Name, noColor)
		var resp *megaport.BuyMCRResponse
		err = utils.WithOrderOnceRetry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.MCRService.BuyMCR(ctx, req)
			return e
		})
		if err != nil {
			createSpinner.Stop()
			err = utils.WrapAPIError(err, "MCR", m.Name)
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision MCR %q: %w", m.Name, err))
		}
		if resp == nil {
			createSpinner.Stop()
			err := fmt.Errorf("empty response from API")
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision MCR %q: %w", m.Name, err))
		}
		uid := strings.TrimSpace(resp.TechnicalServiceUID)
		if uid == "" {
			createSpinner.Stop()
			err := fmt.Errorf("API returned empty UID")
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision MCR %q: %w", m.Name, err))
		}
		uids["mcr"][m.Name] = uid
		created = append(created, createdResource{resType: "MCR", name: m.Name, uid: uid})
		if err := waitForProvision(ctx, provisionTimeout, "MCR", m.Name, uid, func(ctx context.Context) (string, error) {
			mcr, e := client.MCRService.GetMCR(ctx, uid)
			if e != nil {
				return "", e
			}
			if mcr == nil {
				return "", fmt.Errorf("empty response from API")
			}
			return mcr.ProvisioningStatus, nil
		}); err != nil {
			createSpinner.Stop()
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision MCR %q: %w", m.Name, err))
		}
		createSpinner.Stop()
		results = append(results, ApplyResult{Type: "MCR", Name: m.Name, UID: uid, Status: "provisioned"})
		output.PrintResourceCreated("MCR", uid, noColor)
	}

	// 3. MVEs
	for _, mv := range cfg.MVEs {
		normalizedVC, err := normalizeVendorConfigMap(mv.VendorConfig)
		if err != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("invalid vendor_config for MVE %q: %w", mv.Name, err))
		}
		vendorCfg, err := mve.ParseVendorConfig(normalizedVC)
		if err != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("invalid vendor_config for MVE %q: %w", mv.Name, err))
		}
		req := &megaport.BuyMVERequest{
			Name:             mv.Name,
			LocationID:       mv.LocationID,
			Term:             mv.Term,
			VendorConfig:     vendorCfg,
			DiversityZone:    mv.DiversityZone,
			CostCentre:       mv.CostCentre,
			ResourceTags:     mv.ResourceTags,
			WaitForProvision: false,
		}
		if err := validation.ValidateBuyMVERequest(req); err != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("validation failed for MVE %q: %w", mv.Name, err))
		}
		validateSpinner := output.PrintResourceValidating("MVE", noColor)
		if err := client.MVEService.ValidateMVEOrder(ctx, req); err != nil {
			validateSpinner.Stop()
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("server-side validation failed for MVE %q: %w", mv.Name, err))
		}
		validateSpinner.Stop()

		createSpinner := output.PrintResourceCreating("MVE", mv.Name, noColor)
		var resp *megaport.BuyMVEResponse
		err = utils.WithOrderOnceRetry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.MVEService.BuyMVE(ctx, req)
			return e
		})
		if err != nil {
			createSpinner.Stop()
			err = utils.WrapAPIError(err, "MVE", mv.Name)
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision MVE %q: %w", mv.Name, err))
		}
		if resp == nil {
			createSpinner.Stop()
			err := fmt.Errorf("empty response from API")
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision MVE %q: %w", mv.Name, err))
		}
		uid := strings.TrimSpace(resp.TechnicalServiceUID)
		if uid == "" {
			createSpinner.Stop()
			err := fmt.Errorf("API returned empty UID")
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision MVE %q: %w", mv.Name, err))
		}
		uids["mve"][mv.Name] = uid
		created = append(created, createdResource{resType: "MVE", name: mv.Name, uid: uid})
		if err := waitForProvision(ctx, provisionTimeout, "MVE", mv.Name, uid, func(ctx context.Context) (string, error) {
			m, e := client.MVEService.GetMVE(ctx, uid)
			if e != nil {
				return "", e
			}
			if m == nil {
				return "", fmt.Errorf("empty response from API")
			}
			return m.ProvisioningStatus, nil
		}); err != nil {
			createSpinner.Stop()
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision MVE %q: %w", mv.Name, err))
		}
		createSpinner.Stop()
		results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, UID: uid, Status: "provisioned"})
		output.PrintResourceCreated("MVE", uid, noColor)
	}

	// 4. VXCs — resolve {{.type.name}} templates before provisioning
	for _, v := range cfg.VXCs {
		aUID, err := resolveTemplates(v.AEnd.ProductUID, uids)
		if err != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("unresolved template in VXC %q a_end: %w", v.Name, err))
		}
		bUID, err := resolveTemplates(v.BEnd.ProductUID, uids)
		if err != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("unresolved template in VXC %q b_end: %w", v.Name, err))
		}
		req := &megaport.BuyVXCRequest{
			PortUID:   aUID,
			VXCName:   v.Name,
			RateLimit: v.RateLimit,
			Term:      v.Term,
			AEndConfiguration: megaport.VXCOrderEndpointConfiguration{
				ProductUID: aUID,
				VLAN:       v.AEnd.VLAN,
			},
			BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
				ProductUID: bUID,
				VLAN:       v.BEnd.VLAN,
			},
			CostCentre:       v.CostCentre,
			ResourceTags:     v.ResourceTags,
			WaitForProvision: false,
		}
		if err := validation.ValidateVXCRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("validation failed for VXC %q: %w", v.Name, err))
		}
		validateSpinner := output.PrintResourceValidating("VXC", noColor)
		if err := client.VXCService.ValidateVXCOrder(ctx, req); err != nil {
			validateSpinner.Stop()
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("server-side validation failed for VXC %q: %w", v.Name, err))
		}
		validateSpinner.Stop()

		createSpinner := output.PrintResourceCreating("VXC", v.Name, noColor)
		var resp *megaport.BuyVXCResponse
		err = utils.WithOrderOnceRetry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.VXCService.BuyVXC(ctx, req)
			return e
		})
		if err != nil {
			createSpinner.Stop()
			err = utils.WrapAPIError(err, "VXC", v.Name)
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision VXC %q: %w", v.Name, err))
		}
		if resp == nil {
			createSpinner.Stop()
			err := fmt.Errorf("empty response from API")
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision VXC %q: %w", v.Name, err))
		}
		uid := strings.TrimSpace(resp.TechnicalServiceUID)
		if uid == "" {
			createSpinner.Stop()
			err := fmt.Errorf("API returned empty UID")
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision VXC %q: %w", v.Name, err))
		}
		uids["vxc"][v.Name] = uid
		created = append(created, createdResource{resType: "VXC", name: v.Name, uid: uid})
		if err := waitForProvision(ctx, provisionTimeout, "VXC", v.Name, uid, func(ctx context.Context) (string, error) {
			vxc, e := client.VXCService.GetVXC(ctx, uid)
			if e != nil {
				return "", e
			}
			if vxc == nil {
				return "", fmt.Errorf("empty response from API")
			}
			return vxc.ProvisioningStatus, nil
		}); err != nil {
			createSpinner.Stop()
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return handleFailure(client, created, results, outputFormat, noColor, rollback, rollbackTimeout,
				fmt.Errorf("failed to provision VXC %q: %w", v.Name, err))
		}
		createSpinner.Stop()
		results = append(results, ApplyResult{Type: "VXC", Name: v.Name, UID: uid, Status: "provisioned"})
		output.PrintResourceCreated("VXC", uid, noColor)
	}

	return output.PrintOutput(results, outputFormat, noColor)
}

// handleFailure prints results, the failure error, and — when resources were already
// provisioned — either attempts rollback or prints a prominent billing warning with
// exact remediation commands.
func handleFailure(client *megaport.Client, created []createdResource, results []ApplyResult, outputFormat string, noColor bool, rollback bool, rollbackTimeout time.Duration, failErr error) error {
	_ = output.PrintOutput(results, outputFormat, noColor)

	jsonMode := outputFormat == "json"

	if !jsonMode {
		output.PrintError("Apply failed: %v", noColor, failErr)
	}

	if len(created) == 0 {
		return failErr
	}

	if rollback {
		return doRollback(client, created, jsonMode, noColor, rollbackTimeout, failErr)
	}

	if jsonMode {
		parts := []string{"resources created and ARE BILLING:"}
		for _, r := range created {
			parts = append(parts, fmt.Sprintf("%s %q uid: %s; to remove: megaport-cli %s delete %s", r.resType, r.name, r.uid, deleteCLICommand[r.resType], r.uid))
		}
		return fmt.Errorf("%w; %s", failErr, strings.Join(parts, "; "))
	}

	output.PrintError("The following resources were created and ARE BILLING:", noColor)
	for _, r := range created {
		output.PrintError("  %s %q  uid: %s", noColor, r.resType, r.name, r.uid)
		output.PrintError("  To remove: megaport-cli %s delete %s", noColor, deleteCLICommand[r.resType], r.uid)
	}
	return failErr
}

// doRollback deletes created resources in reverse provisioning order. It starts a
// fresh context rather than reusing the provisioning context: the most common
// rollback trigger is a provisioning timeout, which leaves that context already
// expired — reusing it would make every delete fail with "context deadline
// exceeded" and orphan the billing resources rollback exists to clean up. The
// fresh context uses the same timeout the user configured for the run.
func doRollback(client *megaport.Client, created []createdResource, jsonMode bool, noColor bool, rollbackTimeout time.Duration, failErr error) error {
	ctx, cancel := context.WithTimeout(context.Background(), rollbackTimeout)
	defer cancel()

	if !jsonMode {
		output.PrintWarning("Rolling back %d created resource(s)...", noColor, len(created))
	}
	var rollbackResults []string
	for _, r := range slices.Backward(created) {
		err := utils.WithRetry(ctx, func(ctx context.Context) error {
			return deleteResource(ctx, client, r)
		})
		if err != nil {
			if jsonMode {
				rollbackResults = append(rollbackResults, fmt.Sprintf("rollback failed for %s %q (%s): %v; to remove: megaport-cli %s delete %s", r.resType, r.name, r.uid, err, deleteCLICommand[r.resType], r.uid))
			} else {
				output.PrintError("Rollback failed for %s %q (%s): %v", noColor, r.resType, r.name, r.uid, err)
				output.PrintError("  To remove manually: megaport-cli %s delete %s", noColor, deleteCLICommand[r.resType], r.uid)
			}
		} else {
			if jsonMode {
				rollbackResults = append(rollbackResults, fmt.Sprintf("rolled back %s %q (%s)", r.resType, r.name, r.uid))
			} else {
				output.PrintSuccess("Rolled back %s %q (%s)", noColor, r.resType, r.name, r.uid)
			}
		}
	}
	if jsonMode {
		return fmt.Errorf("%w; %s", failErr, strings.Join(rollbackResults, "; "))
	}
	return failErr
}

// waitForProvision applies a per-resource timeout to ctx so --timeout bounds each
// resource rather than the whole run, then delegates to the shared poll loop. The
// order has already been placed by the time this runs, so the caller must have
// recorded the resource as created before calling this.
func waitForProvision(ctx context.Context, timeout time.Duration, resType, name, uid string, getStatus func(ctx context.Context) (string, error)) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return utils.WaitForProvision(ctx, resType, name, uid, getStatus)
}

// deleteResource deletes a single provisioned resource via the appropriate service client.
func deleteResource(ctx context.Context, client *megaport.Client, r createdResource) error {
	switch r.resType {
	case "Port":
		_, err := client.PortService.DeletePort(ctx, &megaport.DeletePortRequest{PortID: r.uid, DeleteNow: true})
		return err
	case "MCR":
		_, err := client.MCRService.DeleteMCR(ctx, &megaport.DeleteMCRRequest{MCRID: r.uid, DeleteNow: true})
		return err
	case "MVE":
		_, err := client.MVEService.DeleteMVE(ctx, &megaport.DeleteMVERequest{MVEID: r.uid})
		return err
	case "VXC":
		return client.VXCService.DeleteVXC(ctx, r.uid, &megaport.DeleteVXCRequest{DeleteNow: true})
	default:
		return fmt.Errorf("unknown resource type %q", r.resType)
	}
}

// validateAll runs SDK-level validation for every resource without provisioning.
// Requests mirror provisioning exactly (minus WaitForProvision/WaitForTime).
func validateAll(ctx context.Context, client *megaport.Client, cfg *InfraConfig, noColor bool, outputFormat string) error {
	var results []ApplyResult

	for _, p := range cfg.Ports {
		req := &megaport.BuyPortRequest{
			Name:                  p.Name,
			LocationId:            p.LocationID,
			PortSpeed:             p.Speed,
			Term:                  p.Term,
			MarketPlaceVisibility: p.MarketplaceVisibility,
			DiversityZone:         p.DiversityZone,
			CostCentre:            p.CostCentre,
			ResourceTags:          p.ResourceTags,
		}
		if err := validation.ValidatePortRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: "invalid: " + err.Error()})
			continue
		}
		err := client.PortService.ValidatePortOrder(ctx, req)
		status := "valid"
		if err != nil {
			status = "invalid: " + err.Error()
		}
		results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: status})
	}

	for _, m := range cfg.MCRs {
		req := &megaport.BuyMCRRequest{
			Name:          m.Name,
			LocationID:    m.LocationID,
			PortSpeed:     m.Speed,
			Term:          m.Term,
			MCRAsn:        m.ASN,
			DiversityZone: m.DiversityZone,
			CostCentre:    m.CostCentre,
			ResourceTags:  m.ResourceTags,
		}
		if err := validation.ValidateMCRRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: "invalid: " + err.Error()})
			continue
		}
		err := client.MCRService.ValidateMCROrder(ctx, req)
		status := "valid"
		if err != nil {
			status = "invalid: " + err.Error()
		}
		results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: status})
	}

	for _, mv := range cfg.MVEs {
		normalizedVC, vcErr := normalizeVendorConfigMap(mv.VendorConfig)
		if vcErr != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: "invalid: " + vcErr.Error()})
			continue
		}
		vendorCfg, vcErr := mve.ParseVendorConfig(normalizedVC)
		if vcErr != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: "invalid: " + vcErr.Error()})
			continue
		}
		req := &megaport.BuyMVERequest{
			Name:          mv.Name,
			LocationID:    mv.LocationID,
			Term:          mv.Term,
			VendorConfig:  vendorCfg,
			DiversityZone: mv.DiversityZone,
			CostCentre:    mv.CostCentre,
			ResourceTags:  mv.ResourceTags,
		}
		if err := validation.ValidateBuyMVERequest(req); err != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: "invalid: " + err.Error()})
			continue
		}
		err := client.MVEService.ValidateMVEOrder(ctx, req)
		status := "valid"
		if err != nil {
			status = "invalid: " + err.Error()
		}
		results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: status})
	}

	// Build a dry-run UID map from declared resources so template references
	// are resolved against known names — typos produce an "invalid" result
	// rather than silently passing with a generic placeholder.
	dryRunUIDs := map[string]map[string]string{
		"port": {},
		"mcr":  {},
		"mve":  {},
		"vxc":  {},
	}
	const dryRunPlaceholder = "00000000-0000-0000-0000-000000000000"
	for _, p := range cfg.Ports {
		dryRunUIDs["port"][p.Name] = dryRunPlaceholder
	}
	for _, m := range cfg.MCRs {
		dryRunUIDs["mcr"][m.Name] = dryRunPlaceholder
	}
	for _, mv := range cfg.MVEs {
		dryRunUIDs["mve"][mv.Name] = dryRunPlaceholder
	}

	for _, v := range cfg.VXCs {
		// Resolve templates against declared resources; literal UIDs pass through.
		aUID, aErr := resolveTemplates(v.AEnd.ProductUID, dryRunUIDs)
		bUID, bErr := resolveTemplates(v.BEnd.ProductUID, dryRunUIDs)
		if aErr != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: "invalid: " + aErr.Error()})
			continue
		}
		if bErr != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: "invalid: " + bErr.Error()})
			continue
		}
		req := &megaport.BuyVXCRequest{
			PortUID:   aUID,
			VXCName:   v.Name,
			RateLimit: v.RateLimit,
			Term:      v.Term,
			AEndConfiguration: megaport.VXCOrderEndpointConfiguration{
				ProductUID: aUID,
				VLAN:       v.AEnd.VLAN,
			},
			BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
				ProductUID: bUID,
				VLAN:       v.BEnd.VLAN,
			},
			CostCentre:   v.CostCentre,
			ResourceTags: v.ResourceTags,
		}
		if err := validation.ValidateVXCRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: "invalid: " + err.Error()})
			continue
		}
		// Skip server-side validation when either endpoint came from a template
		// (placeholder UID) — the real UID only exists after provisioning.
		if aUID == dryRunPlaceholder || bUID == dryRunPlaceholder {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: "skipped: requires provisioning"})
			continue
		}
		err := client.VXCService.ValidateVXCOrder(ctx, req)
		status := "valid"
		if err != nil {
			status = "invalid: " + err.Error()
		}
		results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: status})
	}

	output.PrintInfo("Dry-run: validation results", noColor)
	return output.PrintOutput(results, outputFormat, noColor)
}

// parseConfigFile reads filePath and decodes it into InfraConfig.
// It detects YAML vs JSON by file extension.
func parseConfigFile(filePath string) (*InfraConfig, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	defer f.Close()

	// Read one byte past the cap so an over-limit file is detected rather than
	// silently truncated.
	data, err := io.ReadAll(io.LimitReader(f, maxConfigFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	if len(data) > maxConfigFileSize {
		return nil, fmt.Errorf("config file %q exceeds maximum size of %d bytes", filePath, maxConfigFileSize)
	}

	cfg := &InfraConfig{}
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing JSON config: %w", err)
		}
	default: // .yaml, .yml, or anything else
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing YAML config: %w", err)
		}
	}
	return cfg, nil
}

// resolveTemplates replaces {{.type.name}} placeholders using the uids map.
func resolveTemplates(s string, uids map[string]map[string]string) (string, error) {
	var resolveErr error
	result := templateRe.ReplaceAllStringFunc(s, func(match string) string {
		sub := templateRe.FindStringSubmatch(match)
		if len(sub) != 3 {
			return match
		}
		resType, resName := sub[1], sub[2]
		if typeMap, ok := uids[resType]; ok {
			if uid, ok := typeMap[resName]; ok {
				return uid
			}
		}
		resolveErr = errors.Join(resolveErr, fmt.Errorf("no UID found for reference %q (type=%q name=%q)", match, resType, resName))
		return match
	})
	return result, resolveErr
}

// normalizeVendorConfigMap round-trips a vendor config map through JSON so that
// YAML-decoded integers (int) become float64, matching what ParseVendorConfig expects.
func normalizeVendorConfigMap(m map[string]interface{}) (map[string]interface{}, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("normalizing vendor config: %w", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("normalizing vendor config: %w", err)
	}
	return out, nil
}
