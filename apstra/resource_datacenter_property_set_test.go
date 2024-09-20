package tfapstra_test

import (
	"context"
	"errors"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strings"
	"testing"
)

const (
	data_string = `
					{
						"value_str":"str",
						"value_int":42
					}`
	resourceDatacenterPropertySetTemplateHCL = `
  	resource "apstra_property_set" "test" {
  		name = "TEST_PS"
  		data = jsonencode(%s)
		}

	resource "apstra_datacenter_property_set" "test" {
  		blueprint_id = "%s"
        id = apstra_property_set.test.id
		sync_with_catalog = true
  	}
	`
)

func TestAccResourceDatacenterPropertySet(t *testing.T) {
	ctx := context.Background()

	// BlueprintA returns a bpClient and the template from which the blueprint was created
	bpClient := testutils.MakeOrFindBlueprint(t, ctx, "BPA", testutils.BlueprintA)

	TestPSData := func(state string) error {
		var diags diag.Diagnostics
		if !utils.JSONEqual(types.StringValue(state), types.StringValue(data_string), &diags) {
			return fmt.Errorf("input Data does not match output Input %v. Output %v", data_string, state)
		}
		if diags.HasError() {
			var sb strings.Builder
			for _, d := range diags.Errors() {
				sb.WriteString(d.Summary() + "\n")
				sb.WriteString(d.Detail() + "\n")
			}
			return errors.New(sb.String())
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		// PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Import the property set
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceDatacenterPropertySetTemplateHCL,
					data_string, bpClient.Id()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("apstra_datacenter_property_set.test", "id"),
					resource.TestCheckResourceAttr("apstra_datacenter_property_set.test", "name", "TEST_PS"),
					resource.TestCheckResourceAttrWith("apstra_datacenter_property_set.test", "data",
						TestPSData),
					resource.TestCheckResourceAttr("apstra_datacenter_property_set.test", "stale", "false"),
					resource.TestCheckResourceAttr("apstra_datacenter_property_set.test", "sync_required", "false"),
				),
			},
		},
	})
}
