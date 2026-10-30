//go:build integration

package designtestobjects

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/design"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/apstra-go-sdk/policy"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

// TemplateA creates a template using a single instance of the RackType() test helper.
func TemplateA(ctx context.Context, t testing.TB, client *apstra.Client) string {
	t.Helper()
	rtID := RackTypeA(ctx, t, client)
	rt, err := client.GetRackType2(ctx, rtID)
	require.NoError(t, err)

	tID, err := client.CreateTemplateRackBased2(ctx, design.TemplateRackBased{
		Label: acctest.RandString(6),
		Racks: []design.RackTypeWithCount{
			{
				Count:    1,
				RackType: rt,
			},
		},
		AntiAffinityPolicy:  nil,
		ASNAllocationPolicy: &policy.ASNAllocation{SpineASNScheme: enum.ASNAllocationSchemeDistinct},
		Capability:          &enum.TemplateCapabilityBlueprint,
		DHCPServiceIntent:   policy.DHCPServiceIntent{Active: true},
		Spine: design.Spine{
			Count:                  1,
			LinkPerSuperspineCount: 0,
			LinkPerSuperspineSpeed: "",
			LogicalDevice:          rt.LeafSwitches[0].LogicalDevice,
		},
		VirtualNetworkPolicy: &policy.VirtualNetwork{OverlayControlProtocol: enum.OverlayControlProtocolEVPN},
	})
	require.NoError(t, err)

	testutils.CleanupWithFreshContext(t, testutils.DefaultCleanupTimeout, func(ctx context.Context) error {
		return client.DeleteTemplate2(ctx, tID)
	})

	return tID
}
