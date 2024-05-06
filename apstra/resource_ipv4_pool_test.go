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
	t.Helper()
	result := newTestChecks(rType + "." + rName)

	var totalIPs int
	prefixes := make([]net.IPNet, len(o.subnets))
	for i, s := range o.subnets {
		ip, prefix, err := net.ParseCIDR(s)
		require.NoError(t, err)
		require.EqualValuesf(t, ip.String(), prefix.IP.String(), "%s is not a base address", s)

		prefixes[i] = *prefix

		ones, bits := prefix.Mask.Size()
		totalIPs += int(math.Pow(2, float64(bits-ones)))
	}

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "status", "not_in_use")
	result.append(t, "TestCheckResourceAttr", "total", strconv.Itoa(totalIPs))
	result.append(t, "TestCheckResourceAttr", "used", "0")
	result.append(t, "TestCheckResourceAttr", "used_percentage", "0")
	result.append(t, "TestCheckResourceAttr", "subnets.#", strconv.Itoa(len(o.subnets)))

	for _, p := range prefixes {
		ones, bits := p.Mask.Size()
		v := map[string]string{
			"network":         p.String(),
			"status":          "pool_element_available",
			"total":           strconv.Itoa(int(math.Pow(2, float64(bits-ones)))),
			"used":            "0",
			"used_percentage": "0",
		}
		result.appendSetNestedCheck(t, "subnets.*", v)
	}

	return result
}

func TestAccResourceIpv4Pool(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, testutils.TestCfgFileToEnv())

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
		// AOS-46273
		//"lots": {
		//	stepConfigs: []ipv4PoolConfig{
		//		{
		//			name:    acctest.RandString(6),
		//			subnets: ipv4Subnets(t, "10.0.0.0/8", 28, 50),
		//		},
		//	},
		//},
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
}
