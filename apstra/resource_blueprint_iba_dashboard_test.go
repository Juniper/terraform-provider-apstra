package tfapstra

import (
	"context"
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

	bpClient := testutils.MakeOrFindBlueprint(t, ctx, "BPA", testutils.BlueprintA)

	// Set up Widgets
	widgetIdA, _, widgetIdB, _ := testutils.TestWidgetsAB(t, ctx, bpClient)

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
