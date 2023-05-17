package testutils

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func BlueprintA(ctx context.Context) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient()
	if err != nil {
		return nil, deleteFunc, err
	}

	name := acctest.RandString(10)
	id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  apstra.RefDesignDatacenter,
		Label:      name,
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
	client, err := GetTestClient()
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
		RefDesign:  apstra.RefDesignDatacenter,
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
	client, err := GetTestClient()
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
		RefDesign:  apstra.RefDesignDatacenter,
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
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient()
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
		RefDesign:  apstra.RefDesignDatacenter,
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

func BlueprintE(ctx context.Context) (*apstra.TwoStageL3ClosClient, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient()
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
		RefDesign:  apstra.RefDesignDatacenter,
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
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient()
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
		RefDesign:  apstra.RefDesignDatacenter,
		Label:      name,
		TemplateId: template.Id,
		FabricAddressingPolicy: &apstra.FabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp46,
			SpineLeafLinks:       apstra.AddressingSchemeIp46,
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
