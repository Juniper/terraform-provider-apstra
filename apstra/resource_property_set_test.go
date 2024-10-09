package tfapstra

import (
	"fmt"
	"testing"

	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourcePropertySetTemplateHCL = `
// resource config
resource "apstra_property_set" "test" {
  name = "%s"
  data = jsonencode(%s)
}
`

	data2_string = `{ 
					"value_str":"str",
					"value_int":42
				}`
	data1_string = `{ 
					"value_str":"str",
					"value_int":42,
					"value_json" :{
						"inner_value_str":"innerstr",
						"inner_value_int":4242
					}	
				}`
)

func TestAccResourcePropertySet(t *testing.T) {
	testutils.TestCfgFileToEnv(t)

	var (
		testAccResourcePropertySet1Name = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourcePropertySetCfg1  = fmt.Sprintf(resourcePropertySetTemplateHCL, testAccResourcePropertySet1Name, data1_string)
		testAccResourcePropertySet2Name = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourcePropertySetCfg2  = fmt.Sprintf(resourcePropertySetTemplateHCL, testAccResourcePropertySet2Name, data2_string)
	)

	d := diag.Diagnostics{}
	TestPSData1 := func(state string) error {
		if !utils.JSONEqual(types.StringValue(state), types.StringValue(data1_string), &d) {
			return fmt.Errorf("input Data does not match output Input %v. Output %v", data1_string, state)
		}
		return nil
	}

	TestPSData2 := func(state string) error {
		if !utils.JSONEqual(types.StringValue(state), types.StringValue(data2_string), &d) {
			return fmt.Errorf("input Data does not match output Input %v. Output %v", data2_string, state)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourcePropertySetCfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_property_set.test", "id"),
					// Verify name and data
					resource.TestCheckResourceAttr("apstra_property_set.test", "name", testAccResourcePropertySet1Name),
					resource.TestCheckResourceAttrWith("apstra_property_set.test", "data", TestPSData1),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourcePropertySetCfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_property_set.test", "id"),
					// Verify name and data
					resource.TestCheckResourceAttr("apstra_property_set.test", "name", testAccResourcePropertySet2Name),
					resource.TestCheckResourceAttrWith("apstra_property_set.test", "data", TestPSData2),
				),
			},
		},
	})
}
