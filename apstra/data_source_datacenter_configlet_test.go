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
	client, err := testutils.GetTestClient(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Set up a Catalog Property Set
	ccId, data, deleteFunc, err := testutils.CatalogConfigletA(ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := deleteFunc(ctx, ccId)
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

	bpcid, bpcfgletDelete, err := testutils.BlueprintConfigletA(ctx, bpClient, ccId, condition)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = bpcfgletDelete(ctx, bpcid)
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
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDatacenterConfigletTemplateByIdHCL,
					string(bpClient.Id()), string(bpcid)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "id", bpcid.String()),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "name", data.DisplayName),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "condition", condition),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.template_text", data.Generators[0].TemplateText),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.config_style", data.Generators[0].ConfigStyle.String()),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.section", utils.StringersToFriendlyString(data.Generators[0].Section, data.Generators[0].ConfigStyle)),
				),
			},
			// Read by Name
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDatacenterConfigletTemplateByNameHCL,
					string(bpClient.Id()), data.DisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "id", bpcid.String()),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "name", data.DisplayName),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "condition", condition),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.template_text", data.Generators[0].TemplateText),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.config_style", data.Generators[0].ConfigStyle.String()),
					resource.TestCheckResourceAttr("data.apstra_datacenter_configlet.test", "generators.0.section", utils.StringersToFriendlyString(data.Generators[0].Section, data.Generators[0].ConfigStyle)),
				),
			},
		},
	})
}
