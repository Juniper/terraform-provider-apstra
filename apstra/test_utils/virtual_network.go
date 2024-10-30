package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func VirtualNetworkVxlan(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient, cleanup bool) apstra.ObjectId {
	leafIds := leafSwitches(t, ctx, client)
	vnBindings := make([]apstra.VnBinding, len(leafIds))
	for i, leafId := range leafIds {
		vnBindings[i] = apstra.VnBinding{SystemId: leafId}
	}

	id, err := client.CreateVirtualNetwork(ctx, &apstra.VirtualNetworkData{
		Ipv4Enabled:    true,
		Label:          acctest.RandString(6),
		SecurityZoneId: SecurityZoneA(t, ctx, client, cleanup),
		VnBindings:     vnBindings,
		VnType:         enum.VnTypeVxlan,
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
