//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceDataCenterRackHCL = `
resource %q %q {
  blueprint_id = %q
  rack_type_id = %q
  name         = %q
}`
)

type resourceDataCenterRack struct {
	BlueprintId apstra.ObjectId
	RackTypeId  apstra.ObjectId
	Name        string
}

func (o resourceDataCenterRack) render(rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterRackHCL,
		rType, rName,
		o.BlueprintId,
		o.RackTypeId,
		o.Name,
	)
}

func (o resourceDataCenterRack) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// ensure ID has been set
	result.append(t, "TestCheckResourceAttrSet", "id")

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "rack_type_id", o.RackTypeId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.Name)

	return result
}

func TestResourceDatacenterRack(t *testing.T) {
	ctx := context.Background()

	bp := testutils.BlueprintC(t, ctx)

	type step struct {
		config resourceDataCenterRack
	}

	type testCase struct {
		steps              []step
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"start_with_name": {
			steps: []step{
				{
					config: resourceDataCenterRack{
						BlueprintId: bp.Id(),
						RackTypeId:  "access_switch",
						Name:        acctest.RandString(5),
					},
				},
				{
					config: resourceDataCenterRack{
						BlueprintId: bp.Id(),
						RackTypeId:  "access_switch",
						Name:        acctest.RandString(5),
					},
				},
				{
					config: resourceDataCenterRack{
						BlueprintId: bp.Id(),
						RackTypeId:  "L2_Virtual",
						Name:        acctest.RandString(5),
					},
				},
				{
					config: resourceDataCenterRack{
						BlueprintId: bp.Id(),
						RackTypeId:  "L2_Virtual",
						Name:        acctest.RandString(5),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterRack)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
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
