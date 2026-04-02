package status

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var listPortsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Port, error) {
	return client.PortService.ListPorts(ctx)
}

var listMCRsFunc = func(ctx context.Context, client *megaport.Client, includeInactive bool) ([]*megaport.MCR, error) {
	return client.MCRService.ListMCRs(ctx, &megaport.ListMCRsRequest{IncludeInactive: includeInactive})
}

var listMVEsFunc = func(ctx context.Context, client *megaport.Client, includeInactive bool) ([]*megaport.MVE, error) {
	return client.MVEService.ListMVEs(ctx, &megaport.ListMVEsRequest{IncludeInactive: includeInactive})
}

var listVXCsFunc = func(ctx context.Context, client *megaport.Client, includeInactive bool) ([]*megaport.VXC, error) {
	return client.VXCService.ListVXCs(ctx, &megaport.ListVXCsRequest{IncludeInactive: includeInactive})
}

var listIXsFunc = func(ctx context.Context, client *megaport.Client, includeInactive bool) ([]*megaport.IX, error) {
	return client.IXService.ListIXs(ctx, &megaport.ListIXsRequest{IncludeInactive: includeInactive})
}

// StatusDashboard fetches all resources in parallel and renders a dashboard view.
func StatusDashboard(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	spinner := output.PrintResourceListing("resource", noColor)

	var (
		mu    sync.Mutex
		errs  []error
		ports []*megaport.Port
		mcrs  []*megaport.MCR
		mves  []*megaport.MVE
		vxcs  []*megaport.VXC
		ixs   []*megaport.IX
		wg    sync.WaitGroup
	)

	wg.Add(5)

	go func() {
		defer wg.Done()
		result, fetchErr := listPortsFunc(ctx, client)
		mu.Lock()
		defer mu.Unlock()
		if fetchErr != nil {
			errs = append(errs, fmt.Errorf("ports: %v", fetchErr))
		} else {
			ports = result
		}
	}()

	go func() {
		defer wg.Done()
		result, fetchErr := listMCRsFunc(ctx, client, includeInactive)
		mu.Lock()
		defer mu.Unlock()
		if fetchErr != nil {
			errs = append(errs, fmt.Errorf("MCRs: %v", fetchErr))
		} else {
			mcrs = result
		}
	}()

	go func() {
		defer wg.Done()
		result, fetchErr := listMVEsFunc(ctx, client, includeInactive)
		mu.Lock()
		defer mu.Unlock()
		if fetchErr != nil {
			errs = append(errs, fmt.Errorf("MVEs: %v", fetchErr))
		} else {
			mves = result
		}
	}()

	go func() {
		defer wg.Done()
		result, fetchErr := listVXCsFunc(ctx, client, includeInactive)
		mu.Lock()
		defer mu.Unlock()
		if fetchErr != nil {
			errs = append(errs, fmt.Errorf("VXCs: %v", fetchErr))
		} else {
			vxcs = result
		}
	}()

	go func() {
		defer wg.Done()
		result, fetchErr := listIXsFunc(ctx, client, includeInactive)
		mu.Lock()
		defer mu.Unlock()
		if fetchErr != nil {
			errs = append(errs, fmt.Errorf("IXs: %v", fetchErr))
		} else {
			ixs = result
		}
	}()

	wg.Wait()
	spinner.Stop()

	if len(errs) > 0 {
		for _, e := range errs {
			output.PrintError("Failed to fetch %v", noColor, e)
		}
		return errs[0]
	}

	// Filter inactive ports client-side (PortService.ListPorts has no IncludeInactive param).
	if !includeInactive {
		var activePorts []*megaport.Port
		for _, p := range ports {
			if p != nil &&
				p.ProvisioningStatus != megaport.STATUS_DECOMMISSIONED &&
				p.ProvisioningStatus != megaport.STATUS_CANCELLED &&
				p.ProvisioningStatus != "DECOMMISSIONING" {
				activePorts = append(activePorts, p)
			}
		}
		ports = activePorts
	}

	dashboard, err := buildDashboard(ports, mcrs, mves, vxcs, ixs)
	if err != nil {
		output.PrintError("Failed to build dashboard: %v", noColor, err)
		return fmt.Errorf("error building dashboard: %v", err)
	}

	if err := printDashboard(dashboard, outputFormat, noColor); err != nil {
		output.PrintError("Failed to print dashboard: %v", noColor, err)
		return fmt.Errorf("error printing dashboard: %v", err)
	}

	return nil
}
