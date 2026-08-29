//go:build integration

package dctestobj

import (
	"context"
	"net"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func RoutingPolicyRandom(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient) string {
	t.Helper()

	id, err := bp.CreateRoutingPolicy(ctx, &apstra.DcRoutingPolicyData{
		Label:        acctest.RandString(6),
		Description:  acctest.RandString(12),
		PolicyType:   apstra.DcRoutingPolicyTypeUser,
		ImportPolicy: random.OneOf(apstra.DcRoutingPolicyImportPolicyAll, apstra.DcRoutingPolicyImportPolicyDefaultOnly, apstra.DcRoutingPolicyImportPolicyExtraOnly),
		ExportPolicy: apstra.DcRoutingExportPolicy{
			StaticRoutes:         random.OneOf(true, false),
			Loopbacks:            random.OneOf(true, false),
			SpineSuperspineLinks: random.OneOf(true, false),
			L3EdgeServerLinks:    random.OneOf(true, false),
			SpineLeafLinks:       random.OneOf(true, false),
			L2EdgeSubnets:        random.OneOf(true, false),
		},
		ExpectDefaultIpv4Route: random.OneOf(true, false),
		ExpectDefaultIpv6Route: random.OneOf(true, false),
		AggregatePrefixes:      []net.IPNet{random.NetIPNet(t, "10.0.0.0/16")},
		ExtraImportRoutes: []apstra.PrefixFilter{{
			Action: random.OneOf(apstra.PrefixFilterActionDeny, apstra.PrefixFilterActionPermit),
			Prefix: random.NetIPNet(t, "10.1.0.0/16"),
			GeMask: nil,
			LeMask: nil,
		}},
		ExtraExportRoutes: []apstra.PrefixFilter{{
			Action: random.OneOf(apstra.PrefixFilterActionDeny, apstra.PrefixFilterActionPermit),
			Prefix: random.NetIPNet(t, "10.1.0.0/16"),
			GeMask: nil,
			LeMask: nil,
		}},
	})
	require.NoError(t, err)

	return id.String()
}
