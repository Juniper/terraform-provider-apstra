package tfapstra

import (
	"context"
	"errors"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	resourceBlueprintIbaDashboardTemplateHCL = `

resource "apstra_blueprint_iba_dashboard" "a" {
  blueprint_id = "%s"
  default = false
  %s
  name = "Test Dashboard"
  widget_grid = tolist([
  %s
  ])
}
`
	descString = `description = "The dashboard presents the data of utilization of system cpu,system memory and maximum disk utilization of a partition on every system present."`
	onePane    = `tolist([
      	"%s"
    ])`

	twoPanes = `
    tolist([
    	"%s"    
	]),
    tolist([
    	"%s"
	])
`
)

func TestAccResourceDashboard(t *testing.T) {

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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing. Empty Description Test
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceBlueprintIbaDashboardTemplateHCL,
					bpClient.Id(), "", fmt.Sprintf(onePane, widgetIdA)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any133 value set
					resource.TestCheckResourceAttrSet("apstra_blueprint_iba_dashboard.a", "id"),
					resource.TestCheckResourceAttr("apstra_blueprint_iba_dashboard.a", "widget_grid.0.0", widgetIdA.String()),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceBlueprintIbaDashboardTemplateHCL,
					bpClient.Id(), descString, fmt.Sprintf(twoPanes, widgetIdA, widgetIdB)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_blueprint_iba_dashboard.a", "id"),
					resource.TestCheckResourceAttr("apstra_blueprint_iba_dashboard.a", "widget_grid.0.0", widgetIdA.String()),
					resource.TestCheckResourceAttr("apstra_blueprint_iba_dashboard.a", "widget_grid.1.0", widgetIdB.String()),
				),
			},
		},
	})
}
