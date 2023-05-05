package testutils

import (
	"context"
	"errors"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func TemplateA(ctx context.Context) (*apstra.TemplateRackBased, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient()
	if err != nil {
		return nil, deleteFunc, err
	}

	tagA, tagADelete, err := TagA(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return tagADelete(ctx)
	}

	tagB, tagBDelete, err := TagA(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return errors.Join(tagADelete(ctx), tagBDelete(ctx))
	}

	rackType, err := client.GetRackType(ctx, "one_leaf")
	if err != nil {
		return nil, deleteFunc, err
	}

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Capability:  apstra.TemplateCapabilityBlueprint,
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         4,
			LogicalDevice: "AOS-16x40-1",
			Tags:          []apstra.ObjectId{tagA.Id, tagB.Id},
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			rackType.Id: {Count: 2},
		},
		FabricAddressingPolicy: &apstra.FabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4,
			SpineLeafLinks:       apstra.AddressingSchemeIp4,
		},
		AntiAffinityPolicy: &apstra.AntiAffinityPolicy{
			Algorithm:                apstra.AlgorithmHeuristic,
			MaxLinksPerPort:          1,
			MaxLinksPerSlot:          1,
			MaxPerSystemLinksPerPort: 1,
			MaxPerSystemLinksPerSlot: 1,
			Mode:                     apstra.AntiAffinityModeDisabled,
		},
		AsnAllocationPolicy:  &apstra.AsnAllocationPolicy{SpineAsnScheme: apstra.AsnAllocationSchemeDistinct},
		VirtualNetworkPolicy: &apstra.VirtualNetworkPolicy{OverlayControlProtocol: apstra.OverlayControlProtocolEvpn},
	}
	id, err := client.CreateRackBasedTemplate(ctx, templateRequest)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return errors.Join(tagADelete(ctx), tagBDelete(ctx), client.DeleteTemplate(ctx, id))
	}

	template, err := client.GetRackBasedTemplate(ctx, id)
	return template, deleteFunc, err
}