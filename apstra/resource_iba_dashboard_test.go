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
	resourceIbaDashboardTemplateHCL = `

resource "apstra_iba_dashboard" "a" {
  blueprint_id = "%s"
  default = false
  description = "The dashboard presents the data of utilization of system cpu, system memory and maximum disk utilization of a partition on every system present."
  name = "Test Dashboard"
  widget_grid = tolist([
  %s
  ])
}
`
	one_pane = `tolist([
      	"%s"
    ])`

	two_panes = `
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

	bpClient, bpDelete, err := testutils.MakeOrFindBlueprint(ctx, "1bb57", testutils.BlueprintA)

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
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceIbaDashboardTemplateHCL, bpClient.Id(), fmt.Sprintf(one_pane, widgetIdA)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_iba_dashboard.a", "id"),
					resource.TestCheckResourceAttr("apstra_iba_dashboard.a", "widget_grid.0.0", widgetIdA.String()),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceIbaDashboardTemplateHCL, bpClient.Id(), fmt.Sprintf(two_panes, widgetIdA, widgetIdB)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_iba_dashboard.a", "id"),
					resource.TestCheckResourceAttr("apstra_iba_dashboard.a", "widget_grid.0.0", widgetIdA.String()),
					resource.TestCheckResourceAttr("apstra_iba_dashboard.a", "widget_grid.1.0", widgetIdB.String()),
				),
			},
		},
	})
}
