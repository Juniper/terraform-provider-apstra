package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func VirtualNetworkVxlan(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient, cleanup bool) string {
	leafIds := leafSwitches(t, ctx, client)
	vnBindings := make([]datacenter.VNBinding, len(leafIds))
	for i, leafId := range leafIds {
		vnBindings[i] = datacenter.VNBinding{SystemID: leafId}
	}

	id, err := client.CreateVirtualNetwork(ctx, datacenter.VirtualNetwork{
		IPv4Enabled:    true,
		Label:          acctest.RandString(6),
		SecurityZoneID: SecurityZoneA(t, ctx, client, cleanup),
		Bindings:       vnBindings,
		Type:           enum.VnTypeVxlan,
	})
	require.NoError(t, err)
	if cleanup {
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			require.NoError(t, client.DeleteVirtualNetwork(ctx, id))
		})
	}

	return id
}
