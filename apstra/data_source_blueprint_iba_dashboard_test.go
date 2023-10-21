package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
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

	bpClient, bpDelete, err := testutils.MakeOrFindBlueprint(ctx, "BPA", testutils.BlueprintA)

	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}
	defer func() {
		err = bpDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	// Set up Widgets
	widgetIdA, _, widgetIdB, _, cleanup := testutils.TestWidgetsAB(ctx, t, bpClient)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = cleanup()
		if err != nil {
			t.Error(err)
		}
	}()

	data := apstra.IbaDashboardData{
		Description:   "Test Dashboard",
		Default:       false,
		Label:         "Test Dash",
		IbaWidgetGrid: [][]apstra.ObjectId{{widgetIdA, widgetIdB}},
	}
	dId, err := bpClient.CreateIbaDashboard(ctx, &data)
	if err != nil {
		t.Fatal(err)
	}
	defer bpClient.DeleteIbaDashboard(ctx, dId)

	resource.Test(t, resource.TestCase{
		// PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceBlueprintIbaDashboardTemplateByIdHCL,
					bpClient.Id().String(), dId.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "id", dId.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "name", data.Label),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "widget_grid.0.0", widgetIdA.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "widget_grid.0.1", widgetIdB.String()),
				),
			},
			// Read by Name
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceBlueprintIbaDashboardTemplateByNameHCL,
					bpClient.Id().String(), data.Label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "id", dId.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "name", data.Label),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "widget_grid.0.0", widgetIdA.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_dashboard.test", "widget_grid.0.1", widgetIdB.String()),
				),
			},
		},
	})
}
