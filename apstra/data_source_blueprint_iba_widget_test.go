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
	widgetIdA, widgetDataA, _, _, cleanup := testutils.TestWidgetsAB(ctx, t, bpClient)
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
