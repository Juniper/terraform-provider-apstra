package tfapstra

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceVniPoolTemplateHCL = `
// resource config
resource "apstra_vni_pool" "test" {
  name = "%s"
  ranges = [%s]
}
`
	resourceVniPoolRangeTemplateHCL = "{first = %d, last = %d}"
)

func TestAccResourceVniPool_basic(t *testing.T) {
	var (
		testAccResourceVniPoolCfg1Name   = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourceVniPoolCfg1Ranges = strings.Join([]string{
			fmt.Sprintf(resourceVniPoolRangeTemplateHCL, 10010, 10020),
		}, ",")
		testAccResourceVniPoolCfg1 = fmt.Sprintf(resourceVniPoolTemplateHCL, testAccResourceVniPoolCfg1Name, testAccResourceVniPoolCfg1Ranges)

		testAccResourceVniPoolCfg2Name   = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourceVniPoolCfg2Ranges = strings.Join([]string{
			fmt.Sprintf(resourceVniPoolRangeTemplateHCL, 10001, 10003),
			fmt.Sprintf(resourceVniPoolRangeTemplateHCL, 10005, 10011),
			fmt.Sprintf(resourceVniPoolRangeTemplateHCL, 10015, 10025),
		}, ",")
		testAccResourceVniPoolCfg2 = fmt.Sprintf(resourceVniPoolTemplateHCL, testAccResourceVniPoolCfg2Name, testAccResourceVniPoolCfg2Ranges)
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceVniPoolCfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_vni_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "name", testAccResourceVniPoolCfg1Name),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "total", "11"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "used_percentage", "0"),
					// Verify number of ranges
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.#", "1"),
					// Verify first range
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.first", "10010"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.last", "10020"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.status", "pool_element_available"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.total", "11"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.used", "0"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.used_percentage", "0"),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceVniPoolCfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_vni_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "name", testAccResourceVniPoolCfg2Name),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "total", "21"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "used_percentage", "0"),
					// Verify number of ranges
					resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.#", "3"),
					//// cannot verify ranges here because they're not sorted
					//resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.first", "10001"),
					//resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.last", "10003"),
					//resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.status", "pool_element_available"),
					//resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.total", "3"),
					//resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.used", "0"),
					//resource.TestCheckResourceAttr("apstra_vni_pool.test", "ranges.0.used_percentage", "0"),
				),
			},
		},
	})
}
