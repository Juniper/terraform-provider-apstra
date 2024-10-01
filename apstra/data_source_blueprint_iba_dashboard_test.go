package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	dataSourceBlueprintIbaDashboardTemplateByNameHCL = `
	data "apstra_blueprint_iba_dashboard" "test" {
    	blueprint_id = "%s"
		name = "%s"
	}
	`

	dataSourceBlueprintIbaDashboardTemplateByIdHCL = `
	data "apstra_blueprint_iba_dashboard" "test" {
  		blueprint_id = "%s"
		id = "%s"
	}
	`
)

func TestAccDataSourceIbaDashboard(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)
	clientVersion := version.Must(version.NewVersion(client.ApiVersion()))
	if !compatibility.BpIbaDashboardOk.Check(clientVersion) {
		t.Skipf("skipping due to version constraint %s", compatibility.BpIbaDashboardOk)
	}

	bpClient := testutils.MakeOrFindBlueprint(t, ctx, "BPA", testutils.BlueprintA)

	// Set up Widgets
	widgetIdA, _, widgetIdB, _ := testutils.TestWidgetsAB(t, ctx, bpClient)

	dashboardData := apstra.IbaDashboardData{
		Description:   "Test Dashboard",
		Default:       false,
		Label:         "Test Dash",
		IbaWidgetGrid: [][]apstra.ObjectId{{widgetIdA, widgetIdB}},
	}

	id, err := bpClient.CreateIbaDashboard(ctx, &dashboardData)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, bpClient.DeleteIbaDashboard(ctx, id)) })

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceBlueprintIbaDashboardTemplateByIdHCL, bpClient.Id().String(), id.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "id", id.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "name", dashboardData.Label),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "widget_grid.0.0", widgetIdA.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "widget_grid.0.1", widgetIdB.String()),
				),
			},
			// Read by Name
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceBlueprintIbaDashboardTemplateByNameHCL, bpClient.Id().String(), dashboardData.Label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "id", id.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "name", dashboardData.Label),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "widget_grid.0.0", widgetIdA.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "widget_grid.0.1", widgetIdB.String()),
				),
			},
		},
	})
}
