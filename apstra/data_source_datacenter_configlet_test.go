package tfapstra_test

import (
	"context"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	dataSourceDatacenterConfigletTemplateByNameHCL = `
	data "apstra_datacenter_configlet" "test" {
    	blueprint_id = "%s"
		name = "%s"
	}
	`

	dataSourceDatacenterConfigletTemplateByIdHCL = `
	data "apstra_datacenter_configlet" "test" {
  		blueprint_id = "%s"
		id = "%s"
	}
	`
)

func TestAccDataSourceDatacenterConfiglet(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)

	// Set up a Catalog Property Set
	catalogConfigletId, configletData := testutils.CatalogConfigletA(t, ctx, client)

	// BlueprintA returns a bpClient and the template from which the blueprint was created
	bpClient := testutils.MakeOrFindBlueprint(t, ctx, "BPA", testutils.BlueprintA)

	configletId := testutils.BlueprintConfigletA(t, ctx, bpClient, catalogConfigletId, condition)

	resource.Test(t, resource.TestCase{
		// PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDatacenterConfigletTemplateByIdHCL,
					string(bpClient.Id()), string(configletId)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "id", configletId.String()),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "name", configletData.DisplayName),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "condition", condition),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.template_text", configletData.Generators[0].TemplateText),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.config_style", configletData.Generators[0].ConfigStyle.String()),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.section", utils.StringersToFriendlyString(configletData.Generators[0].Section, configletData.Generators[0].ConfigStyle)),
				),
			},
			// Read by Name
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDatacenterConfigletTemplateByNameHCL,
					string(bpClient.Id()), configletData.DisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "id", configletId.String()),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "name", configletData.DisplayName),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "condition", condition),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.template_text", configletData.Generators[0].TemplateText),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.config_style", configletData.Generators[0].ConfigStyle.String()),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.section", utils.StringersToFriendlyString(configletData.Generators[0].Section, configletData.Generators[0].ConfigStyle)),
				),
			},
		},
	})
}
