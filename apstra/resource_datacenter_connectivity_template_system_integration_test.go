//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const resourceDataCenterConnectivityTemplateSystemHCL = `resource %q %q {
  blueprint_id         = %q
  name                 = %q
  description          = %s
  tags                 = %s
  custom_static_routes = %s
}
`

type resourceDataCenterConnectivityTemplateSystem struct {
	blueprintId        string
	name               string
	description        string
	tags               []string
	customStaticRoutes []resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute
}

func (o resourceDataCenterConnectivityTemplateSystem) render(rType, rName string) string {
	customStaticRoutes := "null"
	if len(o.customStaticRoutes) > 0 {
		sb := new(strings.Builder)
		for _, customStaticRoute := range o.customStaticRoutes {
			sb.WriteString(customStaticRoute.render(2))
		}

		customStaticRoutes = "[\n" + sb.String() + "  ]"
	}

	return fmt.Sprintf(resourceDataCenterConnectivityTemplateSystemHCL,
		rType, rName,
		o.blueprintId,
		o.name,
		stringOrNull(o.description),
		stringSliceOrNull(o.tags),
		customStaticRoutes,
	)
}

func (o resourceDataCenterConnectivityTemplateSystem) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	if o.description != "" {
		result.append(t, "TestCheckResourceAttr", "description", o.description)
	} else {
		result.append(t, "TestCheckNoResourceAttr", "description")
	}

	result.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	for _, tag := range o.tags {
		result.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
	}

	result.append(t, "TestCheckResourceAttr", "custom_static_routes.#", strconv.Itoa(len(o.customStaticRoutes)))
	for _, customStaticRoute := range o.customStaticRoutes {
		result.appendSetNestedCheck(t, "custom_static_routes.*", customStaticRoute.valueAsMapForChecks())
	}

	return result
}

func TestResourceDatacenteConnectivityTemplateSystem(t *testing.T) {
	ctx := context.Background()
	cleanup := true

	// create a blueprint
	bp := testutils.BlueprintG(t, ctx, cleanup)

	type testStep struct {
		config resourceDataCenterConnectivityTemplateSystem
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						description:        acctest.RandString(6),
						tags:               randomStrings(rand.IntN(5)+2, 6),
						customStaticRoutes: randomCustomStaticRoutes(t, ctx, 2, 2, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						description:        acctest.RandString(6),
						tags:               randomStrings(rand.IntN(5)+2, 6),
						customStaticRoutes: randomCustomStaticRoutes(t, ctx, 2, 2, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						description:        acctest.RandString(6),
						tags:               randomStrings(rand.IntN(5)+2, 6),
						customStaticRoutes: randomCustomStaticRoutes(t, ctx, 3, 3, bp, cleanup),
					},
				},
			},
		},
		"change_labels": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						customStaticRoutes: randomCustomStaticRoutes(t, ctx, 2, 2, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						customStaticRoutes: randomCustomStaticRoutes(t, ctx, 2, 2, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						customStaticRoutes: randomCustomStaticRoutes(t, ctx, 2, 2, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSystem{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						customStaticRoutes: randomCustomStaticRoutes(t, ctx, 2, 2, bp, cleanup),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterConnectivityTemplateSystem)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, bp.Id(), resourceType, tName)

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
