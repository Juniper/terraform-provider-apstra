//go:build integration

package designtestobjects

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/design"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

// InterfaceMapA uses the LogicalDeviceB() test helper to create Logical Device links that LD
// to a Juniper_QFX5120-48T_Junos (device profile) interfaces xe-0/0/* using transform #1.
func InterfaceMapA(ctx context.Context, t testing.TB, client *apstra.Client) string {
	t.Helper()

	dpID := "Juniper_QFX5120-48T_Junos"

	// Use a test helper to create a logical device.
	ldID := LogicalDeviceB(ctx, t, client)

	// Count the interfaces on the LD.
	ld, err := client.GetLogicalDevice2(ctx, ldID)
	require.NoError(t, err)
	ldPortCount := 0
	for _, panel := range ld.Panels {
		ldPortCount += panel.PanelLayout.ColumnCount * panel.PanelLayout.RowCount
	}

	dp, err := client.GetDeviceProfile(ctx, dpID)
	require.NoError(t, err)
	dpPortCount := len(dp.Ports)

	interfaces := make([]design.InterfaceMapInterface, dpPortCount)
	for i := range interfaces {
		interfaces[i] = design.InterfaceMapInterface{
			Name:     dp.Ports[i].Transformations[0].Interfaces[0].Name,
			Roles:    []enum.PortRole{enum.PortRoleAccess, enum.PortRoleGeneric, enum.PortRoleLeaf, enum.PortRolePeer, enum.PortRoleSpine, enum.PortRoleSuperspine, enum.PortRoleUnused},
			Position: i + 1,
			State:    enum.InterfaceMapInterfaceStateActive,
			Speed:    "10G",
			Setting: struct {
				Param string `json:"param"`
			}{
				Param: *dp.Ports[i].Transformations[0].Interfaces[0].Setting,
			},
			Mapping: design.InterfaceMapInterfaceMapping{
				DeviceProfilePortID:      i + 1,
				DeviceProfileTransformID: 1,
				DeviceProfileInterfaceID: 1,
				LogicalDevicePanel:       pointer.To(1),
				LogicalDevicePort:        pointer.To(i + 1),
			},
		}
	}

	for i := range dpPortCount - ldPortCount {
		i = i + ldPortCount
		interfaces[i] = design.InterfaceMapInterface{
			Name:     dp.Ports[i].Transformations[0].Interfaces[0].Name,
			Roles:    []enum.PortRole{enum.PortRoleUnused},
			Position: i + 1,
			State:    enum.InterfaceMapInterfaceStateActive,
			Speed:    dp.Ports[i].Transformations[0].Interfaces[0].Speed,
			Setting: struct {
				Param string `json:"param"`
			}{
				Param: *dp.Ports[i].Transformations[0].Interfaces[0].Setting,
			},
			Mapping: design.InterfaceMapInterfaceMapping{
				DeviceProfilePortID:      i + 1,
				DeviceProfileTransformID: 1,
				DeviceProfileInterfaceID: 1,
			},
		}
	}

	imID, err := client.CreateInterfaceMap2(ctx, design.InterfaceMap{
		Label:           acctest.RandString(6),
		DeviceProfileID: "Juniper_QFX5120-48T_Junos",
		LogicalDeviceID: ldID,
		Interfaces:      interfaces,
	})
	require.NoError(t, err)

	testutils.CleanupWithFreshContext(t, testutils.DefaultCleanupTimeout, func(ctx context.Context) error {
		return client.DeleteInterfaceMap2(ctx, imID)
	})

	return imID
}
