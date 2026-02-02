package mcr

import (
	"bytes"
	"context"
	"testing"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestListLookingGlassIPRoutes(t *testing.T) {
	// Store original functions and restore after test
	originalLoginFunc := config.LoginFunc
	originalFunc := listIPRoutesFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		listIPRoutesFunc = originalFunc
	}()

	// Mock the login function
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	}

	metric := 100
	localPref := 200
	age := 3600
	best := true

	// Mock the function
	listIPRoutesFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) ([]*megaport.LookingGlassIPRoute, error) {
		return []*megaport.LookingGlassIPRoute{
			{
				Prefix:    "10.0.0.0/24",
				NextHop:   "192.168.1.1",
				Protocol:  megaport.RouteProtocolBGP,
				Metric:    &metric,
				LocalPref: &localPref,
				ASPath:    []int{65001, 65002},
				Age:       &age,
				Interface: "eth0",
				VXCName:   "Test VXC",
				Best:      &best,
			},
		}, nil
	}

	// Create command
	cmd := &cobra.Command{}
	cmd.Flags().String("protocol", "", "")
	cmd.Flags().String("ip", "", "")

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Execute
	err := ListLookingGlassIPRoutes(cmd, []string{"test-mcr-uid"}, true, "json")
	assert.NoError(t, err)
}

func TestListLookingGlassIPRoutesWithFilter(t *testing.T) {
	// Store original functions and restore after test
	originalLoginFunc := config.LoginFunc
	originalFunc := listIPRoutesWithFilterFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		listIPRoutesWithFilterFunc = originalFunc
	}()

	// Mock the login function
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	}

	// Mock the function
	listIPRoutesWithFilterFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ListIPRoutesRequest) ([]*megaport.LookingGlassIPRoute, error) {
		assert.Equal(t, "test-mcr-uid", req.MCRID)
		assert.Equal(t, megaport.RouteProtocolBGP, req.Protocol)
		return []*megaport.LookingGlassIPRoute{
			{
				Prefix:   "10.0.0.0/24",
				NextHop:  "192.168.1.1",
				Protocol: megaport.RouteProtocolBGP,
			},
		}, nil
	}

	// Create command with filter flags
	cmd := &cobra.Command{}
	cmd.Flags().String("protocol", "BGP", "")
	cmd.Flags().String("ip", "", "")

	// Execute
	err := ListLookingGlassIPRoutes(cmd, []string{"test-mcr-uid"}, true, "json")
	assert.NoError(t, err)
}

func TestListLookingGlassBGPRoutes(t *testing.T) {
	// Store original functions and restore after test
	originalLoginFunc := config.LoginFunc
	originalFunc := listBGPRoutesFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		listBGPRoutesFunc = originalFunc
	}()

	// Mock the login function
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	}

	localPref := 100
	med := 50
	neighborASN := 65001
	age := 7200

	// Mock the function
	listBGPRoutesFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) ([]*megaport.LookingGlassBGPRoute, error) {
		return []*megaport.LookingGlassBGPRoute{
			{
				Prefix:      "10.0.0.0/24",
				NextHop:     "192.168.1.1",
				ASPath:      []int{65001, 65002, 65003},
				LocalPref:   &localPref,
				MED:         &med,
				Origin:      "IGP",
				Communities: []string{"65001:100", "65001:200"},
				Valid:       true,
				Best:        true,
				NeighborIP:  "192.168.1.2",
				NeighborASN: &neighborASN,
				Age:         &age,
				VXCName:     "Test VXC",
			},
		}, nil
	}

	// Create command
	cmd := &cobra.Command{}
	cmd.Flags().String("ip", "", "")

	// Execute
	err := ListLookingGlassBGPRoutes(cmd, []string{"test-mcr-uid"}, true, "json")
	assert.NoError(t, err)
}

