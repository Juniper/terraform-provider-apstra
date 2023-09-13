package tfapstra

import (
	"context"
	"errors"
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
	client, err := testutils.GetTestClient(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Set up a Catalog Property Set
	cc, data, deleteFunc, err := testutils.CatalogConfigletA(ctx, client)
	if err != nil {
		t.Fatal(errors.Join(err, deleteFunc(ctx, cc)))
	}
	defer func() {
		err := deleteFunc(ctx, cc)
		if err != nil {
			t.Error(err)
		}
	}()

	// BlueprintA returns a bpClient and the template from which the blueprint was created
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

	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}

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
