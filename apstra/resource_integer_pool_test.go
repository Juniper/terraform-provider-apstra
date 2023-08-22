package tfapstra

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceIntegerPoolTemplateHCL = `
// resource config
resource "apstra_integer_pool" "test" {
  name = "%s"
  ranges = [%s]
}
`
	resourceIntegerPoolRangeTemplateHCL = "{first = %d, last = %d}"
)

func TestAccResourceIntegerPool_basic(t *testing.T) {
	var (
		testAccResourceIntegerPoolCfg1Name   = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourceIntegerPoolCfg1Ranges = strings.Join([]string{
			fmt.Sprintf(resourceIntegerPoolRangeTemplateHCL, 10010, 10020),
		}, ",")
		testAccResourceIntegerPoolCfg1 = fmt.Sprintf(resourceIntegerPoolTemplateHCL, testAccResourceIntegerPoolCfg1Name, testAccResourceIntegerPoolCfg1Ranges)

		testAccResourceIntegerPoolCfg2Name   = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourceIntegerPoolCfg2Ranges = strings.Join([]string{
			fmt.Sprintf(resourceIntegerPoolRangeTemplateHCL, 10001, 10003),
			fmt.Sprintf(resourceIntegerPoolRangeTemplateHCL, 10005, 10011),
			fmt.Sprintf(resourceIntegerPoolRangeTemplateHCL, 10015, 10025),
		}, ",")
		testAccResourceIntegerPoolCfg2 = fmt.Sprintf(resourceIntegerPoolTemplateHCL, testAccResourceIntegerPoolCfg2Name, testAccResourceIntegerPoolCfg2Ranges)
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceIntegerPoolCfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_integer_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "name", testAccResourceIntegerPoolCfg1Name),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "total", "11"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "used_percentage", "0"),
					// Verify number of ranges
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.#", "1"),
					// Verify first range
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.first", "10010"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.last", "10020"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.status", "pool_element_available"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.total", "11"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.used", "0"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.used_percentage", "0"),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceIntegerPoolCfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_integer_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "name", testAccResourceIntegerPoolCfg2Name),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "total", "21"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "used_percentage", "0"),
					// Verify number of ranges
					resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.#", "3"),
					//// cannot verify ranges here because they're not sorted
					//resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.first", "10001"),
					//resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.last", "10003"),
					//resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.status", "pool_element_available"),
					//resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.total", "3"),
					//resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.used", "0"),
					//resource.TestCheckResourceAttr("apstra_integer_pool.test", "ranges.0.used_percentage", "0"),
				),
			},
		},
	})
}