func TestListLookingGlassBGPSessions(t *testing.T) {
	// Store original functions and restore after test
	originalLoginFunc := config.LoginFunc
	originalFunc := listBGPSessionsFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		listBGPSessionsFunc = originalFunc
	}()

	// Mock the login function
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	}

	uptime := 86400
	prefixesIn := 100
	prefixesOut := 50

	// Mock the function
	listBGPSessionsFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) ([]*megaport.LookingGlassBGPSession, error) {
		return []*megaport.LookingGlassBGPSession{
			{
				SessionID:       "session-123",
				NeighborAddress: "192.168.1.2",
				NeighborASN:     65001,
				LocalASN:        65000,
				Status:          megaport.BGPSessionStatusUp,
				Uptime:          &uptime,
				PrefixesIn:      &prefixesIn,
				PrefixesOut:     &prefixesOut,
				VXCId:           12345,
				VXCName:         "Test VXC",
				Description:     "Test BGP Session",
			},
		}, nil
	}

	// Create command
	cmd := &cobra.Command{}

	// Execute
	err := ListLookingGlassBGPSessions(cmd, []string{"test-mcr-uid"}, true, "json")
	assert.NoError(t, err)
}

func TestListLookingGlassBGPNeighborRoutes(t *testing.T) {
	// Store original functions and restore after test
	originalLoginFunc := config.LoginFunc
	originalFunc := listBGPNeighborRoutesFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		listBGPNeighborRoutesFunc = originalFunc
	}()

	// Mock the login function
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	}

	localPref := 100
	med := 50

	// Mock the function
	listBGPNeighborRoutesFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ListBGPNeighborRoutesRequest) ([]*megaport.LookingGlassBGPNeighborRoute, error) {
		assert.Equal(t, "test-mcr-uid", req.MCRID)
		assert.Equal(t, "session-123", req.SessionID)
		assert.Equal(t, megaport.LookingGlassRouteDirectionReceived, req.Direction)
		return []*megaport.LookingGlassBGPNeighborRoute{
			{
				Prefix:      "10.0.0.0/24",
				NextHop:     "192.168.1.1",
				ASPath:      []int{65001, 65002},
				LocalPref:   &localPref,
				MED:         &med,
				Origin:      "IGP",
				Communities: []string{"65001:100"},
				Valid:       true,
				Best:        true,
			},
		}, nil
	}

	// Create command
	cmd := &cobra.Command{}
	cmd.Flags().String("ip", "", "")

	// Execute
	err := ListLookingGlassBGPNeighborRoutes(cmd, []string{"test-mcr-uid", "session-123", "received"}, true, "json")
	assert.NoError(t, err)
}

func TestListLookingGlassBGPNeighborRoutesInvalidDirection(t *testing.T) {
	// Store original function and restore after test
	originalLoginFunc := config.LoginFunc
	defer func() { config.LoginFunc = originalLoginFunc }()

	// Mock the login function
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	}

	// Create command
	cmd := &cobra.Command{}
	cmd.Flags().String("ip", "", "")

	// Execute with invalid direction
	err := ListLookingGlassBGPNeighborRoutes(cmd, []string{"test-mcr-uid", "session-123", "invalid"}, true, "json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "direction must be 'advertised' or 'received'")
}

// Output conversion tests

func TestToIPRouteOutput(t *testing.T) {
	metric := 100
	localPref := 200
	age := 3661 // 1 hour, 1 minute, 1 second
	best := true

	route := &megaport.LookingGlassIPRoute{
		Prefix:    "10.0.0.0/24",
		NextHop:   "192.168.1.1",
		Protocol:  megaport.RouteProtocolBGP,
		Metric:    &metric,
		LocalPref: &localPref,
		ASPath:    []int{65001, 65002},
		Age:       &age,
		Interface: "eth0",
		VXCName:   "Test VXC",
		Best:      &best,
	}

	output, err := ToIPRouteOutput(route)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.0/24", output.Prefix)
	assert.Equal(t, "192.168.1.1", output.NextHop)
	assert.Equal(t, "BGP", output.Protocol)
	assert.Equal(t, "100", output.Metric)
	assert.Equal(t, "200", output.LocalPref)
	assert.Equal(t, "65001 65002", output.ASPath)
	assert.Equal(t, "1h1m", output.Age)
	assert.Equal(t, "eth0", output.Interface)
	assert.Equal(t, "Test VXC", output.VXCName)
	assert.Equal(t, "Yes", output.Best)
}

