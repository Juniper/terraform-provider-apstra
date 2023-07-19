package testutils

import (
	"context"
	"errors"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func TemplateA(ctx context.Context) (*apstra.TemplateRackBased, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
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

func TemplateB(ctx context.Context) (*apstra.TemplateRackBased, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Capability:  apstra.TemplateCapabilityBlueprint,
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         1,
			LogicalDevice: "AOS-8x10-1",
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			"access_switch":      {Count: 3}, // single-single
			"L2_ESI_Access_dual": {Count: 2}, // ESI-ESI
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
		return client.DeleteTemplate(ctx, id)
	}

	template, err := client.GetRackBasedTemplate(ctx, id)
	return template, deleteFunc, err
}

func TemplateC(ctx context.Context) (*apstra.TemplateRackBased, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Capability:  apstra.TemplateCapabilityBlueprint,
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         1,
			LogicalDevice: "AOS-8x10-1",
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			"L2_ESI_Access_dual": {Count: 1}, // ESI-ESI
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
		return client.DeleteTemplate(ctx, id)
	}

	template, err := client.GetRackBasedTemplate(ctx, id)
	return template, deleteFunc, err
}

func TemplateD(ctx context.Context) (*apstra.TemplateRackBased, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	rackTypeA, rackTypeADelete, err := RackTypeA(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return rackTypeADelete(ctx)
	}

	rackTypeB, rackTypeBDelete, err := RackTypeB(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return errors.Join(
			rackTypeADelete(ctx),
			rackTypeBDelete(ctx),
		)
	}

	rackTypeC, rackTypeCDelete, err := RackTypeC(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return errors.Join(
			rackTypeADelete(ctx),
			rackTypeBDelete(ctx),
			rackTypeCDelete(ctx),
		)
	}

	rackTypeD, rackTypeDDelete, err := RackTypeD(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return errors.Join(
			rackTypeADelete(ctx),
			rackTypeBDelete(ctx),
			rackTypeCDelete(ctx),
			rackTypeDDelete(ctx),
		)
	}

	raid := rackTypeA.Id
	rbid := rackTypeB.Id
	rcid := rackTypeC.Id
	rdid := rackTypeD.Id

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Capability:  apstra.TemplateCapabilityBlueprint,
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         1,
			LogicalDevice: "AOS-8x10-1",
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			raid: {Count: 1},
			rbid: {Count: 1},
			rcid: {Count: 1},
			rdid: {Count: 1},
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
		return errors.Join(
			rackTypeADelete(ctx),
			rackTypeBDelete(ctx),
			rackTypeCDelete(ctx),
			rackTypeDDelete(ctx),
			client.DeleteTemplate(ctx, id),
		)
	}

	template, err := client.GetRackBasedTemplate(ctx, id)
	return template, deleteFunc, err
}

func TemplateE(ctx context.Context) (*apstra.TemplateRackBased, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }

	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	rackTypeF, rackTypeFDelete, err := RackTypeF(ctx)
	if err != nil {
		return nil, nil, err
	}
	deleteFunc = func(ctx context.Context) error {
		return rackTypeFDelete(ctx)
	}

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Capability:  apstra.TemplateCapabilityBlueprint,
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         1,
			LogicalDevice: "AOS-32x40-1",
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			rackTypeF.Id: {Count: 1},
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
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}
	deleteFunc = func(ctx context.Context) error {
		return errors.Join(
			rackTypeFDelete(ctx),
			client.DeleteTemplate(ctx, id),
		)
	}

	template, err := client.GetRackBasedTemplate(ctx, id)
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}
	return template, deleteFunc, err
}
