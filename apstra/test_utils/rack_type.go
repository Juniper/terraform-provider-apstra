package testutils

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

// RackTypeA has:
// - 1 leaf switch with 10G uplink
// - 1 access switch
func RackTypeA(t testing.TB, ctx context.Context) *apstra.RackType {
	t.Helper()

	client := GetTestClient(t, ctx)

	leafLabel := "rack type A leaf"
	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-A-" + acctest.RandString(10),
		FabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
		LeafSwitches: []apstra.RackElementLeafSwitchRequest{
			{
				Label:             leafLabel,
				LinkPerSpineCount: 1,
				LinkPerSpineSpeed: "10G",
				LogicalDeviceId:   "AOS-9x10-Leaf",
			},
		},
		AccessSwitches: []apstra.RackElementAccessSwitchRequest{
			{
				Label:           "rack type A access",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteRackType(ctx, id)) })

	result, err := client.GetRackType(ctx, id)
	require.NoError(t, err)

	return result
}

// RackTypeB has:
// - 1 leaf switch ESI pair with 10G uplink
// - 1 access switch dual-homed to ESI leaf "A"
// - 1 pair (count = 2) access switches single-homed to ESI leaf "B"
func RackTypeB(t testing.TB, ctx context.Context) *apstra.RackType {
	t.Helper()

	client := GetTestClient(t, ctx)

	leafLabel := "rack type B leaf"
	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-B-" + acctest.RandString(10),
		FabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
		LeafSwitches: []apstra.RackElementLeafSwitchRequest{
			{
				Label:              leafLabel,
				LinkPerSpineCount:  1,
				LinkPerSpineSpeed:  "10G",
				LogicalDeviceId:    "AOS-9x10-Leaf",
				RedundancyProtocol: apstra.LeafRedundancyProtocolEsi,
			},
		},
		AccessSwitches: []apstra.RackElementAccessSwitchRequest{
			{
				Label:           "rack type B access 1",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 2,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						SwitchPeer:         apstra.RackLinkSwitchPeerFirst,
					},
				},
			},
			{
				Label:           "rack type B access 2",
				InstanceCount:   2,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						SwitchPeer:         apstra.RackLinkSwitchPeerSecond,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteRackType(ctx, id)) })

	result, err := client.GetRackType(ctx, id)
	require.NoError(t, err)
	return result
}

// RackTypeC has:
// - 1 leaf switch ESI pair with 10G uplink
// - 1 access switch dual-homed to both ESI leaf switches
func RackTypeC(t testing.TB, ctx context.Context) *apstra.RackType {
	t.Helper()

	client := GetTestClient(t, ctx)

	leafLabel := "rack type C leaf"
	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-C-" + acctest.RandString(10),
		FabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
		LeafSwitches: []apstra.RackElementLeafSwitchRequest{
			{
				Label:              leafLabel,
				LinkPerSpineCount:  1,
				LinkPerSpineSpeed:  "10G",
				LogicalDeviceId:    "AOS-9x10-Leaf",
				RedundancyProtocol: apstra.LeafRedundancyProtocolEsi,
			},
		},
		AccessSwitches: []apstra.RackElementAccessSwitchRequest{
			{
				Label:           "rack type C access",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						AttachmentType:     apstra.RackLinkAttachmentTypeDual,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteRackType(ctx, id)) })

	result, err := client.GetRackType(ctx, id)
	require.NoError(t, err)

	return result
}

