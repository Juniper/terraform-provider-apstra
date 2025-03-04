package testutils

import (
	"context"
	"sync"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/hashicorp/go-version"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

// MakeOrFindBlueprintMutex is created by TestMain()
var MakeOrFindBlueprintMutex *sync.Mutex

type bFunc func(t testing.TB, ctx context.Context, name ...string) *apstra.TwoStageL3ClosClient

func MakeOrFindBlueprint(t testing.TB, ctx context.Context, name string, f bFunc) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)

	MakeOrFindBlueprintMutex.Lock()

	status, err := client.GetBlueprintStatusByName(ctx, name)
	if err != nil {
		defer MakeOrFindBlueprintMutex.Unlock()
		if utils.IsApstra404(err) {
			return f(t, ctx, name)
		}

		t.Fatal(err)
	}

	MakeOrFindBlueprintMutex.Unlock()

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, status.Id)
	require.NoError(t, err)

	return bpClient
}

func BlueprintA(t testing.TB, ctx context.Context, name ...string) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)

	var bpname string
	if name == nil {
		bpname = acctest.RandString(10)
	} else {
		bpname = name[0]
	}

	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  enum.RefDesignDatacenter,
		Label:      bpname,
		TemplateId: "L2_Virtual_EVPN",
		FabricSettings: &apstra.FabricSettings{
			SpineSuperspineLinks: utils.ToPtr(apstra.AddressingSchemeIp4),
			SpineLeafLinks:       utils.ToPtr(apstra.AddressingSchemeIp4),
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	require.NoError(t, err)

	return bpClient
}

func BlueprintB(t testing.TB, ctx context.Context) (*apstra.TwoStageL3ClosClient, apstra.ObjectId) {
	t.Helper()

	client := GetTestClient(t, ctx)
	template := TemplateA(t, ctx)
	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  enum.RefDesignDatacenter,
		Label:      name,
		TemplateId: template.Id,
		FabricSettings: &apstra.FabricSettings{
			SpineSuperspineLinks: utils.ToPtr(apstra.AddressingSchemeIp4),
			SpineLeafLinks:       utils.ToPtr(apstra.AddressingSchemeIp4),
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	require.NoError(t, err)

	return bpClient, template.Id
}

func BlueprintC(t testing.TB, ctx context.Context) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)
	template := TemplateB(t, ctx)
	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  enum.RefDesignDatacenter,
		Label:      name,
		TemplateId: template.Id,
		FabricSettings: &apstra.FabricSettings{
			SpineSuperspineLinks: utils.ToPtr(apstra.AddressingSchemeIp4),
			SpineLeafLinks:       utils.ToPtr(apstra.AddressingSchemeIp4),
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	require.NoError(t, err)

	return bpClient
}

func BlueprintD(t testing.TB, ctx context.Context) *apstra.TwoStageL3ClosClient {
	client := GetTestClient(t, ctx)
	template := TemplateC(t, ctx)
	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  enum.RefDesignDatacenter,
		Label:      name,
		TemplateId: template.Id,
		FabricSettings: &apstra.FabricSettings{
			SpineSuperspineLinks: utils.ToPtr(apstra.AddressingSchemeIp4),
			SpineLeafLinks:       utils.ToPtr(apstra.AddressingSchemeIp4),
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	require.NoError(t, err)

	leafQuery := new(apstra.PathQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(bpClient.Id()).
		SetClient(bpClient.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal("leaf")},
			{Key: "name", Value: apstra.QEStringVal("n_leaf")},
		})
	var leafQueryResult struct {
		Items []struct {
			Leaf struct {
				Id string `json:"id"`
			} `json:"n_leaf"`
		} `json:"items"`
	}
	require.NoError(t, leafQuery.Do(ctx, &leafQueryResult))

	accessQuery := new(apstra.PathQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(bpClient.Id()).
		SetClient(bpClient.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal("access")},
			{Key: "name", Value: apstra.QEStringVal("n_access")},
		})
	var accessQueryResult struct {
		Items []struct {
			Access struct {
				Id string `json:"id"`
			} `json:"n_access"`
		} `json:"items"`
	}
	require.NoError(t, accessQuery.Do(ctx, &accessQueryResult))

	leafIds := make([]string, len(leafQueryResult.Items))
	accessIds := make([]string, len(accessQueryResult.Items))
	assignments := make(apstra.SystemIdToInterfaceMapAssignment, len(leafIds)+len(accessIds))

	for i, item := range leafQueryResult.Items {
		leafIds[i] = item.Leaf.Id
		assignments[item.Leaf.Id] = "Juniper_vQFX__AOS-7x10-Leaf"
	}
	for i, item := range accessQueryResult.Items {
		accessIds[i] = item.Access.Id
		assignments[item.Access.Id] = "Juniper_vQFX__AOS-8x10-1"
	}

	require.NoError(t, bpClient.SetInterfaceMapAssignments(ctx, assignments))

	return bpClient
}

