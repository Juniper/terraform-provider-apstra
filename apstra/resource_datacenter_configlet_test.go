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
	resourceDatacenterConfigletTemplateByIdHCL = `
  	resource "apstra_datacenter_configlet" "test" {
  		blueprint_id = "%s"
  		catalog_configlet_id = "%s"
  		condition = %q
  	}
	`
	resourceDatacenterConfigletTemplateByDataHCL = `
	data "apstra_configlet" "cat_cfg" {
   		name = "CatalogConfigletA"
  	}

  	resource "apstra_datacenter_configlet" "test" {
  		blueprint_id = "%s"
  		condition = %q
  		name = data.apstra_configlet.cat_cfg.name
  		generators = data.apstra_configlet.cat_cfg.generators
  	}
	`

	condition = "role in [\"spine\"]"
)

func TestAccResourceDatacenterConfiglet(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)

	// Set up a Catalog Property Set
	cc, data := testutils.CatalogConfigletA(t, ctx, client)

	// BlueprintA returns a bpClient and the template from which the blueprint was created
	bpClient := testutils.MakeOrFindBlueprint(t, ctx, "BPA", testutils.BlueprintA)

	resource.Test(t, resource.TestCase{
		// PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceDatacenterConfigletTemplateByIdHCL,
					bpClient.Id().String(), cc.String(), condition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("apstra_datacenter_configlet.test", "id"),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "name", data.DisplayName),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "condition", condition),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "generators.0.template_text", data.Generators[0].TemplateText),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "generators.0.config_style", data.Generators[0].ConfigStyle.String()),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "generators.0.section", utils.StringersToFriendlyString(data.Generators[0].Section, data.Generators[0].ConfigStyle)),
				),
			},
			// Create by Data
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceDatacenterConfigletTemplateByDataHCL,
					bpClient.Id().String(), condition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("apstra_datacenter_configlet.test", "id"),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "name", data.DisplayName),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "condition", condition),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "generators.0.template_text", data.Generators[0].TemplateText),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "generators.0.config_style", data.Generators[0].ConfigStyle.String()),
					resource.TestCheckResourceAttr("apstra_datacenter_configlet.test", "generators.0.section", utils.StringersToFriendlyString(data.Generators[0].Section, data.Generators[0].ConfigStyle)),
				),
			},
		},
	})
}
