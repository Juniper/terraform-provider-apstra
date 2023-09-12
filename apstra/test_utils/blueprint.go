package testutils

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

type Bfunc func(ctx context.Context, name ...string) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error)

func MakeOrFindBlueprint(ctx context.Context, name string, f Bfunc) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }

	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	status, err := client.GetBlueprintStatusByName(ctx, name)
	if err != nil {
		if utils.IsApstra404(err) {
			return f(ctx, name)
		}
		return nil, deleteFunc, err
	}

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, status.Id)
	if err != nil {
		return nil, deleteFunc, err
	}

	return bpClient, deleteFunc, nil
}

func BlueprintA(ctx context.Context, name ...string) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }

	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

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
		FabricAddressingPolicy: &apstra.FabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4,
			SpineLeafLinks:       apstra.AddressingSchemeIp4,
		},
	})

	if err != nil {
		return nil, deleteFunc, fmt.Errorf("error creating blueprint %w", err)
	}

	deleteFunc = func(ctx context.Context) error {
		err := client.DeleteBlueprint(ctx, id)
		if err != nil {
			return err
		}
		return nil
	}

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	if err != nil {
		return nil, deleteFunc, err
	}

	return bpClient, deleteFunc, nil
}

func BlueprintB(ctx context.Context) (*apstra.TwoStageL3ClosClient, apstra.ObjectId, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, "", deleteFunc, err
	}

	template, templateDelete, err := TemplateA(ctx)
	if err != nil {
		return nil, "", deleteFunc, errors.Join(err, templateDelete(ctx))
	}
	deleteFunc = func(ctx context.Context) error {
		return templateDelete(ctx)
	}

	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
		Label:      name,
		TemplateId: template.Id,
		FabricAddressingPolicy: &apstra.FabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4,
			SpineLeafLinks:       apstra.AddressingSchemeIp4,
		},
	})
	if err != nil {
		return nil, template.Id, deleteFunc, fmt.Errorf("error creating blueprint %w", err)
	}

	deleteFunc = func(ctx context.Context) error {
		return errors.Join(client.DeleteBlueprint(ctx, id), templateDelete(ctx))
	}

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	if err != nil {
		return nil, template.Id, deleteFunc, err
	}

	return bpClient, template.Id, deleteFunc, nil
}

func BlueprintC(ctx context.Context) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	template, templateDelete, err := TemplateB(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return templateDelete(ctx)
	}

	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
		Label:      name,
		TemplateId: template.Id,
		FabricAddressingPolicy: &apstra.FabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4,
			SpineLeafLinks:       apstra.AddressingSchemeIp4,
		},
	})
	if err != nil {
		return nil, deleteFunc, fmt.Errorf("error creating blueprint %w", err)
	}

	deleteFunc = func(ctx context.Context) error {
		return client.DeleteBlueprint(ctx, id)
	}

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	if err != nil {
		return nil, deleteFunc, err
	}

	return bpClient, deleteFunc, nil
}

func BlueprintD(ctx context.Context) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error) {
	client, err := GetTestClient(ctx)
	deleteFunc := func(ctx context.Context) error { return nil }
	if err != nil {
		return nil, deleteFunc, err
	}

	template, templateDelete, err := TemplateC(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return templateDelete(ctx)
	}

	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
		Label:      name,
		TemplateId: template.Id,
		FabricAddressingPolicy: &apstra.FabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4,
			SpineLeafLinks:       apstra.AddressingSchemeIp4,
		},
	})
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}

	deleteFunc = func(ctx context.Context) error {
		return errors.Join(templateDelete(ctx), client.DeleteBlueprint(ctx, id))
	}

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}

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
	err = leafQuery.Do(ctx, &leafQueryResult)
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}

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
	err = accessQuery.Do(ctx, &accessQueryResult)
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}

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

	err = bpClient.SetInterfaceMapAssignments(ctx, assignments)
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}

	return bpClient, deleteFunc, nil
}

func BlueprintE(ctx context.Context) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	template, templateDelete, err := TemplateD(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return templateDelete(ctx)
	}

	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
		Label:      name,
		TemplateId: template.Id,
		FabricAddressingPolicy: &apstra.FabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4,
			SpineLeafLinks:       apstra.AddressingSchemeIp4,
		},
	})
	if err != nil {
		return nil, deleteFunc, fmt.Errorf("error creating blueprint %w", err)
	}
	deleteFunc = func(ctx context.Context) error {
		return errors.Join(client.DeleteBlueprint(ctx, id), templateDelete(ctx))
	}

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	if err != nil {
		return nil, deleteFunc, err
	}

	return bpClient, deleteFunc, nil
}

func BlueprintF(ctx context.Context) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error) {
	client, err := GetTestClient(ctx)
	deleteFunc := func(ctx context.Context) error { return nil }
	if err != nil {
		return nil, deleteFunc, err
	}

	template, templateDelete, err := TemplateE(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return templateDelete(ctx)
	}

	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  apstra.RefDesignTwoStageL3Clos,
		Label:      name,
		TemplateId: template.Id,
		FabricAddressingPolicy: &apstra.FabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4,
			SpineLeafLinks:       apstra.AddressingSchemeIp4,
		},
	})
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}
	deleteFunc = func(ctx context.Context) error {
		return errors.Join(client.DeleteBlueprint(ctx, id), templateDelete(ctx))
	}

	bpClient, err := client.NewTwoStageL3ClosClient(ctx, id)
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}

	return bpClient, deleteFunc, nil
}
