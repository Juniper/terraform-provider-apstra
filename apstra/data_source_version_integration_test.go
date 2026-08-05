//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	itestutils "github.com/Juniper/terraform-provider-apstra/internal/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const dataSourceVersionHCL = `data %q %q {
  checks = %s
}`

type dataSourceVersion struct {
	checks   map[string]string
	expected map[string]bool
}

func (d dataSourceVersion) render(rType, rName string) string {
	return fmt.Sprintf(dataSourceVersionHCL,
		rType, rName,
		stringMapOrNull(d.checks, 1),
	)
}

func (d dataSourceVersion) testChecks(t testing.TB, _ context.Context, rType, rName string, client *apstra.Client) testChecks {
	result := newTestChecks(rType + "." + rName)
	result.append(t, "TestCheckResourceAttr", "version", client.ApiVersion())
	if d.checks == nil {
		result.append(t, "TestCheckNoResourceAttr", "results")
	} else {
		result.append(t, "TestCheckResourceAttr", "results.%", strconv.Itoa(len(d.checks)))
		for k := range d.checks {
			result.append(t, "TestCheckResourceAttrSet", "results."+k)
		}
	}
	for k, v := range d.expected {
		result.append(t, "TestCheckResourceAttr", "results."+k, strconv.FormatBool(v))
	}
	return result
}

func TestAccDatasourceVersion(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		constraints version.Constraints
		config      dataSourceVersion
	}

	testCases := map[string]testCase{
		"empty": {
			config: dataSourceVersion{},
		},
		"version_zero": {
			config: dataSourceVersion{
				checks: map[string]string{
					"less_than_zero":        "<0.0.0",
					"less_or_equal_zero":    "<=0.0.0",
					"equal_zero":            "0.0.0",
					"greater_or_equal_zero": ">=0.0.0",
					"greater_than_zero":     ">0.0.0",
				},
				expected: map[string]bool{
					"less_than_zero":        false,
					"less_or_equal_zero":    false,
					"equal_zero":            false,
					"greater_or_equal_zero": true,
					"greater_than_zero":     true,
				},
			},
		},
		"version_hundred": {
			config: dataSourceVersion{
				checks: map[string]string{
					"less_than_hundred":        "<100.0.0",
					"less_or_equal_hundred":    "<=100.0.0",
					"equal_hundred":            "100.0.0",
					"greater_or_equal_hundred": ">=100.0.0",
					"greater_than_hundred":     ">100.0.0",
				},
				expected: map[string]bool{
					"less_than_hundred":        true,
					"less_or_equal_hundred":    true,
					"equal_hundred":            false,
					"greater_or_equal_hundred": false,
					"greater_than_hundred":     false,
				},
			},
		},
		"version_six_one_two": {
			constraints: itestutils.Must(version.NewConstraint("6.1.2")),
			config: dataSourceVersion{
				checks: map[string]string{
					"less_than_six_one_two":        "<6.1.2",
					"less_or_equal_six_one_two":    "<=6.1.2",
					"equal_six_one_two":            "6.1.2",
					"greater_or_equal_six_one_two": ">=6.1.2",
					"greater_than_six_one_two":     ">6.1.2",
				},
				expected: map[string]bool{
					"less_than_six_one_two":        false,
					"less_or_equal_six_one_two":    true,
					"equal_six_one_two":            true,
					"greater_or_equal_six_one_two": true,
					"greater_than_six_one_two":     false,
				},
			},
		},
	}

	resourceType := tfapstra.DatasourceName(ctx, &tfapstra.DataSourceVersion)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			v := version.Must(version.NewVersion(testutils.GetTestClient(t, ctx).ApiVersion()))
			for _, constraint := range tCase.constraints {
				if !constraint.Check(v) {
					t.Skipf("skipping with Apstra %s due to version constraints %q", v, constraint)
				}
			}

			config := tCase.config.render(resourceType, tName)
			checks := tCase.config.testChecks(t, ctx, "data."+resourceType, tName, testutils.GetTestClient(t, ctx))

			chkLog := checks.string()
			stepName := fmt.Sprintf("test case %q", tName)

			t.Logf("\n// ------ begin config for %s ------\n%s\n// -------- end config for %s ------\n\n", stepName, config, stepName)
			t.Logf("\n// ------ begin checks for %s ------\n%s\n// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}},
			})
		})
	}
}