func TestToIPRouteOutputNil(t *testing.T) {
	_, err := ToIPRouteOutput(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid route: nil value")
}

func TestToBGPRouteOutput(t *testing.T) {
	localPref := 100
	med := 50
	neighborASN := 65001
	age := 90061 // 1 day, 1 hour, 1 minute, 1 second

	route := &megaport.LookingGlassBGPRoute{
		Prefix:      "10.0.0.0/24",
		NextHop:     "192.168.1.1",
		ASPath:      []int{65001, 65002, 65003},
		LocalPref:   &localPref,
		MED:         &med,
		Origin:      "IGP",
		Communities: []string{"65001:100", "65001:200"},
		Valid:       true,
		Best:        false,
		NeighborIP:  "192.168.1.2",
		NeighborASN: &neighborASN,
		Age:         &age,
		VXCName:     "Test VXC",
	}

	output, err := ToBGPRouteOutput(route)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.0/24", output.Prefix)
	assert.Equal(t, "192.168.1.1", output.NextHop)
	assert.Equal(t, "65001 65002 65003", output.ASPath)
	assert.Equal(t, "100", output.LocalPref)
	assert.Equal(t, "50", output.MED)
	assert.Equal(t, "IGP", output.Origin)
	assert.Equal(t, "65001:100, 65001:200", output.Communities)
	assert.Equal(t, "Yes", output.Valid)
	assert.Equal(t, "No", output.Best)
	assert.Equal(t, "192.168.1.2", output.NeighborIP)
	assert.Equal(t, "65001", output.NeighborASN)
	assert.Equal(t, "1d1h", output.Age)
	assert.Equal(t, "Test VXC", output.VXCName)
}

func TestToBGPSessionOutput(t *testing.T) {
	uptime := 86400
	prefixesIn := 100
	prefixesOut := 50

	session := &megaport.LookingGlassBGPSession{
		SessionID:       "session-123",
		NeighborAddress: "192.168.1.2",
		NeighborASN:     65001,
		LocalASN:        65000,
		Status:          megaport.BGPSessionStatusUp,
		Uptime:          &uptime,
		PrefixesIn:      &prefixesIn,
		PrefixesOut:     &prefixesOut,
		VXCId:           12345,
		VXCName:         "Test VXC",
		Description:     "Test BGP Session",
	}

	output, err := ToBGPSessionOutput(session)
	assert.NoError(t, err)
	assert.Equal(t, "session-123", output.SessionID)
	assert.Equal(t, "192.168.1.2", output.NeighborAddress)
	assert.Equal(t, 65001, output.NeighborASN)
	assert.Equal(t, 65000, output.LocalASN)
	assert.Equal(t, "UP", output.Status)
	assert.Equal(t, "1d0h", output.Uptime)
	assert.Equal(t, "100", output.PrefixesIn)
	assert.Equal(t, "50", output.PrefixesOut)
	assert.Equal(t, "Test VXC", output.VXCName)
	assert.Equal(t, "Test BGP Session", output.Description)
}

func TestToBGPNeighborRouteOutput(t *testing.T) {
	localPref := 100
	med := 50

	route := &megaport.LookingGlassBGPNeighborRoute{
		Prefix:      "10.0.0.0/24",
		NextHop:     "192.168.1.1",
		ASPath:      []int{65001, 65002},
		LocalPref:   &localPref,
		MED:         &med,
		Origin:      "IGP",
		Communities: []string{"65001:100"},
		Valid:       true,
		Best:        true,
	}

	output, err := ToBGPNeighborRouteOutput(route)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.0/24", output.Prefix)
	assert.Equal(t, "192.168.1.1", output.NextHop)
	assert.Equal(t, "65001 65002", output.ASPath)
	assert.Equal(t, "100", output.LocalPref)
	assert.Equal(t, "50", output.MED)
	assert.Equal(t, "IGP", output.Origin)
	assert.Equal(t, "65001:100", output.Communities)
	assert.Equal(t, "Yes", output.Valid)
	assert.Equal(t, "Yes", output.Best)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{30, "30s"},
		{61, "1m1s"},
		{3661, "1h1m"},
		{90061, "1d1h"},
		{180000, "2d2h"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.seconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBoolToYesNo(t *testing.T) {
	assert.Equal(t, "Yes", boolToYesNo(true))
	assert.Equal(t, "No", boolToYesNo(false))
}