// RackTypeD has:
// - 1 leaf switch ESI pair with 10G uplink
// - 1 access switch single homed to ESI leaf "A"
// - 1 access switch ESI pair
// - 1 access switch single homed to ESI leaf "B"
func RackTypeD(t testing.TB, ctx context.Context) *apstra.RackType {
	t.Helper()

	client := GetTestClient(t, ctx)

	leafLabel := "rack type D leaf"
	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-D-" + acctest.RandString(10),
		FabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
		LeafSwitches: []apstra.RackElementLeafSwitchRequest{
			{
				Label:              leafLabel,
				LinkPerSpineCount:  1,
				LinkPerSpineSpeed:  "10G",
				LogicalDeviceId:    "AOS-9x10-Leaf",
				RedundancyProtocol: apstra.LeafRedundancyProtocolEsi,
			},
		},
		AccessSwitches: []apstra.RackElementAccessSwitchRequest{
			{
				Label:           "rack type D access 1",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						AttachmentType:     apstra.RackLinkAttachmentTypeSingle,
						SwitchPeer:         apstra.RackLinkSwitchPeerFirst,
					},
				},
			},
			{
				Label:           "rack type D access 2",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						AttachmentType:     apstra.RackLinkAttachmentTypeDual,
					},
				},
				RedundancyProtocol: apstra.AccessRedundancyProtocolEsi,
				EsiLagInfo: &apstra.EsiLagInfo{
					AccessAccessLinkCount: 1,
					AccessAccessLinkSpeed: "10G",
				},
			},
			{
				Label:           "rack type D access 3",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						AttachmentType:     apstra.RackLinkAttachmentTypeSingle,
						SwitchPeer:         apstra.RackLinkSwitchPeerSecond,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteRackType(ctx, id)) })

	result, err := client.GetRackType(ctx, id)
	require.NoError(t, err)

	return result
}

// RackTypeE has:
// - 1 leaf switch MLAG pair with 10G uplink
// - 1 access switch homed to the first MLAG peer
// - 1 access switch homed to the second MLAG peer
// - 1 access switch homed to both MLAG peers
func RackTypeE(t testing.TB, ctx context.Context) *apstra.RackType {
	t.Helper()

	client := GetTestClient(t, ctx)

	leafLabel := "rack type E leaf"
	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-E-" + acctest.RandString(10),
		FabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
		LeafSwitches: []apstra.RackElementLeafSwitchRequest{
			{
				Label:              leafLabel,
				LinkPerSpineCount:  1,
				LinkPerSpineSpeed:  "10G",
				LogicalDeviceId:    "AOS-9x10-Leaf",
				RedundancyProtocol: apstra.LeafRedundancyProtocolMlag,
				MlagInfo: &apstra.LeafMlagInfo{
					LeafLeafLinkCount:         1,
					LeafLeafLinkPortChannelId: 1,
					LeafLeafLinkSpeed:         "10G",
					MlagVlanId:                1,
				},
			},
		},
		AccessSwitches: []apstra.RackElementAccessSwitchRequest{
			{
				Label:           "rack type E access 1",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						AttachmentType:     apstra.RackLinkAttachmentTypeSingle,
						SwitchPeer:         apstra.RackLinkSwitchPeerFirst,
					},
				},
			},
			{
				Label:           "rack type E access 2",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						AttachmentType:     apstra.RackLinkAttachmentTypeSingle,
						SwitchPeer:         apstra.RackLinkSwitchPeerSecond,
					},
				},
			},
			{
				Label:           "rack type E access 3",
				InstanceCount:   1,
				LogicalDeviceId: "AOS-9x10-Leaf",
				Links: []apstra.RackLinkRequest{
					{
						Label:              acctest.RandString(10),
						TargetSwitchLabel:  leafLabel,
						LinkPerSwitchCount: 1,
						LinkSpeed:          "10G",
						LagMode:            apstra.RackLinkLagModeActive,
						AttachmentType:     apstra.RackLinkAttachmentTypeDual,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteRackType(ctx, id)) })

	result, err := client.GetRackType(ctx, id)
	require.NoError(t, err)

	return result
}

func RackTypeF(t testing.TB, ctx context.Context) *apstra.RackType {
	t.Helper()

	client := GetTestClient(t, ctx)

	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "type F - " + acctest.RandString(5),
		FabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
		LeafSwitches: []apstra.RackElementLeafSwitchRequest{
			{
				Label:             "rack type F leaf",
				LinkPerSpineCount: 1,
				LinkPerSpineSpeed: "40G",
				LogicalDeviceId:   "AOS-48x10_6x40-1",
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteRackType(ctx, id)) })

	result, err := client.GetRackType(ctx, id)
	require.NoError(t, err)

	return result
}
