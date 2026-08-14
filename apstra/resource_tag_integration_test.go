//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	versionconstraints "github.com/chrismarget-j/version-constraints"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

const (
	datasourceTagHCL = `data %q %q {
  id   = %s
  name = %s
}`

	resourceTagHCL = `resource %q %q {
  name        = %q
  description = %s
}`
)

type resourceTag struct {
	name        string
	description string
}

func (rt resourceTag) render(rType, rName string) string {
	resourceBlock := fmt.Sprintf(resourceTagHCL, rType, rName,
		rt.name,
		stringOrNull(rt.description),
	)
	datasourceBlockByID := fmt.Sprintf(datasourceTagHCL, rType, rName+"_by_id", fmt.Sprintf("%s.%s.id", rType, rName), "null")
	datasourceBlockByName := fmt.Sprintf(datasourceTagHCL, rType, rName+"_by_name", "null", fmt.Sprintf("%s.%s.name", rType, rName))

	return resourceBlock + "\n\n" + datasourceBlockByID + "\n\n" + datasourceBlockByName + "\n"
}

func (rt resourceTag) testChecks(t testing.TB, rType, rName string) []testChecks {
	resourceChecks := newTestChecks(rType + "." + rName)
	dataByIDChecks := newTestChecks("data." + rType + "." + rName + "_by_id")
	dataByNameChecks := newTestChecks("data." + rType + "." + rName + "_by_name")

	// required and computed attributes can always be checked
	resourceChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	resourceChecks.append(t, "TestCheckResourceAttr", "name", rt.name)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "name", rt.name)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "definition.name", rt.name)
	resourceChecks.append(t, "TestCheckResourceAttr", "definition.name", rt.name)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "definition.name", rt.name)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "definition.name", rt.name)

	if rt.description != "" {
		resourceChecks.append(t, "TestCheckResourceAttr", "description", rt.description)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "description", rt.description)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "description", rt.description)
		resourceChecks.append(t, "TestCheckResourceAttr", "definition.description", rt.description)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "definition.description", rt.description)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "definition.description", rt.description)
	} else {
		resourceChecks.append(t, "TestCheckNoResourceAttr", "description")
		dataByIDChecks.append(t, "TestCheckNoResourceAttr", "description")
		dataByNameChecks.append(t, "TestCheckNoResourceAttr", "description")
		resourceChecks.append(t, "TestCheckNoResourceAttr", "definition.description")
		dataByIDChecks.append(t, "TestCheckNoResourceAttr", "definition.description")
		dataByNameChecks.append(t, "TestCheckNoResourceAttr", "definition.description")
	}

	return []testChecks{resourceChecks, dataByIDChecks, dataByNameChecks}
}

func TestResourceTag(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	clientVersion, err := version.NewVersion(client.ApiVersion())
	require.NoError(t, err)

	type testStep struct {
		config                  resourceTag
		preApplyResourceActions []plancheck.ResourceActionType
	}

	type testCase struct {
		versionConstraints []versionconstraints.Constraints
		steps              []testStep
	}

	testCases := map[string]testCase{
		"simple": {
			versionConstraints: nil,
			steps: []testStep{
				{
					config: resourceTag{
						name:        random.PersistentString("TestResoruceTag_simple", 8),
						description: acctest.RandString(8),
					},
				},
				{
					config: resourceTag{
						name:        random.PersistentString("TestResoruceTag_simple", 8),
						description: acctest.RandString(8),
					},
				},
			},
		},
		"nullify_description": {
			versionConstraints: nil,
			steps: []testStep{
				{
					config: resourceTag{
						name:        random.PersistentString("TestResoruceTag_nullify_description", 8),
						description: acctest.RandString(8),
					},
				},
				{
					config: resourceTag{
						name: random.PersistentString("TestResoruceTag_nullify_description", 8),
					},
				},
				{
					config: resourceTag{
						name:        random.PersistentString("TestResoruceTag_nullify_description", 8),
						description: acctest.RandString(8),
					},
				},
			},
		},
		"start_with_null_description": {
			versionConstraints: nil,
			steps: []testStep{
				{
					config: resourceTag{
						name: random.PersistentString("TestResoruceTag_start_with_null_description", 8),
					},
				},
				{
					config: resourceTag{
						name:        random.PersistentString("TestResoruceTag_start_with_null_description", 8),
						description: acctest.RandString(8),
					},
				},
				{
					config: resourceTag{
						name: random.PersistentString("TestResoruceTag_start_with_null_description", 8),
					},
				},
			},
		},
		"rename_forces_replace": {
			steps: []testStep{
				{
					config: resourceTag{
						name: acctest.RandString(8),
					},
				},
				{
					preApplyResourceActions: []plancheck.ResourceActionType{plancheck.ResourceActionDestroyBeforeCreate},
					config: resourceTag{
						name: acctest.RandString(8),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceTag)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			for _, versionConstraint := range tCase.versionConstraints {
				if !versionConstraint.Check(clientVersion) {
					t.Skipf("Skipping %s: version %s not supported by test case", tName, clientVersion)
				}
			}

			steps := make([]resource.TestStep, 0, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				var checkLog string
				var checkFuncs []resource.TestCheckFunc
				for _, checkList := range checks {
					checkLog = checkLog + checkList.string(len(checkFuncs))
					checkFuncs = append(checkFuncs, checkList.checks...)
				}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, checkLog, stepName)

				preApplyPlanChecks := make([]plancheck.PlanCheck, len(step.preApplyResourceActions))
				for i, preApplyPlanCheck := range step.preApplyResourceActions {
					preApplyPlanChecks[i] = plancheck.ExpectResourceAction(resourceType+"."+tName, preApplyPlanCheck)
				}

				steps = append(steps, resource.TestStep{
					Config:           insecureProviderConfigHCL + config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: preApplyPlanChecks},
					Check:            resource.ComposeAggregateTestCheckFunc(checkFuncs...),
				})
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
