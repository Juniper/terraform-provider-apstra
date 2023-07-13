package testutils

import (
	"context"
	"errors"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// RackTypeA has:
// - 1 leaf switch with 10G uplink
// - 1 access switch
func RackTypeA(ctx context.Context) (*apstra.RackType, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	leafLabel := "rack type A leaf"

	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-A-" + acctest.RandString(10),
		FabricConnectivityDesign: apstra.FabricConnectivityDesignL3Clos,
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
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeleteRackType(ctx, id)
	}

	result, err := client.GetRackType(ctx, id)
	return result, deleteFunc, err
}

// RackTypeB has:
// - 1 leaf switch ESI pair with 10G uplink
// - 1 access switch dual-homed to ESI leaf "A"
// - 1 pair (count = 2) access switches single-homed to ESI leaf "B"
func RackTypeB(ctx context.Context) (*apstra.RackType, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	leafLabel := "rack type B leaf"

	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-B-" + acctest.RandString(10),
		FabricConnectivityDesign: apstra.FabricConnectivityDesignL3Clos,
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
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeleteRackType(ctx, id)
	}

	result, err := client.GetRackType(ctx, id)
	return result, deleteFunc, err
}

// RackTypeC has:
// - 1 leaf switch ESI pair with 10G uplink
// - 1 access switch dual-homed to both ESI leaf switches
func RackTypeC(ctx context.Context) (*apstra.RackType, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	leafLabel := "rack type C leaf"

	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-C-" + acctest.RandString(10),
		FabricConnectivityDesign: apstra.FabricConnectivityDesignL3Clos,
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
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeleteRackType(ctx, id)
	}

	result, err := client.GetRackType(ctx, id)
	return result, deleteFunc, err
}

// RackTypeD has:
// - 1 leaf switch ESI pair with 10G uplink
// - 1 access switch single homed to ESI leaf "A"
// - 1 access switch ESI pair
// - 1 access switch single homed to ESI leaf "B"
func RackTypeD(ctx context.Context) (*apstra.RackType, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	leafLabel := "rack type D leaf"

	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-D-" + acctest.RandString(10),
		FabricConnectivityDesign: apstra.FabricConnectivityDesignL3Clos,
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
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeleteRackType(ctx, id)
	}

	result, err := client.GetRackType(ctx, id)
	return result, deleteFunc, err
}

// RackTypeE has:
// - 1 leaf switch MLAG pair with 10G uplink
// - 1 access switch homed to the first MLAG peer
// - 1 access switch homed to the second MLAG peer
// - 1 access switch homed to both MLAG peers
func RackTypeE(ctx context.Context) (*apstra.RackType, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }
	client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	leafLabel := "rack type E leaf"

	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "aaa-E-" + acctest.RandString(10),
		FabricConnectivityDesign: apstra.FabricConnectivityDesignL3Clos,
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
	if err != nil {
		return nil, deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeleteRackType(ctx, id)
	}

	result, err := client.GetRackType(ctx, id)
	return result, deleteFunc, err
}

func RackTypeF(ctx context.Context) (*apstra.RackType, func(context.Context) error, error) {
	deleteFunc := func(ctx context.Context) error { return nil }

  client, err := GetTestClient(ctx)
	if err != nil {
		return nil, deleteFunc, err
	}

	id, err := client.CreateRackType(ctx, &apstra.RackTypeRequest{
		DisplayName:              "rack type F",
		FabricConnectivityDesign: apstra.FabricConnectivityDesignL3Clos,
		LeafSwitches: []apstra.RackElementLeafSwitchRequest{
			{
				Label:             "rack type F leaf",
				LinkPerSpineCount: 1,
				LinkPerSpineSpeed: "40G",
				LogicalDeviceId:   "AOS-48x10_6x40-1",
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeleteRackType(ctx, id)
	}

	result, err := client.GetRackType(ctx, id)
	if err != nil {
		return nil, nil, errors.Join(err, deleteFunc(ctx))
	}
	return result, deleteFunc, err
}
