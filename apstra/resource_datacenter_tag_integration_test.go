package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceDataCenterTag = `
resource %q %q {
  blueprint_id = %q
  name         = %q
  description  = %s
}
`
)

type testDatacenterTag struct {
	name        string
	description string
}

func (o testDatacenterTag) render(bpId apstra.ObjectId, rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterTag,
		rType, rName,
		bpId,
		o.name,
		stringOrNull(o.description),
	)
}

func (o testDatacenterTag) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	if o.description != "" {
		result.append(t, "TestCheckResourceAttr", "description", o.description)
	} else {
		result.append(t, "TestCheckNoResourceAttr", "description")
	}

	return result
}

func TestResourceDatacenterTag(t *testing.T) {
	ctx := context.Background()

	// create a blueprint
	bp := testutils.BlueprintA(t, ctx)

	type testStep struct {
		config testDatacenterTag
	}

	type testCase struct {
		steps []testStep
	}

	testCases := map[string]testCase{
		"start_minimal": {
			steps: []testStep{
				{
					config: testDatacenterTag{
						name: "start_minimal",
					},
				},
				{
					config: testDatacenterTag{
						name:        "start_minimal",
						description: acctest.RandString(6),
					},
				},
				{
					config: testDatacenterTag{
						name: "start_minimal",
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: testDatacenterTag{
						name:        "start_maximal",
						description: acctest.RandString(6),
					},
				},
				{
					config: testDatacenterTag{
						name: "start_maximal",
					},
				},
				{
					config: testDatacenterTag{
						name:        "start_maximal",
						description: acctest.RandString(6),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterTag)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(bp.Id(), resourceType, tName)
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
