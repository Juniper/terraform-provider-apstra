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
	dataAsnPoolByNameHCL = `
data %q "%s_by_name" {
  name = %s
}
`
	dataAsnPoolByIdHCL = `
data %q "%s_by_id" {
  id = %s
}
`
	resourceAsnPoolHCL = `
resource %q %q {
  name = "%s"
  ranges = [%s
  ]
}
`
	resourceAsnPoolRangeHCL = "\n    {first = %d, last = %d},"
)

type asnRange struct {
	first int
	last  int
}

type resourceAsnPool struct {
	name      string
	asnRanges []asnRange
}

func (o resourceAsnPool) render(rType, rName string) string {
	asnRanges := make([]string, len(o.asnRanges))
	for i, asnRange := range o.asnRanges {
		asnRanges[i] = fmt.Sprintf(resourceAsnPoolRangeHCL, asnRange.first, asnRange.last)
	}

	return "" +
		fmt.Sprintf(resourceAsnPoolHCL,
			rType, rName,
			o.name,
			strings.Join(asnRanges, ""),
		) +
		fmt.Sprintf(dataAsnPoolByIdHCL, rType, rName, fmt.Sprintf("%s.%s.id", rType, rName)) +
		fmt.Sprintf(dataAsnPoolByNameHCL, rType, rName, fmt.Sprintf("%s.%s.name", rType, rName))
}

func (o resourceAsnPool) testChecks(t testing.TB, rType, rName string) testChecks {
	checks := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	checks.append(t, "TestCheckResourceAttrSet", "id")
	checks.append(t, "TestCheckResourceAttr", "name", o.name)
	checks.append(t, "TestCheckNoResourceAttr", "total")
	checks.append(t, "TestCheckNoResourceAttr", "status")
	checks.append(t, "TestCheckNoResourceAttr", "used")
	checks.append(t, "TestCheckNoResourceAttr", "used_percentage")

	checks.append(t, "TestCheckResourceAttr", "ranges.#", strconv.Itoa(len(o.asnRanges)))

	for _, r := range o.asnRanges {
		checks.appendSetNestedCheck(t, "ranges.*", map[string]string{
			"first": strconv.Itoa(r.first),
			"last":  strconv.Itoa(r.last),
		})
	}

	checks.setPath("data." + rType + "." + rName + "_by_id")
	var total int
	for _, r := range o.asnRanges {
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
	for _, r := range o.asnRanges {
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

func TestAccResourceAsnPool(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	randomRanges := func(t testing.TB, count int) []asnRange {
		ints := randIntSet(t, 10000, 19999, count*2)
		sort.Ints(ints)

		result := make([]asnRange, count)
		for i := range count {
			result[i] = asnRange{
				first: ints[(2 * i)],
				last:  ints[(2*i)+1],
			}
		}

		return result
	}

	type testStep struct {
		config resourceAsnPool
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"simple_case": {
			steps: []testStep{
				{
					config: resourceAsnPool{
						name:      acctest.RandString(6),
						asnRanges: randomRanges(t, rand.Intn(5)+1),
					},
				},
				{
					config: resourceAsnPool{
						name:      acctest.RandString(6),
						asnRanges: randomRanges(t, rand.Intn(5)+1),
					},
				},
				{
					config: resourceAsnPool{
						name:      acctest.RandString(6),
						asnRanges: randomRanges(t, rand.Intn(5)+1),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceAsnPool)

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