//func BlueprintE(t testing.TB, ctx context.Context) *apstra.TwoStageL3ClosClient {
//}

// BlueprintF creates a blueprint with switches where a single port is represented
// by two different names:
// - xe-0/0/0 - transform 1
// - ge-0/0/0 - transform 2 and 3
func BlueprintF(t testing.TB, ctx context.Context) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)

	rackId, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              acctest.RandString(6),
		FabricConnectivityDesign: apstra.FabricConnectivityDesignL3Clos,
		LeafSwitches: []apstra.RackElementLeafSwitchRequest{
			{
				Label:              acctest.RandString(6),
				LinkPerSpineCount:  1,
				LinkPerSpineSpeed:  "40G",
				RedundancyProtocol: apstra.LeafRedundancyProtocolEsi,
				LogicalDeviceId:    "AOS-48x10_6x40-2",
			},
			{
				Label:              acctest.RandString(6),
				LinkPerSpineCount:  1,
				LinkPerSpineSpeed:  "40G",
				RedundancyProtocol: apstra.LeafRedundancyProtocolEsi,
				LogicalDeviceId:    "AOS-48x10_6x40-2",
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteRackType(ctx, rackId)) })

	var aap *apstra.AntiAffinityPolicy
	if compatibility.TemplateRequiresAntiAffinityPolicy.Check(version.Must(version.NewVersion(client.ApiVersion()))) {
		aap = &apstra.AntiAffinityPolicy{
			Algorithm: apstra.AlgorithmHeuristic,
			Mode:      apstra.AntiAffinityModeDisabled,
		}
	}

	templateId, err := client.CreateRackBasedTemplate(ctx, &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(6),
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         1,
			LogicalDevice: "AOS-16x40-1",
		},
		RackInfos:            map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{rackId: {Count: 1}},
		AsnAllocationPolicy:  &apstra.AsnAllocationPolicy{SpineAsnScheme: apstra.AsnAllocationSchemeDistinct},
		VirtualNetworkPolicy: &apstra.VirtualNetworkPolicy{OverlayControlProtocol: apstra.OverlayControlProtocolEvpn},
		AntiAffinityPolicy:   aap,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteTemplate(ctx, templateId)) })

	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  enum.RefDesignDatacenter,
		Label:      acctest.RandString(10),
		TemplateId: templateId,
		FabricSettings: &apstra.FabricSettings{
			SpineSuperspineLinks: utils.ToPtr(apstra.AddressingSchemeIp4),
			SpineLeafLinks:       utils.ToPtr(apstra.AddressingSchemeIp4),
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	require.NoError(t, err)

	// set interface map on all leafs
	leafs := GetSystemIds(t, ctx, bpClient, "leaf")
	assignents := make(apstra.SystemIdToInterfaceMapAssignment, len(leafs))
	for _, leaf := range leafs {
		assignents[leaf.String()] = "Juniper_QFX5100-48S__AOS-48x10_6x40-2"
	}
	err = bpClient.SetInterfaceMapAssignments(ctx, assignents)
	require.NoError(t, err)

	// enable IPv6
	settings, err := bpClient.GetFabricSettings(ctx)
	require.NoError(t, err)
	settings.Ipv6Enabled = utils.ToPtr(true)
	err = bpClient.SetFabricSettings(ctx, settings)
	require.NoError(t, err)

	return bpClient
}

func BlueprintG(t testing.TB, ctx context.Context, cleanup bool) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)

	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  enum.RefDesignDatacenter,
		Label:      acctest.RandString(8),
		TemplateId: "L2_Virtual_EVPN",
		FabricSettings: &apstra.FabricSettings{
			SpineSuperspineLinks: utils.ToPtr(apstra.AddressingSchemeIp4),
			SpineLeafLinks:       utils.ToPtr(apstra.AddressingSchemeIp4),
		},
	})
	require.NoError(t, err)
	if cleanup {
		t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })
	}

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	require.NoError(t, err)

	return bpClient
}

//func BlueprintH(t testing.TB, ctx context.Context, cleanup bool) *apstra.TwoStageL3ClosClient {
//}

