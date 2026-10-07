//go:build integration

package resourcetestobj

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func RandomIPv6Pool(t testing.TB, ctx context.Context, client *apstra.Client, cidr string, bits, n int) string {
	t.Helper()

	subnetCount := 3

	subnets := random.Prefixes(t, cidr, bits, n)

	request := apstra.NewIpPoolRequest{
		DisplayName: acctest.RandStringFromCharSet(6, acctest.CharSetAlpha),
		Subnets:     make([]apstra.NewIpSubnet, subnetCount),
	}

	for i := range subnets {
		request.Subnets[i] = apstra.NewIpSubnet{Network: subnets[i].String()}
	}

	id, err := client.CreateIp6Pool(ctx, &request)
	require.NoError(t, err)

	testutils.CleanupWithFreshContext(t, testutils.DefaultCleanupTimeout, func(ctx context.Context) error {
		return client.DeleteIp6Pool(ctx, id)
	})

	return string(id)
}
