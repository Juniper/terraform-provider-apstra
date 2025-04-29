// Copyright (c) Juniper Networks, Inc., 2022. All rights reserved.

package blueprint

import (
	"context"
	"encoding/json"
	"net"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSviIp_Request(t *testing.T) {
	tests := []struct {
		name     string
		sviIp    SviIp
		expected *apstra.SviIp
		wantErr  bool
	}{
		{
			name: "IPv4 only",
			sviIp: SviIp{
				SystemId:    types.StringValue("system-123"),
				IPv4Address: types.StringValue("192.0.2.2/24"),
				IPv4Mode:    types.StringValue("enabled"),
				IPv6Address: types.StringNull(),
				IPv6Mode:    types.StringValue("disabled"),
			},
			expected: &apstra.SviIp{
				SystemId: "system-123",
				Ipv4Mode: enum.SviIpv4ModeEnabled,
				Ipv6Mode: enum.SviIpv6ModeDisabled,
			},
			wantErr: false,
		},
		{
			name: "IPv6 only",
			sviIp: SviIp{
				SystemId:    types.StringValue("system-456"),
				IPv4Address: types.StringNull(),
				IPv4Mode:    types.StringValue("disabled"),
				IPv6Address: types.StringValue("2001:db8::2/64"),
				IPv6Mode:    types.StringValue("enabled"),
			},
			expected: &apstra.SviIp{
				SystemId: "system-456",
				Ipv4Mode: enum.SviIpv4ModeDisabled,
				Ipv6Mode: enum.SviIpv6ModeEnabled,
			},
			wantErr: false,
		},
		{
			name: "Both IPv4 and IPv6",
			sviIp: SviIp{
				SystemId:    types.StringValue("system-789"),
				IPv4Address: types.StringValue("192.0.2.3/24"),
				IPv4Mode:    types.StringValue("enabled"),
				IPv6Address: types.StringValue("2001:db8::3/64"),
				IPv6Mode:    types.StringValue("enabled"),
			},
			expected: &apstra.SviIp{
				SystemId: "system-789",
				Ipv4Mode: enum.SviIpv4ModeEnabled,
				Ipv6Mode: enum.SviIpv6ModeEnabled,
			},
			wantErr: false,
		},
		{
			name: "Invalid IPv4 address",
			sviIp: SviIp{
				SystemId:    types.StringValue("system-invalid"),
				IPv4Address: types.StringValue("invalid-ip"),
				IPv4Mode:    types.StringValue("enabled"),
				IPv6Address: types.StringNull(),
				IPv6Mode:    types.StringValue("disabled"),
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := tc.sviIp.Request(context.Background(), &diags)

			if tc.wantErr {
				assert.True(t, diags.HasError())
				return
			}

			require.False(t, diags.HasError())
			require.NotNil(t, result)
			
			assert.Equal(t, tc.expected.SystemId, result.SystemId)
			assert.Equal(t, tc.expected.Ipv4Mode, result.Ipv4Mode)
			assert.Equal(t, tc.expected.Ipv6Mode, result.Ipv6Mode)

			// For IPv4 test
			if !tc.sviIp.IPv4Address.IsNull() {
				_, expectedNet, _ := net.ParseCIDR(tc.sviIp.IPv4Address.ValueString())
				assert.NotNil(t, result.Ipv4Addr)
				assert.Equal(t, expectedNet.String(), result.Ipv4Addr.String())
			}

			// For IPv6 test
			if !tc.sviIp.IPv6Address.IsNull() {
				_, expectedNet, _ := net.ParseCIDR(tc.sviIp.IPv6Address.ValueString())
				assert.NotNil(t, result.Ipv6Addr)
				assert.Equal(t, expectedNet.String(), result.Ipv6Addr.String())
			}
		})
	}
}

func TestSviIp_LoadApiData(t *testing.T) {
	// Test setup - create a sample SviIp
	_, ipv4Net, _ := net.ParseCIDR("192.0.2.5/24")
	ip4 := net.ParseIP("192.0.2.5")
	ipv4Net.IP = ip4

	_, ipv6Net, _ := net.ParseCIDR("2001:db8::5/64")
	ip6 := net.ParseIP("2001:db8::5")
	ipv6Net.IP = ip6

	apiSviIp := apstra.SviIp{
		SystemId: "system-test",
		Ipv4Addr: ipv4Net,
		Ipv4Mode: enum.SviIpv4ModeEnabled,
		Ipv6Addr: ipv6Net,
		Ipv6Mode: enum.SviIpv6ModeEnabled,
	}

	var tfSviIp SviIp
	var diags diag.Diagnostics

	// Test loading from API data
	tfSviIp.LoadApiData(context.Background(), apiSviIp, &diags)

	require.False(t, diags.HasError())
	assert.Equal(t, "system-test", tfSviIp.SystemId.ValueString())
	assert.Equal(t, "192.0.2.5/24", tfSviIp.IPv4Address.ValueString())
	assert.Equal(t, "enabled", tfSviIp.IPv4Mode.ValueString())
	assert.Equal(t, "2001:db8::5/64", tfSviIp.IPv6Address.ValueString())
	assert.Equal(t, "enabled", tfSviIp.IPv6Mode.ValueString())
}

func TestLoadApiSviIps(t *testing.T) {
	// Test with empty slice
	result := LoadApiSviIps(context.Background(), []apstra.SviIp{}, &diag.Diagnostics{})
	assert.True(t, result.IsNull())

	// Test with populated slice
	_, ipv4Net1, _ := net.ParseCIDR("192.0.2.2/24")
	ipv4Net1.IP = net.ParseIP("192.0.2.2")
	
	_, ipv4Net2, _ := net.ParseCIDR("192.0.2.3/24")
	ipv4Net2.IP = net.ParseIP("192.0.2.3")

	apiSviIps := []apstra.SviIp{
		{
			SystemId: "system-1",
			Ipv4Addr: ipv4Net1,
			Ipv4Mode: enum.SviIpv4ModeEnabled,
			Ipv6Mode: enum.SviIpv6ModeDisabled,
		},
		{
			SystemId: "system-2",
			Ipv4Addr: ipv4Net2,
			Ipv4Mode: enum.SviIpv4ModeEnabled,
			Ipv6Mode: enum.SviIpv6ModeDisabled,
		},
	}

	var diags diag.Diagnostics
	result = LoadApiSviIps(context.Background(), apiSviIps, &diags)

	require.False(t, diags.HasError())
	assert.False(t, result.IsNull())
	
	// Verify set contents
	elements := result.Elements()
	assert.Equal(t, 2, len(elements))
}

func TestSviIpJsonMarshalUnmarshal(t *testing.T) {
	// This test ensures the SVI IP struct aligns with the SDK's JSON format
	apiSviIpJson := `{
		"ipv4_addr": "192.0.2.5/24",
		"ipv4_mode": "enabled",
		"ipv6_addr": "2001:db8::5/64",
		"ipv6_mode": "enabled",
		"system_id": "system-test"
	}`

	var apiSviIp apstra.SviIp
	err := json.Unmarshal([]byte(apiSviIpJson), &apiSviIp)
	require.NoError(t, err)

	assert.Equal(t, "system-test", string(apiSviIp.SystemId))
	assert.Equal(t, enum.SviIpv4ModeEnabled, apiSviIp.Ipv4Mode)
	assert.Equal(t, enum.SviIpv6ModeEnabled, apiSviIp.Ipv6Mode)
	assert.NotNil(t, apiSviIp.Ipv4Addr)
	assert.NotNil(t, apiSviIp.Ipv6Addr)
	assert.Equal(t, "192.0.2.5/24", apiSviIp.Ipv4Addr.String())
	assert.Equal(t, "2001:db8::5/64", apiSviIp.Ipv6Addr.String())

	// Test serialization
	serialized, err := json.Marshal(apiSviIp)
	require.NoError(t, err)

	var unmarshaled apstra.SviIp
	err = json.Unmarshal(serialized, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, apiSviIp.SystemId, unmarshaled.SystemId)
	assert.Equal(t, apiSviIp.Ipv4Mode, unmarshaled.Ipv4Mode)
	assert.Equal(t, apiSviIp.Ipv6Mode, unmarshaled.Ipv6Mode)
	assert.Equal(t, apiSviIp.Ipv4Addr.String(), unmarshaled.Ipv4Addr.String())
	assert.Equal(t, apiSviIp.Ipv6Addr.String(), unmarshaled.Ipv6Addr.String())
}