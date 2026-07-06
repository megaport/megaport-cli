package nat_gateway

import (
	"errors"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The conflict guard runs before login, so these cases never reach the API.

func TestCreateNATGateway_InteractiveConflict(t *testing.T) {
	tests := []struct {
		name    string
		setFlag func(*cobra.Command)
	}{
		{
			name: "interactive with value flag",
			setFlag: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("name", "My NAT GW"))
			},
		},
		{
			name: "interactive with json",
			setFlag: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("json", `{"name":"GW","term":12,"speed":1000,"locationId":1}`))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockNATGatewayService{}
			defer setupMockNATGateway(mock)()

			cmd := newTestCmd("create")
			require.NoError(t, cmd.Flags().Set("interactive", "true"))
			tt.setFlag(cmd)

			err := CreateNATGateway(cmd, nil, true)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be combined with")
			assert.Nil(t, mock.CapturedCreateReq)
		})
	}
}

func TestCreateNATGateway_InteractiveConflict_UsageCode(t *testing.T) {
	mock := &MockNATGatewayService{}
	defer setupMockNATGateway(mock)()

	cmd := newTestCmd("create")
	require.NoError(t, cmd.Flags().Set("interactive", "true"))
	require.NoError(t, cmd.Flags().Set("name", "My NAT GW"))

	err := CreateNATGateway(cmd, nil, true)
	require.Error(t, err)

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.Usage, cliErr.Code)
}

func TestUpdateNATGateway_InteractiveConflict(t *testing.T) {
	tests := []struct {
		name    string
		setFlag func(*cobra.Command)
	}{
		{
			name: "interactive with value flag",
			setFlag: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("name", "New Name"))
			},
		},
		{
			name: "interactive with json",
			setFlag: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("json", `{"name":"Updated"}`))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockNATGatewayService{}
			defer setupMockNATGateway(mock)()

			cmd := newTestCmd("update")
			require.NoError(t, cmd.Flags().Set("interactive", "true"))
			tt.setFlag(cmd)

			err := UpdateNATGateway(cmd, []string{"uid-conflict"}, true)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be combined with")
			assert.Nil(t, mock.CapturedUpdateReq)
		})
	}
}

func TestUpdateNATGateway_InteractiveConflict_UsageCode(t *testing.T) {
	mock := &MockNATGatewayService{}
	defer setupMockNATGateway(mock)()

	cmd := newTestCmd("update")
	require.NoError(t, cmd.Flags().Set("interactive", "true"))
	require.NoError(t, cmd.Flags().Set("name", "New Name"))

	err := UpdateNATGateway(cmd, []string{"uid-conflict"}, true)
	require.Error(t, err)

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.Usage, cliErr.Code)
}
