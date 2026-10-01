package dctestobj

import (
	"context"
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
