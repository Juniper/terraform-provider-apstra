//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"math/rand"
	"strconv"
	"testing"
)

const (
	resourceFreeformConfigTemplateHcl = `
resource %q %q {
  blueprint_id = %q
  name = %q
  text = %q
  tags = %s
}
`
)

type resourceFreeformConfigTemplate struct {
	blueprintId string
	name        string
	text        string
	tags        []string
}

func (o resourceFreeformConfigTemplate) render(rType, rName string) string {
	return fmt.Sprintf(resourceFreeformConfigTemplateHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		o.text,
		stringSetOrNull(o.tags),
	)
}

func (o resourceFreeformConfigTemplate) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "text", o.text)

	if len(o.tags) > 0 {
		result.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
		for _, tag := range o.tags {
			result.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		}
	}

	return result
}

func TestResourceFreeformConfigTemplate(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint
	bp := testutils.FfBlueprintA(t, ctx)

	type testStep struct {
		config resourceFreeformConfigTemplate
	}
	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"start_with_no_tags": {
			steps: []testStep{
				{
					config: resourceFreeformConfigTemplate{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6) + ".jinja",
						text:        acctest.RandString(6),
					},
				},
				{
					config: resourceFreeformConfigTemplate{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6) + ".jinja",
						text:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
					},
				},
				{
					config: resourceFreeformConfigTemplate{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6) + ".jinja",
						text:        acctest.RandString(6),
					},
				},
			},
		},
		"start_with_tags": {
			steps: []testStep{
				{
					config: resourceFreeformConfigTemplate{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6) + ".jinja",
						text:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
					},
				},
				{
					config: resourceFreeformConfigTemplate{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6) + ".jinja",
						text:        acctest.RandString(6),
					},
				},
				{
					config: resourceFreeformConfigTemplate{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6) + ".jinja",
						text:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformConfigTemplate)

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
