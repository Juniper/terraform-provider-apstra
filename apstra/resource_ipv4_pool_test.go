package tfapstra

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceIpv4PoolTemplateHCL = `
// resource config
resource "apstra_ipv4_pool" "test" {
  name = "%s"
  subnets = [%s]
}
`
	resourceIpv4PoolSubnetTemplateHCL = "{network = %q}"
)

func TestAccResourceIpv4Pool_basic(t *testing.T) {
	var (
		testAccResourceIpv4PoolCfg1Name    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourceIpv4PoolCfg1Subnets = strings.Join([]string{
			fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, "192.168.0.0/16"),
		}, ",")
		testAccResourceIpv4PoolCfg1 = fmt.Sprintf(resourceIpv4PoolTemplateHCL, testAccResourceIpv4PoolCfg1Name, testAccResourceIpv4PoolCfg1Subnets)

		testAccResourceIpv4PoolCfg2Name    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourceIpv4PoolCfg2Subnets = strings.Join([]string{
			fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, "192.168.1.0/24"),
			fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, "192.168.0.0/24"),
			fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, "192.168.2.0/23"),
		}, ",")
		testAccResourceIpv4PoolCfg2 = fmt.Sprintf(resourceIpv4PoolTemplateHCL, testAccResourceIpv4PoolCfg2Name, testAccResourceIpv4PoolCfg2Subnets)
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceIpv4PoolCfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_ipv4_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "name", testAccResourceIpv4PoolCfg1Name),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "total", "65536"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "used_percentage", "0"),
					// Verify number of subnets
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.#", "1"),
					// Verify first subnet
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.network", "192.168.0.0/16"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.status", "pool_element_available"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.total", "65536"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.used", "0"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.used_percentage", "0"),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceIpv4PoolCfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_ipv4_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "name", testAccResourceIpv4PoolCfg2Name),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "total", "1024"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "used_percentage", "0"),
					// Verify number of subnets
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.#", "3"),
					//// cannot verify subnets here because they're not sorted
					//resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.network", "192.168.0.0/24"),
					//resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.1.network", "192.168.1.0/24"),
					//resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.2.network", "192.168.2.0/23"),
				),
			},
		},
	})
}
