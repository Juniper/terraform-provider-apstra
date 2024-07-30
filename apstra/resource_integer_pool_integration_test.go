//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
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
	dataIntegerPoolByNameHCL = `
data %q "%s_by_name" {
  name = %s
}
`
	dataIntegerPoolByIdHCL = `
data %q "%s_by_id" {
  id = %s
}
`
	resourceIntegerPoolHCL = `
resource %q %q {
  name = "%s"
  ranges = [%s
  ]
}
`
	resourceIntegerPoolRangeHCL = "\n    {first = %d, last = %d},"
)

type integerRange struct {
	first int
	last  int
}

type resourceIntegerPool struct {
	name          string
	integerRanges []integerRange
}

func (o resourceIntegerPool) render(rType, rName string) string {
	integerRanges := make([]string, len(o.integerRanges))
	for i, integerRange := range o.integerRanges {
		integerRanges[i] = fmt.Sprintf(resourceIntegerPoolRangeHCL, integerRange.first, integerRange.last)
	}

	return "" +
		fmt.Sprintf(resourceIntegerPoolHCL,
			rType, rName,
			o.name,
			strings.Join(integerRanges, ""),
		) +
		fmt.Sprintf(dataIntegerPoolByIdHCL, rType, rName, fmt.Sprintf("%s.%s.id", rType, rName)) +
		fmt.Sprintf(dataIntegerPoolByNameHCL, rType, rName, fmt.Sprintf("%s.%s.name", rType, rName))
}

func (o resourceIntegerPool) testChecks(t testing.TB, rType, rName string) testChecks {
	checks := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	checks.append(t, "TestCheckResourceAttrSet", "id")
	checks.append(t, "TestCheckResourceAttr", "name", o.name)
	checks.append(t, "TestCheckNoResourceAttr", "total")
	checks.append(t, "TestCheckNoResourceAttr", "status")
	checks.append(t, "TestCheckNoResourceAttr", "used")
	checks.append(t, "TestCheckNoResourceAttr", "used_percentage")

	checks.append(t, "TestCheckResourceAttr", "ranges.#", strconv.Itoa(len(o.integerRanges)))

	for _, r := range o.integerRanges {
		checks.appendSetNestedCheck(t, "ranges.*", map[string]string{
			"first": strconv.Itoa(r.first),
			"last":  strconv.Itoa(r.last),
		})
	}

	checks.setPath("data." + rType + "." + rName + "_by_id")
	var total int
	for _, r := range o.integerRanges {
		thisRangeTotal := 1 + r.last - r.first

		checks.appendSetNestedCheck(t, "ranges.*", map[string]string{
			"first":           strconv.Itoa(r.first),
			"last":            strconv.Itoa(r.last),
			"total":           strconv.Itoa(thisRangeTotal),
			"status":          "pool_element_available",
			"used":            "0",
			"used_percentage": "0",
		})

		total += thisRangeTotal
	}

	checks.append(t, "TestCheckResourceAttrSet", "id")
	checks.append(t, "TestCheckResourceAttr", "name", o.name)
	checks.append(t, "TestCheckResourceAttr", "total", strconv.Itoa(total))
	checks.append(t, "TestCheckResourceAttr", "status", "not_in_use")
	checks.append(t, "TestCheckResourceAttr", "used", "0")
	checks.append(t, "TestCheckResourceAttr", "used_percentage", "0")

	checks.setPath("data." + rType + "." + rName + "_by_name")
	for _, r := range o.integerRanges {
		thisRangeTotal := 1 + r.last - r.first

		checks.appendSetNestedCheck(t, "ranges.*", map[string]string{
			"first":           strconv.Itoa(r.first),
			"last":            strconv.Itoa(r.last),
			"total":           strconv.Itoa(thisRangeTotal),
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

func TestAccResourceIntegerPool(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	randomRanges := func(t testing.TB, count int) []integerRange {
		ints := randIntSet(t, 10000, 19999, count*2)
		sort.Ints(ints)

		result := make([]integerRange, count)
		for i := range count {
			result[i] = integerRange{
				first: ints[(2 * i)],
				last:  ints[(2*i)+1],
			}
		}

		return result
	}

	type testStep struct {
		config resourceIntegerPool
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"simple_case": {
			steps: []testStep{
				{
					config: resourceIntegerPool{
						name:          acctest.RandString(6),
						integerRanges: randomRanges(t, rand.Intn(5)+1),
					},
				},
				{
					config: resourceIntegerPool{
						name:          acctest.RandString(6),
						integerRanges: randomRanges(t, rand.Intn(5)+1),
					},
				},
				{
					config: resourceIntegerPool{
						name:          acctest.RandString(6),
						integerRanges: randomRanges(t, rand.Intn(5)+1),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceIntegerPool)

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
