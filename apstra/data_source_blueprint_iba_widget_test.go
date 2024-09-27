package tfapstra

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	dataSourceBlueprintIbaWidgetTemplateByNameHCL = `
	data "apstra_blueprint_iba_widget" "test" {
    	blueprint_id = "%s"
		name = "%s"
	}
	`

	dataSourceBlueprintIbaWidgetTemplateByIdHCL = `
	data "apstra_blueprint_iba_widget" "test" {
  		blueprint_id = "%s"
		id = "%s"
	}
	`
)

func TestAccDataSourceIbaWidget(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)
	clientVersion := version.Must(version.NewVersion(client.ApiVersion()))
	if !compatibility.BpIbaWidgetOk.Check(clientVersion) {
		t.Skipf("skipping due to version constraint %s", compatibility.BpIbaWidgetOk)
	}

	bpClient := testutils.MakeOrFindBlueprint(t, ctx, "BPA", testutils.BlueprintA)

	// Set up Widgets
	widgetIdA, widgetDataA, _, _ := testutils.TestWidgetsAB(t, ctx, bpClient)

	resource.Test(t, resource.TestCase{
		// PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceBlueprintIbaWidgetTemplateByIdHCL,
					string(bpClient.Id()), string(widgetIdA)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_widget.test", "id", widgetIdA.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_widget.test", "name",
						widgetDataA.Label),
				),
			},
			// Read by Name
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceBlueprintIbaWidgetTemplateByNameHCL,
					string(bpClient.Id()), widgetDataA.Label),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_widget.test", "id", widgetIdA.String()),
					resource.TestCheckResourceAttr("data.apstra_blueprint_iba_widget.test", "name",
						widgetDataA.Label),
				),
			},
		},
	})
}
