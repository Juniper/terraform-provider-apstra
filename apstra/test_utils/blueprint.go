package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"testing"
)

type bFunc func(t testing.TB, ctx context.Context, name ...string) *apstra.TwoStageL3ClosClient

func MakeOrFindBlueprint(t testing.TB, ctx context.Context, name string, f bFunc) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)

	status, err := client.GetBlueprintStatusByName(ctx, name)
	if err != nil {
		if utils.IsApstra404(err) {
			return f(t, ctx, name)
		}

		require.NoError(t, err)
	}

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
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
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
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
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
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
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
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
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

func BlueprintE(t testing.TB, ctx context.Context) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)
	template := TemplateD(t, ctx)
	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
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

func BlueprintF(t testing.TB, ctx context.Context) *apstra.TwoStageL3ClosClient {
	t.Helper()

	client := GetTestClient(t, ctx)
	template := TemplateE(t, ctx)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
		Label:      acctest.RandString(10),
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
