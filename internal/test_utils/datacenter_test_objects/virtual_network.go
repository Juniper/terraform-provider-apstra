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

// VirtualNetworkA creates a minimally configured EVPN routing (security) zone
func VirtualNetworkA(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient, rzID string) string {
	t.Helper()

	id, err := client.CreateVirtualNetwork(ctx, datacenter.VirtualNetwork{
		IPv4Enabled:    true,
		IPv6Enabled:    true,
		Label:          acctest.RandString(10),
		SecurityZoneID: rzID,
		Type:           enum.VnTypeVxlan,
	})
	require.NoError(t, err)

	return id
}
