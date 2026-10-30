//go:build integration

package designtestobjects

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/design"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/apstra-go-sdk/speed"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

// RackTypeA creates an L3Clos rack type with an ESI leaf pair and two standalone leafs.
//
// Each standalone leaf has:
//   - 3 accesss switches.
//
// Each ESI pair has:
//   - 4 single-homed access switches (each connects to one leaf switch)
//   - 2 dual-homed access switches (each connects to both leaf switches)
//   - 2 dual-homed access switch pairs (each access pair member is dual-homed)
//   - 2 single-homed access switch pairs (each member of the pair connects to one leaf switch)
func RackTypeA(ctx context.Context, t testing.TB, client *apstra.Client) string {
	t.Helper()

	imID := InterfaceMapA(ctx, t, client)
	im, err := client.GetInterfaceMap2(ctx, imID)
	require.NoError(t, err)

	ldID := im.LogicalDeviceID
	ld, err := client.GetLogicalDevice2(ctx, ldID)

	rtID, err := client.CreateRackType2(ctx, design.RackType{
		Label:                    acctest.RandString(6),
		FabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
		LeafSwitches: []design.RackTypeLeafSwitch{
			{
				Label:              "sl_1",
				LinkPerSpineCount:  pointer.To(1),
				LinkPerSpineSpeed:  pointer.To(speed.Speed("10G")),
				LogicalDevice:      ld,
				RedundancyProtocol: enum.LeafRedundancyProtocolNone,
			},
			{
				Label:              "sl_2",
				LinkPerSpineCount:  pointer.To(1),
				LinkPerSpineSpeed:  pointer.To(speed.Speed("10G")),
				LogicalDevice:      ld,
				RedundancyProtocol: enum.LeafRedundancyProtocolNone,
			},
			{
				Label:              "pl",
				LinkPerSpineCount:  pointer.To(1),
				LinkPerSpineSpeed:  pointer.To(speed.Speed("10G")),
				LogicalDevice:      ld,
				RedundancyProtocol: enum.LeafRedundancyProtocolESI,
			},
		},
		AccessSwitches: []design.RackTypeAccessSwitch{
			{
				Count: 1,
				Label: "sl_1-sa_1",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "sl_1",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "sl_1-sa_2",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "sl_1",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "sl_1-sa_3",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "sl_1",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "sl_2-sa_1",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "sl_2",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "sl_2-sa_2",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "sl_2",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "sl_2-sa_3",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "sl_2",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_1-sa_1",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerFirst,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_1-sa_2",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerFirst,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_2-sa_1",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerSecond,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_2-sa_2",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerSecond,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_12-sa_1",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeDual,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_12-sa_2",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeDual,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_12-pa_1",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeDual,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				ESILAGInfo: &design.RackTypeAccessSwitchESILAGInfo{
					LinkCount: 1,
					LinkSpeed: "10G",
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_12-pa_2",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeDual,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				ESILAGInfo: &design.RackTypeAccessSwitchESILAGInfo{
					LinkCount: 1,
					LinkSpeed: "10G",
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_12-pa_3",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				ESILAGInfo: &design.RackTypeAccessSwitchESILAGInfo{
					LinkCount: 1,
					LinkSpeed: "10G",
				},
				LogicalDevice: ld,
			},
			{
				Count: 1,
				Label: "pl_12-pa_4",
				Links: []design.RackTypeLink{
					{
						Label:              "l",
						TargetSwitchLabel:  "pl",
						LinkPerSwitchCount: 1,
						Speed:              "10G",
						AttachmentType:     enum.LinkAttachmentTypeSingle,
						LAGMode:            enum.LAGModeActiveLACP,
						SwitchPeer:         enum.LinkSwitchPeerUnspecified,
					},
				},
				ESILAGInfo: &design.RackTypeAccessSwitchESILAGInfo{
					LinkCount: 1,
					LinkSpeed: "10G",
				},
				LogicalDevice: ld,
			},
		},
	})
	require.NoError(t, err)

	testutils.CleanupWithFreshContext(t, testutils.DefaultCleanupTimeout, func(ctx context.Context) error {
		return client.DeleteRackType2(ctx, rtID)
	})

	return rtID
}
