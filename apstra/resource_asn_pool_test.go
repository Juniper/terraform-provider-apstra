package tfapstra

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceAsnPoolBasicCfg = `
// provider config
%s

// resource config
resource "apstra_asn_pool" "test" {
  name = "%s"
  ranges = [%s]
}
`
	resourceAsnPoolRange = "{first = %d, last = %d}"
)

func TestAccResourceAsnPool_basic(t *testing.T) {
	name1 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	ranges1 := fmt.Sprintf(resourceAsnPoolRange, 10, 20)

	name2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	ranges2 := strings.Join([]string{
		fmt.Sprintf(resourceAsnPoolRange, 1, 3),
		fmt.Sprintf(resourceAsnPoolRange, 5, 11),
		fmt.Sprintf(resourceAsnPoolRange, 15, 25),
	}, ",")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(resourceAsnPoolBasicCfg, insecureProviderConfigHCL, name1, ranges1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_asn_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "name", name1),
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
				Config: fmt.Sprintf(resourceAsnPoolBasicCfg, insecureProviderConfigHCL, name2, ranges2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_asn_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_asn_pool.test", "name", name2),
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
