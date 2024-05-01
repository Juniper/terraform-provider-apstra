//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"math"
	"net"
	"strconv"
	"strings"
	"testing"

	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceIpv4PoolTemplateHCL = `
resource %q %q {
  name = "%s"
  subnets = [%s
  ]
}
`
	resourceIpv4PoolSubnetTemplateHCL = "\n    {network = %q},"
)

func ipv4Subnets(t testing.TB, block string, size, count int) []string {
	base, n, err := net.ParseCIDR(block)
	require.NoError(t, err)
	require.Equalf(t, base.String(), n.IP.String(), "%s is not a base network address", block)

	ones, bits := n.Mask.Size()
	a := math.Pow(2, float64(bits-ones))
	b := math.Pow(2, float64(bits-size)) * float64(count)
	require.Greaterf(t, a, b, "%d %d-bit subnets won't fit in %s", count, size, block)

	result := make([]string, count)
	for i := 0; i < count; i++ {
		subnet, err := cidr.Subnet(n, size-ones, i)
		require.NoError(t, err)
		result[i] = subnet.String()
	}

	return result
}

type ipv4PoolConfig struct {
	name    string
	subnets []string
}

func (o ipv4PoolConfig) render(rType, rName string) string {
	sb := new(strings.Builder)
	for _, subnet := range o.subnets {
		sb.WriteString(fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, subnet))
	}

	return fmt.Sprintf(resourceIpv4PoolTemplateHCL, rType, rName,
		o.name,
		sb.String(),
	)
}

func (o ipv4PoolConfig) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "subnets.#", strconv.Itoa(len(o.subnets)))

	for i, s := range o.subnets {
		// todo: add tests for each subnet
		_, _ = i, s
	}

	return result
}

func TestAccResourceIpv4Pool(t *testing.T) {
	ctx := context.Background()
	testutils.TestCfgFileToEnv()

	type testCase struct {
		stepConfigs []ipv4PoolConfig
	}

	testCases := map[string]testCase{
		"simple": {
			stepConfigs: []ipv4PoolConfig{
				{
					name:    acctest.RandString(6),
					subnets: []string{"10.0.0.0/24", "10.0.1.0/24"},
				},
				{
					name:    acctest.RandString(6),
					subnets: []string{"10.1.0.0/24", "10.1.1.0/24", "10.1.2.0/24"},
				},
			},
		},
		"lots": {
			stepConfigs: []ipv4PoolConfig{
				{
					name:    acctest.RandString(6),
					subnets: ipv4Subnets(t, "10.0.0.0/8", 28, 50),
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceIpv4Pool)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			steps := make([]resource.TestStep, len(tCase.stepConfigs))
			for i, stepConfig := range tCase.stepConfigs {
				config := stepConfig.render(resourceType, tName)
				checks := stepConfig.testChecks(t, resourceType, tName)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}

	//var (
	//	testAccResourceIpv4PoolCfg1Name    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	//	testAccResourceIpv4PoolCfg1Subnets = strings.Join([]string{
	//		fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, "192.168.0.0/16"),
	//	}, ",")
	//	testAccResourceIpv4PoolCfg1 = fmt.Sprintf(resourceIpv4PoolTemplateHCL, testAccResourceIpv4PoolCfg1Name, testAccResourceIpv4PoolCfg1Subnets)
	//
	//	testAccResourceIpv4PoolCfg2Name    = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	//	testAccResourceIpv4PoolCfg2Subnets = strings.Join([]string{
	//		fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, "192.168.1.0/24"),
	//		fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, "192.168.0.0/24"),
	//		fmt.Sprintf(resourceIpv4PoolSubnetTemplateHCL, "192.168.2.0/23"),
	//	}, ",")
	//	testAccResourceIpv4PoolCfg2 = fmt.Sprintf(resourceIpv4PoolTemplateHCL, testAccResourceIpv4PoolCfg2Name, testAccResourceIpv4PoolCfg2Subnets)
	//)
	//
	//resource.Test(t, resource.TestCase{
	//	ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
	//	Steps: []resource.TestStep{
	//		// Create and Read testing
	//		{
	//			Config: insecureProviderConfigHCL + testAccResourceIpv4PoolCfg1,
	//			Check: resource.ComposeAggregateTestCheckFunc(
	//				// Verify ID has any value set
	//				resource.TestCheckResourceAttrSet("apstra_ipv4_pool.test", "id"),
	//				// Verify name and overall usage statistics
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "name", testAccResourceIpv4PoolCfg1Name),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "status", "not_in_use"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "total", "65536"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "used", "0"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "used_percentage", "0"),
	//				// Verify number of subnets
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.#", "1"),
	//				// Verify first subnet
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.network", "192.168.0.0/16"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.status", "pool_element_available"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.total", "65536"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.used", "0"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.used_percentage", "0"),
	//			),
	//		},
	//		// Update and Read testing
	//		{
	//			Config: insecureProviderConfigHCL + testAccResourceIpv4PoolCfg2,
	//			Check: resource.ComposeAggregateTestCheckFunc(
	//				// Verify ID has any value set
	//				resource.TestCheckResourceAttrSet("apstra_ipv4_pool.test", "id"),
	//				// Verify name and overall usage statistics
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "name", testAccResourceIpv4PoolCfg2Name),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "status", "not_in_use"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "total", "1024"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "used", "0"),
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "used_percentage", "0"),
	//				// Verify number of subnets
	//				resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.#", "3"),
	//				//// cannot verify subnets here because they're not sorted
	//				//resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.0.network", "192.168.0.0/24"),
	//				//resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.1.network", "192.168.1.0/24"),
	//				//resource.TestCheckResourceAttr("apstra_ipv4_pool.test", "subnets.2.network", "192.168.2.0/23"),
	//			),
	//		},
	//	},
	//})
}
