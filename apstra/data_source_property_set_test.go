package tfapstra

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
	dataSourcePropertySetTemplateByNameHCL = `
data "apstra_property_set" "test" {
  name = "%s"
}
`

	dataSourcePropertySetTemplateByIdHCL = `
data "apstra_property_set" "test" {
  id = "%s"
}
`
)

func TestAccDataSourcePropertySet(t *testing.T) {
	ctx := context.Background()

	ps := testutils.PropertySetA(t, ctx)

	TestPSData := func(state string) error {
		var diags diag.Diagnostics
		if !utils.JSONEqual(types.StringValue(state), types.StringValue(string(ps.Data.Values)), &diags) {
			return fmt.Errorf("input Data does not match output Input %v. Output %v", ps.Data, state)
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
		//PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourcePropertySetTemplateByIdHCL, string(ps.Id)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_property_set.test", "id", ps.Id.String()),
					resource.TestCheckResourceAttr("data.apstra_property_set.test", "name", ps.Data.Label),
					resource.TestCheckResourceAttrWith("data.apstra_property_set.test", "data", TestPSData),
				),
			},
			// Read by Name
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourcePropertySetTemplateByNameHCL, ps.Data.Label),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_property_set.test", "id", ps.Id.String()),
					resource.TestCheckResourceAttr("data.apstra_property_set.test", "name", ps.Data.Label),
					resource.TestCheckResourceAttrWith("data.apstra_property_set.test", "data", TestPSData),
				),
			},
		},
	})
}
