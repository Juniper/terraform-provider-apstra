package tfapstra

import (
	"context"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	ps, deleteFunc, err := testutils.PropertySetA(ctx)
	if err != nil {
		t.Error(err)
		t.Fatal(deleteFunc(ctx))
	}
	defer func() {
		err := deleteFunc(ctx)
		if err != nil {
			t.Error(err)
		}
	}()
	d := diag.Diagnostics{}
	TestPSData := func(state string) error {
		if !utils.JSONEqual(types.StringValue(state), types.StringValue(string(ps.Data.Values)), &d) {
			return fmt.Errorf("input Data does not match output Input %v. Output %v", ps.Data, state)
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
