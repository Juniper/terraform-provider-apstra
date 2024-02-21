package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"testing"
)

func VniPool(t testing.TB, ctx context.Context, first, last uint32) *apstra.VniPool {
	t.Helper()

	client := GetTestClient(t, ctx)

	id, err := client.CreateVniPool(ctx, &apstra.VniPoolRequest{
		DisplayName: acctest.RandString(5),
		Ranges: []apstra.IntfIntRange{
			apstra.IntRange{
				First: first,
				Last:  last,
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteVniPool(ctx, id)) })

	pool, err := client.GetVniPool(ctx, id)
	require.NoError(t, err)

	return pool
}
