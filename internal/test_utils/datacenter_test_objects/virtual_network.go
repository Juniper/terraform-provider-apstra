package dctestobj

import (
	"context"
	"net"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

// VirtualNetworkA creates a minimally configured Virtual Network in the given routing zone and returns its ID.
func VirtualNetworkA(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient, rzID string) string {
	t.Helper()

	id, err := bp.CreateVirtualNetwork(ctx, datacenter.VirtualNetwork{
		IPv4Enabled:    true,
		IPv6Enabled:    true,
		Label:          acctest.RandString(10),
		SecurityZoneID: rzID,
		Type:           enum.VnTypeVxlan,
	})
	require.NoError(t, err)

	return id
}

// VirtualNetworkB creates a minimally configured Virtual Network in the given routing
// zone and returns its ID. The Virtual Network will be bound to all leaf switches.
func VirtualNetworkB(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient, rzID string) string {
	t.Helper()

	query := new(apstra.PathQuery).SetBlueprintId(bp.Id()).SetClient(bp.Client()).Node([]apstra.QEEAttribute{
		apstra.NodeTypeSystem.QEEAttribute(),
		{Key: "system_type", Value: apstra.QEStringVal("switch")},
		{Key: "role", Value: apstra.QEStringVal("leaf")},
		{Key: "name", Value: apstra.QEStringVal("n_system")},
	})
	var result struct {
		Items []struct {
			System struct {
				ID string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}
	require.NoError(t, query.Do(ctx, &result))
	bindings := make([]datacenter.VNBinding, len(result.Items))
	for i, item := range result.Items {
		bindings[i] = datacenter.VNBinding{SystemID: item.System.ID}
	}

	id, err := bp.CreateVirtualNetwork(ctx, datacenter.VirtualNetwork{
		IPv4Enabled:    true,
		IPv6Enabled:    true,
		Label:          acctest.RandString(10),
		SecurityZoneID: rzID,
		Type:           enum.VnTypeVxlan,
		Bindings:       bindings,
	})
	require.NoError(t, err)

	return id
}

// VirtualNetworkC creates a Virtual Network enabled for both IPv4 and IPv6 using the provided rzID,
// ipv4cidr and ipv6cidr details. The virtual gateway for each address family will be enabled and set to
// the first usable address in the provided subnet. The function returns the ID of the new Virtual Network.
func VirtualNetworkC(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient, rzID, ipv4cidr, ipv6cidr string) string {
	t.Helper()

	nextIP := func(ip net.IP) net.IP {
		b := append(net.IP(nil), ip...) // copy so we don't mutate input
		if len(b) == 0 {
			return b // no such thing as next
		}

		for i := len(b) - 1; i >= 0; i-- {
			old := b[i]
			b[i]++
			if b[i] >= old {
				return b // no overflow
			}
			// overflow: keep propagating carry to the left
		}

		return b // overflowed all the way through; this means 255...255 + 1
	}

	ipv4, ipv4Subnet, err := net.ParseCIDR(ipv4cidr)
	require.NoError(t, err)
	require.NotNil(t, ipv4Subnet)
	require.True(t, ipv4Subnet.IP.Equal(ipv4)) // require that the provided IP is the base address
	ipv4Gateway := nextIP(ipv4)
	require.True(t, ipv4Subnet.Contains(ipv4Gateway))

	ipv6, ipv6Subnet, err := net.ParseCIDR(ipv6cidr)
	require.NoError(t, err)
	require.NotNil(t, ipv6Subnet)
	require.True(t, ipv6Subnet.IP.Equal(ipv6)) // require that the provided IP is the base address
	ipv6Gateway := nextIP(ipv6)
	require.True(t, ipv6Subnet.Contains(ipv6Gateway))

	id, err := bp.CreateVirtualNetwork(ctx, datacenter.VirtualNetwork{
		IPv4Enabled:               true,
		IPv4Subnet:                ipv4Subnet,
		VirtualGatewayIPv4Enabled: true,
		VirtualGatewayIPv4:        ipv4Gateway,
		IPv6Enabled:               true,
		IPv6Subnet:                ipv6Subnet,
		VirtualGatewayIPv6Enabled: true,
		VirtualGatewayIPv6:        ipv6Gateway,
		Label:                     acctest.RandString(10),
		SecurityZoneID:            rzID,
		Type:                      enum.VnTypeVxlan,
	})
	require.NoError(t, err)

	return id
}
