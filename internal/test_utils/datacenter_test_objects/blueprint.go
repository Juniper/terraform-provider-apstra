//go:build integration

package dctestobj

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/compatibility"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

// BlueprintA creates a blueprint from the L2_Virtual_EVPN template with a dual-stack underlay.
// The devices all have associated interface maps assigned, but no resource pools are assigned.
func BlueprintA(t testing.TB, ctx context.Context, client *apstra.Client) *apstra.TwoStageL3ClosClient {
	t.Helper()

	request := apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  enum.RefDesignDatacenter,
		Label:      acctest.RandString(6),
		TemplateId: "L2_Virtual_EVPN",
		FabricSettings: &apstra.FabricSettings{
			SpineSuperspineLinks: pointer.To(apstra.AddressingSchemeIp46),
			SpineLeafLinks:       pointer.To(apstra.AddressingSchemeIp46),
		},
	}

	// Dual-stack config for Apstra 6.1.2+
	if compatibility.DatacenterPolicyAddressFamilyRequired.Check(version.Must(version.NewVersion(client.ApiVersion()))) {
		request.AddressingPolicy = &apstra.AddressingPolicy{
			AddressingSupport: pointer.To(enum.AddressingSchemeIPv46),
			DisableIPv4:       pointer.To(false),
			VTEPAddressing:    pointer.To(enum.AddressingSchemeIPv4),
		}
	}

	// Create the blueprint and register
	id, err := client.CreateBlueprintFromTemplate(ctx, &request)
	require.NoError(t, err)
	testutils.CleanupWithFreshContext(
		t, testutils.DefaultTimeout,
		func(ctx context.Context) error {
			return client.DeleteBlueprint(ctx, id)
		},
	)

	bp, err := client.NewTwoStageL3ClosClient(ctx, id)
	require.NoError(t, err)

	// Assign interface maps to all devices in the blueprint.
	query := new(apstra.PathQuery).SetBlueprintId(id).SetClient(client).Node([]apstra.QEEAttribute{
		apstra.NodeTypeSystem.QEEAttribute(),
		{Key: "system_type", Value: apstra.QEStringVal("switch")},
		{Key: "name", Value: apstra.QEStringVal("n_system")},
	})
	var result struct {
		Items []struct {
			System struct {
				ID   string              `json:"id"`
				Role enum.SystemNodeRole `json:"role"`
			} `json:"n_system"`
		} `json:"items"`
	}
	require.NoError(t, query.Do(ctx, &result))
	assignments := make(apstra.SystemIdToInterfaceMapAssignment, len(result.Items))
	for _, item := range result.Items {
		switch item.System.Role {
		case enum.SystemNodeRoleLeaf:
			assignments[item.System.ID] = "Juniper_vQFX__AOS-7x10-Leaf"
		case enum.SystemNodeRoleSpine:
			assignments[item.System.ID] = "Juniper_vQFX__AOS-7x10-Spine"
		}
	}
	require.NoError(t, bp.SetInterfaceMapAssignments(ctx, assignments))

	return bp
}
