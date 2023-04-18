package tfapstra

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceIpv6PoolTemplateHCL = `
// resource config
resource "apstra_ipv6_pool" "test" {
  name = "%s"
  subnets = [%s]
}
`
	resourceIpv6PoolSubnetTemplateHCL = "{network = %q}"
)

var (
	testAccResourceIpv6PoolCfg1Name    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	testAccResourceIpv6PoolCfg1Subnets = strings.Join([]string{
		fmt.Sprintf(resourceIpv6PoolSubnetTemplateHCL, "2001:db8::/66"),
	}, ",")
	testAccResourceIpv6PoolCfg1 = fmt.Sprintf(resourceIpv6PoolTemplateHCL, testAccResourceIpv6PoolCfg1Name, testAccResourceIpv6PoolCfg1Subnets)

	testAccResourceIpv6PoolCfg2Name    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	testAccResourceIpv6PoolCfg2Subnets = strings.Join([]string{
		fmt.Sprintf(resourceIpv6PoolSubnetTemplateHCL, "2001:db8:1::/96"),
		fmt.Sprintf(resourceIpv6PoolSubnetTemplateHCL, "2001:db8:2::/96"),
		fmt.Sprintf(resourceIpv6PoolSubnetTemplateHCL, "2001:db8:3::/96"),
	}, ",")
	testAccResourceIpv6PoolCfg2 = fmt.Sprintf(resourceIpv6PoolTemplateHCL, testAccResourceIpv6PoolCfg2Name, testAccResourceIpv6PoolCfg2Subnets)
)

func TestAccResourceIpv6Pool_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceIpv6PoolCfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_ipv6_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "name", testAccResourceIpv6PoolCfg1Name),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "total", "4611686018427387904"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "used_percentage", "0"),
					// Verify number of subnets
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.#", "1"),
					// Verify first subnet
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.0.network", "2001:db8::/66"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.0.status", "pool_element_available"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.0.total", "4611686018427387904"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.0.used", "0"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.0.used_percentage", "0"),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceIpv6PoolCfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_ipv6_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "name", testAccResourceIpv6PoolCfg2Name),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "total", "12884901888"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "used_percentage", "0"),
					// Verify number of subnets
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.#", "3"),
					//// cannot verify subnets here because they're not sorted
					//resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.0.network", "2001:db8:1::/96"),
					//resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.1.network", "2001:db8:2::/96"),
					//resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "subnets.2.network", "2001:db8:3::/96"),
				),
			},
		},
	})
}
