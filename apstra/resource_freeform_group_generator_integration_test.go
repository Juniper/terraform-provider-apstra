//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceFreeformGroupGeneratorHcl = `
resource %q %q {
  blueprint_id      = %q
  name              = %q
  scope             = %q
 }
`
)

type resourceFreeformGroupGenerator struct {
	blueprintId string
	name        string
	scope       string
}

func (o resourceFreeformGroupGenerator) render(rType, rName string) string {
	return fmt.Sprintf(resourceFreeformGroupGeneratorHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		o.scope,
	)
}

func (o resourceFreeformGroupGenerator) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "scope", o.scope)

	return result
}

func TestResourceFreeformGroupGenerator(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint and a group...
	bp, _ := testutils.FfBlueprintC(t, ctx)

	type testStep struct {
		config resourceFreeformGroupGenerator
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"start_asn_resource_generator": {
			steps: []testStep{
				{
					config: resourceFreeformGroupGenerator{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						scope:       "node('system', name='target')",
					},
				},
				{
					config: resourceFreeformGroupGenerator{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						scope:       "node('system', deploy_mode='deploy', name='target')",
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformGroupGenerator)

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
