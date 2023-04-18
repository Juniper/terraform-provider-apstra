package tfapstra

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceIpv4PoolBasicCfg = `
// provider config
%s

// resource config
resource "apstra_ipv4_pool" "test" {
  name = "%s"
  subnets = [%s]
}
`
	resourceIpv4PoolSubnet = "{network = %q}"
)

func TestAccResourceIpv4Pool_basic(t *testing.T) {
	name1 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	subnets1 := fmt.Sprintf(resourceIpv4PoolSubnet, "192.168.0.0/16")

	name2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	subnets2 := strings.Join([]string{
		fmt.Sprintf(resourceIpv4PoolSubnet, "192.168.1.0/24"),
		fmt.Sprintf(resourceIpv4PoolSubnet, "192.168.0.0/24"),
		fmt.Sprintf(resourceIpv4PoolSubnet, "192.168.2.0/23"),
	}, ",")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(resourceIpv4PoolBasicCfg, insecureProviderConfigHCL, name1, subnets1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_ipv4_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "name", name1),
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
				Config: fmt.Sprintf(resourceIpv4PoolBasicCfg, insecureProviderConfigHCL, name2, subnets2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_ipv4_pool.test", "id"),
					// Verify name and overall usage statistics
					resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "name", name2),
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
