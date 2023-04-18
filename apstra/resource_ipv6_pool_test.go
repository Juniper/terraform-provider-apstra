package tfapstra

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceIpv6PoolBasicCfg = `
// provider config
%s

// resource config
resource "apstra_ipv6_pool" "test" {
  name = "%s"
  subnets = [%s]
}
`
	resourceIpv6PoolSubnet = "{network = %q}"
)

func TestAccResourceIpv6Pool_basic(t *testing.T) {
	name1 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	subnets1 := fmt.Sprintf(resourceIpv6PoolSubnet, "2001:db8::/66")

	name2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	subnets2 := strings.Join([]string{
		fmt.Sprintf(resourceIpv6PoolSubnet, "2001:db8:1::/96"),
		fmt.Sprintf(resourceIpv6PoolSubnet, "2001:db8:2::/96"),
		fmt.Sprintf(resourceIpv6PoolSubnet, "2001:db8:3::/96"),
	}, ",")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(resourceIpv6PoolBasicCfg, insecureProviderConfigHCL, name1, subnets1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_ipv6_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "name", name1),
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
				Config: fmt.Sprintf(resourceIpv6PoolBasicCfg, insecureProviderConfigHCL, name2, subnets2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_ipv6_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_ipv6_pool.test", "name", name2),
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
