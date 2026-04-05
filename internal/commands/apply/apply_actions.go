package apply

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
)

// templateRe matches {{.type.name}} references in config values.
var templateRe = regexp.MustCompile(`\{\{\.(\w+)\.([^}]+)\}\}`)

// ApplyConfig is the entry point for `megaport apply`.
func ApplyConfig(cmd *cobra.Command, _ []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	filePath, _ := cmd.Flags().GetString("file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")

	if filePath == "" {
		output.PrintError("--file is required", noColor)
		return exitcodes.NewUsageError(fmt.Errorf("--file is required"))
	}

	cfg, err := parseConfigFile(filePath)
	if err != nil {
		output.PrintError("Failed to parse config file: %v", noColor, err)
		return err
	}

	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, defaultWaitTime)
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
			WaitForProvision:      true,
			WaitForTime:           defaultWaitTime,
		}
		if err := validation.ValidatePortRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("validation failed for port %q: %w", p.Name, err))
		}
		validateSpinner := output.PrintResourceValidating("Port", noColor)
		if err := client.PortService.ValidatePortOrder(ctx, req); err != nil {
			validateSpinner.Stop()
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("server-side validation failed for port %q: %w", p.Name, err))
		}
		validateSpinner.Stop()

		createSpinner := output.PrintResourceCreating("Port", p.Name, noColor)
		var resp *megaport.BuyPortResponse
		err = utils.WithRetry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.PortService.BuyPort(ctx, req)
			return e
		})
		if err != nil {
			createSpinner.Stop()
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("failed to provision port %q: %w", p.Name, err))
		}
		createSpinner.Stop()
		if len(resp.TechnicalServiceUIDs) == 0 {
			err := fmt.Errorf("API returned no UID")
			results = append(results, ApplyResult{Type: "Port", Name: p.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("failed to provision port %q: %w", p.Name, err))
		}
		uid := resp.TechnicalServiceUIDs[0]
		uids["port"][p.Name] = uid
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
			WaitForProvision: true,
			WaitForTime:      defaultWaitTime,
		}
		if err := validation.ValidateMCRRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("validation failed for MCR %q: %w", m.Name, err))
		}
		validateSpinner := output.PrintResourceValidating("MCR", noColor)
		if err := client.MCRService.ValidateMCROrder(ctx, req); err != nil {
			validateSpinner.Stop()
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("server-side validation failed for MCR %q: %w", m.Name, err))
		}
		validateSpinner.Stop()

		createSpinner := output.PrintResourceCreating("MCR", m.Name, noColor)
		var resp *megaport.BuyMCRResponse
		err = utils.WithRetry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.MCRService.BuyMCR(ctx, req)
			return e
		})
		if err != nil {
			createSpinner.Stop()
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("failed to provision MCR %q: %w", m.Name, err))
		}
		createSpinner.Stop()
		uid := strings.TrimSpace(resp.TechnicalServiceUID)
		if uid == "" {
			err := fmt.Errorf("API returned empty UID")
			results = append(results, ApplyResult{Type: "MCR", Name: m.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("failed to provision MCR %q: %w", m.Name, err))
		}
		uids["mcr"][m.Name] = uid
		results = append(results, ApplyResult{Type: "MCR", Name: m.Name, UID: uid, Status: "provisioned"})
		output.PrintResourceCreated("MCR", uid, noColor)
	}

	// 3. MVEs
	for _, mv := range cfg.MVEs {
		normalizedVC, err := normalizeVendorConfigMap(mv.VendorConfig)
		if err != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("invalid vendor_config for MVE %q: %w", mv.Name, err))
		}
		vendorCfg, err := mve.ParseVendorConfig(normalizedVC)
		if err != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
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
			WaitForProvision: true,
			WaitForTime:      defaultWaitTime,
		}
		if err := validation.ValidateBuyMVERequest(req); err != nil {
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("validation failed for MVE %q: %w", mv.Name, err))
		}
		validateSpinner := output.PrintResourceValidating("MVE", noColor)
		if err := client.MVEService.ValidateMVEOrder(ctx, req); err != nil {
			validateSpinner.Stop()
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("server-side validation failed for MVE %q: %w", mv.Name, err))
		}
		validateSpinner.Stop()

		createSpinner := output.PrintResourceCreating("MVE", mv.Name, noColor)
		var resp *megaport.BuyMVEResponse
		err = utils.WithRetry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.MVEService.BuyMVE(ctx, req)
			return e
		})
		if err != nil {
			createSpinner.Stop()
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("failed to provision MVE %q: %w", mv.Name, err))
		}
		createSpinner.Stop()
		uid := strings.TrimSpace(resp.TechnicalServiceUID)
		if uid == "" {
			err := fmt.Errorf("API returned empty UID")
			results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("failed to provision MVE %q: %w", mv.Name, err))
		}
		uids["mve"][mv.Name] = uid
		results = append(results, ApplyResult{Type: "MVE", Name: mv.Name, UID: uid, Status: "provisioned"})
		output.PrintResourceCreated("MVE", uid, noColor)
	}

	// 4. VXCs — resolve {{.type.name}} templates before provisioning
	for _, v := range cfg.VXCs {
		aUID, err := resolveTemplates(v.AEnd.ProductUID, uids)
		if err != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("unresolved template in VXC %q a_end: %w", v.Name, err))
		}
		bUID, err := resolveTemplates(v.BEnd.ProductUID, uids)
		if err != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
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
			WaitForProvision: true,
			WaitForTime:      defaultWaitTime,
		}
		if err := validation.ValidateVXCRequest(req); err != nil {
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("validation failed for VXC %q: %w", v.Name, err))
		}
		validateSpinner := output.PrintResourceValidating("VXC", noColor)
		if err := client.VXCService.ValidateVXCOrder(ctx, req); err != nil {
			validateSpinner.Stop()
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("server-side validation failed for VXC %q: %w", v.Name, err))
		}
		validateSpinner.Stop()

		createSpinner := output.PrintResourceCreating("VXC", v.Name, noColor)
		var resp *megaport.BuyVXCResponse
		err = utils.WithRetry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.VXCService.BuyVXC(ctx, req)
			return e
		})
		if err != nil {
			createSpinner.Stop()
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("failed to provision VXC %q: %w", v.Name, err))
		}
		createSpinner.Stop()
		uid := strings.TrimSpace(resp.TechnicalServiceUID)
		if uid == "" {
			err := fmt.Errorf("API returned empty UID")
			results = append(results, ApplyResult{Type: "VXC", Name: v.Name, Status: statusError + ": " + err.Error()})
			return printResultsAndError(results, outputFormat, noColor,
				fmt.Errorf("failed to provision VXC %q: %w", v.Name, err))
		}
		uids["vxc"][v.Name] = uid
		results = append(results, ApplyResult{Type: "VXC", Name: v.Name, UID: uid, Status: "provisioned"})
		output.PrintResourceCreated("VXC", uid, noColor)
	}

	return output.PrintOutput(results, outputFormat, noColor)
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
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
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
		resolveErr = fmt.Errorf("no UID found for reference %q (type=%q name=%q)", match, resType, resName)
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

// printResultsAndError prints the partial results table then returns err.
func printResultsAndError(results []ApplyResult, outputFormat string, noColor bool, err error) error {
	_ = output.PrintOutput(results, outputFormat, noColor)
	output.PrintError("Apply failed: %v", noColor, err)
	return err
}
