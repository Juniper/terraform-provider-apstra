package testutils

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func TemplateA(t testing.TB, ctx context.Context) *apstra.TemplateRackBased {
	t.Helper()

	client := GetTestClient(t, ctx)

	tagA := TagA(t, ctx)
	tagB := TagA(t, ctx)

	rackType, err := client.GetRackType(ctx, "one_leaf")
	require.NoError(t, err)

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         4,
			LogicalDevice: "AOS-16x40-1",
			Tags:          []apstra.ObjectId{tagA.Id, tagB.Id},
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			rackType.Id: {Count: 2},
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
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteTemplate(ctx, id)) })

	template, err := client.GetRackBasedTemplate(ctx, id)
	require.NoError(t, err)

	return template
}

func TemplateB(t testing.TB, ctx context.Context) *apstra.TemplateRackBased {
	t.Helper()

	client := GetTestClient(t, ctx)

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         1,
			LogicalDevice: "AOS-8x10-1",
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			"access_switch":      {Count: 3}, // single-single
			"L2_ESI_Access_dual": {Count: 2}, // ESI-ESI
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
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteTemplate(ctx, id)) })

	template, err := client.GetRackBasedTemplate(ctx, id)
	require.NoError(t, err)

	return template
}

func TemplateC(t testing.TB, ctx context.Context) *apstra.TemplateRackBased {
	t.Helper()

	client := GetTestClient(t, ctx)

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         1,
			LogicalDevice: "AOS-8x10-1",
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			"L2_ESI_Access_dual": {Count: 1}, // ESI-ESI
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
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteTemplate(ctx, id)) })

	template, err := client.GetRackBasedTemplate(ctx, id)
	require.NoError(t, err)

	return template
}

func TemplateD(t testing.TB, ctx context.Context) *apstra.TemplateRackBased {
	t.Helper()

	rackTypeA := RackTypeA(t, ctx)
	rackTypeB := RackTypeB(t, ctx)
	rackTypeC := RackTypeC(t, ctx)
	rackTypeD := RackTypeD(t, ctx)

	raid := rackTypeA.Id
	rbid := rackTypeB.Id
	rcid := rackTypeC.Id
	rdid := rackTypeD.Id

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
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

	client := GetTestClient(t, ctx)

	id, err := client.CreateRackBasedTemplate(ctx, templateRequest)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteTemplate(ctx, id)) })

	template, err := client.GetRackBasedTemplate(ctx, id)
	require.NoError(t, err)

	return template
}

func TemplateE(t testing.TB, ctx context.Context) *apstra.TemplateRackBased {
	t.Helper()

	client := GetTestClient(t, ctx)

	rackTypeF := RackTypeF(t, ctx)

	templateRequest := &apstra.CreateRackBasedTemplateRequest{
		DisplayName: acctest.RandString(10),
		Spine: &apstra.TemplateElementSpineRequest{
			Count:         1,
			LogicalDevice: "AOS-32x40-1",
		},
		RackInfos: map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo{
			rackTypeF.Id: {Count: 1},
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
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteTemplate(ctx, id)) })

	template, err := client.GetRackBasedTemplate(ctx, id)
	require.NoError(t, err)

	return template
}
