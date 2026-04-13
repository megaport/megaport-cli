package validation

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestValidateCreateNATGatewayRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *megaport.CreateNATGatewayRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "My NAT GW",
				LocationID:  123,
				Speed:       1000,
				Term:        12,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: &megaport.CreateNATGatewayRequest{
				LocationID: 123,
				Speed:      1000,
				Term:       12,
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "missing location",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				Speed:       1000,
				Term:        12,
			},
			wantErr: true,
			errMsg:  "location ID",
		},
		{
			name: "invalid location (negative)",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  -1,
				Speed:       1000,
				Term:        12,
			},
			wantErr: true,
			errMsg:  "location ID",
		},
		{
			name: "missing speed",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  123,
				Term:        12,
			},
			wantErr: true,
			errMsg:  "speed",
		},
		{
			name: "invalid speed (negative)",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  123,
				Speed:       -100,
				Term:        12,
			},
			wantErr: true,
			errMsg:  "speed",
		},
		{
			name: "invalid term",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  123,
				Speed:       1000,
				Term:        5,
			},
			wantErr: true,
			errMsg:  "contract term",
		},
		{
			name: "valid term 1",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  1,
				Speed:       100,
				Term:        1,
			},
			wantErr: false,
		},
		{
			name: "valid term 24",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  1,
				Speed:       100,
				Term:        24,
			},
			wantErr: false,
		},
		{
			name: "valid term 36",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  1,
				Speed:       100,
				Term:        36,
			},
			wantErr: false,
		},
		{
			name: "negative session count",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  1,
				Speed:       100,
				Term:        12,
				Config:      megaport.NATGatewayNetworkConfig{SessionCount: -1},
			},
			wantErr: true,
			errMsg:  "session count",
		},
		{
			name: "negative ASN",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  1,
				Speed:       100,
				Term:        12,
				Config:      megaport.NATGatewayNetworkConfig{ASN: -1},
			},
			wantErr: true,
			errMsg:  "ASN",
		},
		{
			name: "zero session count is valid (unset/default)",
			req: &megaport.CreateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  1,
				Speed:       100,
				Term:        12,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateNATGatewayRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUpdateNATGatewayRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *megaport.UpdateNATGatewayRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &megaport.UpdateNATGatewayRequest{
				ProductUID:  "uid-123",
				ProductName: "Updated GW",
				LocationID:  100,
				Speed:       2000,
				Term:        12,
			},
			wantErr: false,
		},
		{
			name: "empty product UID",
			req: &megaport.UpdateNATGatewayRequest{
				ProductName: "GW",
				LocationID:  100,
				Speed:       1000,
				Term:        12,
			},
			wantErr: true,
			errMsg:  "product UID",
		},
		{
			name: "empty name",
			req: &megaport.UpdateNATGatewayRequest{
				ProductUID: "uid-123",
				LocationID: 100,
				Speed:      1000,
				Term:       12,
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "missing location",
			req: &megaport.UpdateNATGatewayRequest{
				ProductUID:  "uid-123",
				ProductName: "GW",
				Speed:       1000,
				Term:        12,
			},
			wantErr: true,
			errMsg:  "location ID",
		},
		{
			name: "missing speed",
			req: &megaport.UpdateNATGatewayRequest{
				ProductUID:  "uid-123",
				ProductName: "GW",
				LocationID:  100,
				Term:        12,
			},
			wantErr: true,
			errMsg:  "speed",
		},
		{
			name: "invalid term",
			req: &megaport.UpdateNATGatewayRequest{
				ProductUID:  "uid-123",
				ProductName: "GW",
				LocationID:  100,
				Speed:       1000,
				Term:        7,
			},
			wantErr: true,
			errMsg:  "contract term",
		},
		{
			name: "negative session count",
			req: &megaport.UpdateNATGatewayRequest{
				ProductUID:  "uid-123",
				ProductName: "GW",
				LocationID:  100,
				Speed:       1000,
				Term:        12,
				Config:      megaport.NATGatewayNetworkConfig{SessionCount: -1},
			},
			wantErr: true,
			errMsg:  "session count",
		},
		{
			name: "negative ASN",
			req: &megaport.UpdateNATGatewayRequest{
				ProductUID:  "uid-123",
				ProductName: "GW",
				LocationID:  100,
				Speed:       1000,
				Term:        12,
				Config:      megaport.NATGatewayNetworkConfig{ASN: -1},
			},
			wantErr: true,
			errMsg:  "ASN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateNATGatewayRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
