//go:build integration

package resourcetestobj

import (
	"context"
	"slices"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func RandomASNPool(t testing.TB, ctx context.Context, client *apstra.Client, min, max int) string {
	t.Helper()

	rangeCount := 3

	ints, err := random.Ints(min, max, 2*rangeCount)
	require.NoError(t, err)

	slices.Sort(ints)

	request := apstra.AsnPoolRequest{
		DisplayName: acctest.RandStringFromCharSet(6, acctest.CharSetAlpha),
		Ranges:      make([]apstra.IntfIntRange, rangeCount),
	}

	for i := range rangeCount {
		request.Ranges[i] = apstra.IntRange{
			First: uint32(ints[i*2]),
			Last:  uint32(ints[i*2+1]),
		}
	}

	id, err := client.CreateAsnPool(ctx, &request)
	require.NoError(t, err)

	testutils.CleanupWithFreshContext(t, testutils.DefaultCleanupTimeout, func(ctx context.Context) error {
		return client.DeleteAsnPool(ctx, id)
	})

	return string(id)
}
