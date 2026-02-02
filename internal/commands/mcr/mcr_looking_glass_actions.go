package mcr

import (
	"context"
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// ListLookingGlassIPRoutes lists IP routes from the MCR Looking Glass
func ListLookingGlassIPRoutes(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	// Get optional filters
	protocol, _ := cmd.Flags().GetString("protocol")
	ipFilter, _ := cmd.Flags().GetString("ip")

	spinner := output.PrintResourceListing("IP routes", noColor)

	var routes []*megaport.LookingGlassIPRoute

	if protocol != "" || ipFilter != "" {
		req := &megaport.ListIPRoutesRequest{
			MCRID:    mcrUID,
			IPFilter: ipFilter,
		}
		if protocol != "" {
			req.Protocol = megaport.RouteProtocol(protocol)
		}
		routes, err = listIPRoutesWithFilterFunc(ctx, client, req)
	} else {
		routes, err = listIPRoutesFunc(ctx, client, mcrUID)
	}

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list IP routes: %v", noColor, err)
		return fmt.Errorf("error listing IP routes: %v", err)
	}

	if len(routes) == 0 {
		output.PrintWarning("No IP routes found", noColor)
	}

	return printIPRoutes(routes, outputFormat, noColor)
}

// ListLookingGlassBGPRoutes lists BGP routes from the MCR Looking Glass
func ListLookingGlassBGPRoutes(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	// Get optional filter
	ipFilter, _ := cmd.Flags().GetString("ip")

	spinner := output.PrintResourceListing("BGP routes", noColor)

	var routes []*megaport.LookingGlassBGPRoute

	if ipFilter != "" {
		req := &megaport.ListBGPRoutesRequest{
			MCRID:    mcrUID,
			IPFilter: ipFilter,
		}
		routes, err = listBGPRoutesWithFilterFunc(ctx, client, req)
	} else {
		routes, err = listBGPRoutesFunc(ctx, client, mcrUID)
	}

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list BGP routes: %v", noColor, err)
		return fmt.Errorf("error listing BGP routes: %v", err)
	}

	if len(routes) == 0 {
		output.PrintWarning("No BGP routes found", noColor)
	}

	return printBGPRoutes(routes, outputFormat, noColor)
}

// ListLookingGlassBGPSessions lists BGP sessions from the MCR Looking Glass
func ListLookingGlassBGPSessions(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	spinner := output.PrintResourceListing("BGP sessions", noColor)

	sessions, err := listBGPSessionsFunc(ctx, client, mcrUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list BGP sessions: %v", noColor, err)
		return fmt.Errorf("error listing BGP sessions: %v", err)
	}

	if len(sessions) == 0 {
		output.PrintWarning("No BGP sessions found", noColor)
	}

	return printBGPSessions(sessions, outputFormat, noColor)
}

// ListLookingGlassBGPNeighborRoutes lists routes advertised to or received from a specific BGP neighbor
func ListLookingGlassBGPNeighborRoutes(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]
	sessionID := args[1]
	direction := args[2]

	// Validate direction
	if direction != "advertised" && direction != "received" {
		return fmt.Errorf("direction must be 'advertised' or 'received', got: %s", direction)
	}

	// Get optional filter
	ipFilter, _ := cmd.Flags().GetString("ip")

	spinner := output.PrintResourceListing("BGP neighbor routes", noColor)

	req := &megaport.ListBGPNeighborRoutesRequest{
		MCRID:     mcrUID,
		SessionID: sessionID,
		Direction: megaport.LookingGlassRouteDirection(direction),
		IPFilter:  ipFilter,
	}

	routes, err := listBGPNeighborRoutesFunc(ctx, client, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list BGP neighbor routes: %v", noColor, err)
		return fmt.Errorf("error listing BGP neighbor routes: %v", err)
	}

	if len(routes) == 0 {
		output.PrintWarning("No BGP neighbor routes found", noColor)
	}

	return printBGPNeighborRoutes(routes, outputFormat, noColor)
}
