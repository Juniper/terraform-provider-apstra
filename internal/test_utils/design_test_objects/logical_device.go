package designtestobjects

import (
	"context"
	"testing"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/design"
	"github.com/Juniper/apstra-go-sdk/enum"
	testutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

func LogicalDeviceA(ctx context.Context, t testing.TB, client *apstra.Client) string {
	t.Helper()

	portRoles := design.LogicalDevicePortRoles{
		enum.PortRoleAccess, enum.PortRoleGeneric, enum.PortRoleLeaf, enum.PortRolePeer,
		enum.PortRoleSpine, enum.PortRoleSuperspine, enum.PortRoleUnused,
	}

	id, err := client.CreateLogicalDevice2(ctx, design.LogicalDevice{
		Label: acctest.RandString(6),
		Panels: []design.LogicalDevicePanel{
			{
				PanelLayout: design.LogicalDevicePanelLayout{RowCount: 4, ColumnCount: 24},
				PortGroups: []design.LogicalDevicePanelPortGroup{
					{Count: 24, Speed: "1G", Roles: portRoles},
					{Count: 24, Speed: "10G", Roles: portRoles},
					{Count: 24, Speed: "25G", Roles: portRoles},
					{Count: 24, Speed: "100G", Roles: portRoles},
				},
				PortIndexing: enum.DesignLogicalDevicePanelPortIndexingLRTB,
			},
		},
	})
	require.NoError(t, err)

	testutils.CleanupWithFreshContext(t, 10*time.Second, func(ctx context.Context) error {
		return client.DeleteLogicalDevice2(ctx, id)
	})

	return id
}
