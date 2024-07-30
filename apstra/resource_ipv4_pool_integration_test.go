//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	dataIpv4PoolByNameHCL = `
data %q "%s_by_name" {
  name = %s
}
`
	dataIpv4PoolByIdHCL = `
data %q "%s_by_id" {
  id = %s
}
`
	resourceIpv4PoolHCL = `
resource %q %q {
  name = "%s"
  subnets = [%s
  ]
}
`
	resourceIpv4PoolRangeHCL = "\n    { network = %q },"
)

type resourceIpv4Pool struct {
	name    string
	subnets []net.IPNet
}

func (o resourceIpv4Pool) render(rType, rName string) string {
	ipv4Ranges := make([]string, len(o.subnets))
	for i, ipv4Range := range o.subnets {
		ipv4Ranges[i] = fmt.Sprintf(resourceIpv4PoolRangeHCL, ipv4Range.String())
	}

	return "" +
		fmt.Sprintf(resourceIpv4PoolHCL,
			rType, rName,
			o.name,
			strings.Join(ipv4Ranges, ""),
		) +
		fmt.Sprintf(dataIpv4PoolByIdHCL, rType, rName, fmt.Sprintf("%s.%s.id", rType, rName)) +
		fmt.Sprintf(dataIpv4PoolByNameHCL, rType, rName, fmt.Sprintf("%s.%s.name", rType, rName))
}

func (o resourceIpv4Pool) testChecks(t testing.TB, rType, rName string) testChecks {
	checks := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	checks.append(t, "TestCheckResourceAttrSet", "id")
	checks.append(t, "TestCheckResourceAttr", "name", o.name)
	checks.append(t, "TestCheckNoResourceAttr", "total")
	checks.append(t, "TestCheckNoResourceAttr", "status")
	checks.append(t, "TestCheckNoResourceAttr", "used")
	checks.append(t, "TestCheckNoResourceAttr", "used_percentage")

	checks.append(t, "TestCheckResourceAttr", "subnets.#", strconv.Itoa(len(o.subnets)))

	for _, subnet := range o.subnets {
		checks.appendSetNestedCheck(t, "subnets.*", map[string]string{
			"network": subnet.String(),
		})
	}

	// -----------------------------
	// DATA SOURCE "by_id" checks below here
	// -----------------------------
	checks.setPath("data." + rType + "." + rName + "_by_id")
	var total int
	for _, subnet := range o.subnets {
		ones, _ := subnet.Mask.Size()
		thisSubnetTotal := 1 << (32 - ones)

		checks.appendSetNestedCheck(t, "subnets.*", map[string]string{
			"network":         subnet.String(),
			"total":           strconv.Itoa(thisSubnetTotal),
			"status":          "pool_element_available",
			"used":            "0",
			"used_percentage": "0",
		})

		total += thisSubnetTotal
	}

	checks.append(t, "TestCheckResourceAttrSet", "id")
	checks.append(t, "TestCheckResourceAttr", "name", o.name)
	checks.append(t, "TestCheckResourceAttr", "total", strconv.Itoa(total))
	checks.append(t, "TestCheckResourceAttr", "status", "not_in_use")
	checks.append(t, "TestCheckResourceAttr", "used", "0")
	checks.append(t, "TestCheckResourceAttr", "used_percentage", "0")

	// -----------------------------
	// DATA SOURCE "by_name" checks below here
	// -----------------------------
	checks.setPath("data." + rType + "." + rName + "_by_name")
	for _, subnet := range o.subnets {
		ones, _ := subnet.Mask.Size()
		thisSubnetTotal := 1 << (32 - ones)

		checks.appendSetNestedCheck(t, "subnets.*", map[string]string{
			"network":         subnet.String(),
			"total":           strconv.Itoa(thisSubnetTotal),
			"status":          "pool_element_available",
			"used":            "0",
			"used_percentage": "0",
		})
	}

	checks.append(t, "TestCheckResourceAttrSet", "id")
	checks.append(t, "TestCheckResourceAttr", "name", o.name)
	checks.append(t, "TestCheckResourceAttr", "total", strconv.Itoa(total))
	checks.append(t, "TestCheckResourceAttr", "status", "not_in_use")
	checks.append(t, "TestCheckResourceAttr", "used", "0")
	checks.append(t, "TestCheckResourceAttr", "used_percentage", "0")

	return checks
}

func TestAccResourceIpv4Pool(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	type testStep struct {
		config resourceIpv4Pool
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"simple_case": {
			steps: []testStep{
				{
					config: resourceIpv4Pool{
						name: acctest.RandString(6),
						subnets: []net.IPNet{
							randomPrefix(t, "10.0.0.0/16", 24),
							randomPrefix(t, "10.1.0.0/16", 25),
							randomPrefix(t, "10.2.0.0/16", 26),
							randomPrefix(t, "10.3.0.0/16", 27),
						},
					},
				},
				{
					config: resourceIpv4Pool{
						name: acctest.RandString(6),
						subnets: []net.IPNet{
							randomPrefix(t, "10.4.0.0/16", 24),
							randomPrefix(t, "10.5.0.0/16", 25),
						},
					},
				},
				{
					config: resourceIpv4Pool{
						name: acctest.RandString(6),
						subnets: []net.IPNet{
							randomPrefix(t, "10.6.0.0/16", 24),
							randomPrefix(t, "10.7.0.0/16", 25),
							randomPrefix(t, "10.8.0.0/16", 26),
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceIpv4Pool)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.apiVersionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------%s// -------- end config for %s ------\n\n", stepName, config, stepName)
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