func BlueprintI(t testing.TB, ctx context.Context) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)

	bpId, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  enum.RefDesignDatacenter,
		Label:      acctest.RandString(6),
		TemplateId: "L3_Collapsed_ESI",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, bpId)) })

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, bpId)
	require.NoError(t, err)

	// assign leaf interface maps
	leafIds := GetSystemIds(t, ctx, bpClient, "leaf")
	mappings := make(apstra.SystemIdToInterfaceMapAssignment, len(leafIds))
	for _, leafId := range leafIds {
		mappings[leafId.String()] = "Juniper_vQFX__AOS-7x10-Leaf"
	}
	err = bpClient.SetInterfaceMapAssignments(ctx, mappings)
	require.NoError(t, err)

	// set leaf loopback pool
	err = bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: apstra.ResourceTypeIp4Pool,
			Name: apstra.ResourceGroupNameLeafIp4,
		},
		PoolIds: []apstra.ObjectId{"Private-10_0_0_0-8"},
	})
	require.NoError(t, err)

	// set leaf-leaf pool
	err = bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: apstra.ResourceTypeIp4Pool,
			Name: apstra.ResourceGroupNameLeafLeafIp4,
		},
		PoolIds: []apstra.ObjectId{"Private-10_0_0_0-8"},
	})
	require.NoError(t, err)

	// set leaf ASN pool
	err = bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: apstra.ResourceTypeAsnPool,
			Name: apstra.ResourceGroupNameLeafAsn,
		},
		PoolIds: []apstra.ObjectId{"Private-64512-65534"},
	})
	require.NoError(t, err)

	// set VN VNI pool
	err = bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: apstra.ResourceTypeVniPool,
			Name: apstra.ResourceGroupNameEvpnL3Vni,
		},
		PoolIds: []apstra.ObjectId{"Default-10000-20000"},
	})
	require.NoError(t, err)

	// set VN VNI pool
	err = bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: apstra.ResourceTypeVniPool,
			Name: apstra.ResourceGroupNameVxlanVnIds,
		},
		PoolIds: []apstra.ObjectId{"Default-10000-20000"},
	})
	require.NoError(t, err)

	// commit
	bpStatus, err := client.GetBlueprintStatus(ctx, bpClient.Id())
	require.NoError(t, err)
	_, err = client.DeployBlueprint(ctx, &apstra.BlueprintDeployRequest{
		Id:          bpClient.Id(),
		Description: "initial commit in test: " + t.Name(),
		Version:     bpStatus.Version,
	})
	require.NoError(t, err)

	return bpClient
}

func FfBlueprintA(t testing.TB, ctx context.Context) *apstra.FreeformClient {
	t.Helper()

	client := GetTestClient(t, ctx)

	id, err := client.CreateFreeformBlueprint(ctx, acctest.RandString(6))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

	bpClient, err := client.NewFreeformClient(ctx, id)
	require.NoError(t, err)

	return bpClient
}

func FfBlueprintB(t testing.TB, ctx context.Context, intSystemCount, extSystemCount int) (*apstra.FreeformClient, []apstra.ObjectId, []apstra.ObjectId) {
	t.Helper()

	client := GetTestClient(t, ctx)

	id, err := client.CreateFreeformBlueprint(ctx, acctest.RandString(6))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

	c, err := client.NewFreeformClient(ctx, id)
	require.NoError(t, err)

	dpId, err := c.ImportDeviceProfile(ctx, "Juniper_vEX")
	require.NoError(t, err)

	intSystemIds := make([]apstra.ObjectId, intSystemCount)
	for i := range intSystemIds {
		intSystemIds[i], err = c.CreateSystem(ctx, &apstra.FreeformSystemData{
			Type:            apstra.SystemTypeInternal,
			Label:           acctest.RandString(6),
			DeviceProfileId: &dpId,
		})
		require.NoError(t, err)
	}

	extSystemIds := make([]apstra.ObjectId, extSystemCount)
	for i := range extSystemIds {
		extSystemIds[i], err = c.CreateSystem(ctx, &apstra.FreeformSystemData{
			Type:  apstra.SystemTypeExternal,
			Label: acctest.RandString(6),
		})
		require.NoError(t, err)
	}

	return c, intSystemIds, extSystemIds
}

// FfBlueprintC creates a freeform blueprint with a single resource group inside.
// Returned values are the blueprint client and the resource group ID.
func FfBlueprintC(t testing.TB, ctx context.Context) (*apstra.FreeformClient, apstra.ObjectId) {
	t.Helper()

	client := GetTestClient(t, ctx)

	id, err := client.CreateFreeformBlueprint(ctx, acctest.RandString(6))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

	c, err := client.NewFreeformClient(ctx, id)
	require.NoError(t, err)

	group, err := c.CreateRaGroup(ctx, &apstra.FreeformRaGroupData{Label: acctest.RandString(6)})
	require.NoError(t, err)

	return c, group
}
