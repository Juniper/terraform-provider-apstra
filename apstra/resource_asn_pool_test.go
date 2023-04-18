package tfapstra

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceAsnPoolTemplateHCL = `
// resource config
resource "apstra_asn_pool" "test" {
  name = "%s"
  ranges = [%s]
}
`
	resourceAsnPoolRangeTemplateHCL = "{first = %d, last = %d}"
)

var (
	testAccResourceAsnPoolCfg1Name   = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	testAccResourceAsnPoolCfg1Ranges = strings.Join([]string{
		fmt.Sprintf(resourceAsnPoolRangeTemplateHCL, 10, 20),
	}, ",")
	testAccResourceAsnPoolCfg1 = fmt.Sprintf(resourceAsnPoolTemplateHCL, testAccResourceAsnPoolCfg1Name, testAccResourceAsnPoolCfg1Ranges)

	testAccResourceAsnPoolCfg2Name   = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	testAccResourceAsnPoolCfg2Ranges = strings.Join([]string{
		fmt.Sprintf(resourceAsnPoolRangeTemplateHCL, 1, 3),
		fmt.Sprintf(resourceAsnPoolRangeTemplateHCL, 5, 11),
		fmt.Sprintf(resourceAsnPoolRangeTemplateHCL, 15, 25),
	}, ",")
	testAccResourceAsnPoolCfg2 = fmt.Sprintf(resourceAsnPoolTemplateHCL, testAccResourceAsnPoolCfg2Name, testAccResourceAsnPoolCfg2Ranges)
)

func TestAccResourceAsnPool_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceAsnPoolCfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_asn_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "name", testAccResourceAsnPoolCfg1Name),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "total", "11"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "used_percentage", "0"),
					// Verify number of ranges
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.#", "1"),
					// Verify first range
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.first", "10"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.last", "20"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.status", "pool_element_available"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.total", "11"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.used", "0"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.used_percentage", "0"),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceAsnPoolCfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_asn_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "name", testAccResourceAsnPoolCfg2Name),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "total", "21"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "used_percentage", "0"),
					// Verify number of ranges
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.#", "3"),
					//// cannot verify ranges here because they're not sorted
					//resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.first", "1"),
					//resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.last", "3"),
					//resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.status", "pool_element_available"),
					//resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.total", "3"),
					//resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.used", "0"),
					//resource.TestCheckResourceAttr("apstra_asn_pool.test", "ranges.0.used_percentage", "0"),
				),
			},
		},
	})
}
