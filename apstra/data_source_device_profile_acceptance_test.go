//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	versionconstraints "github.com/chrismarget-j/version-constraints"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

const (
	datasourceDeviceProfileHCL = `data %q %q {
  id   = %s
  name = %s
}`
)

type datasourceDeviceProfile struct {
	name string
	id   string
}

func (dp datasourceDeviceProfile) render(rType, rName string) string {
	return fmt.Sprintf(datasourceDeviceProfileHCL, rType, rName,
		stringOrNull(dp.id),
		stringOrNull(dp.name),
	)
}

func (dp datasourceDeviceProfile) testChecks(t testing.TB, rType, rName string) testChecks {
	idToName := map[string]string{ // expected mapping of canned device profiles
		"Arista_vEOS": "Arista vEOS",
		"Cisco_NXOSv": "Cisco NXOSv",
		"Juniper_vEX": "Juniper vEX",
	}
	nameToID := make(map[string]string, len(idToName))
	for k, v := range idToName {
		nameToID[v] = k
	}

	switch {
	case dp.id == "":
		dp.id = nameToID[dp.name]
	case dp.name == "":
		dp.name = idToName[dp.id]
	}

	checks := newTestChecks("data." + rType + "." + rName)
	checks.append(t, "TestCheckResourceAttr", "id", dp.id)
	checks.append(t, "TestCheckResourceAttr", "name", dp.name)

	return checks
}

func TestAccDatasourceDeviceProfile(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, os.Setenv(constants.EnvApiTimeout, "120"))
	client := testutils.GetTestClient(t, ctx)
	clientVersion, err := version.NewVersion(client.ApiVersion())
	require.NoError(t, err)

	type testStep struct {
		config                  datasourceDeviceProfile
		preApplyResourceActions []plancheck.ResourceActionType
	}

	type testCase struct {
		versionConstraints []versionconstraints.Constraints
		steps              []testStep
	}

	testCases := map[string]testCase{
		"Arista_vEOS_by_id":   {steps: []testStep{{config: datasourceDeviceProfile{id: "Arista_vEOS"}}}},
		"Cisco_NXOSv_by_id":   {steps: []testStep{{config: datasourceDeviceProfile{id: "Cisco_NXOSv"}}}},
		"Juniper_vEX_by_id":   {steps: []testStep{{config: datasourceDeviceProfile{id: "Juniper_vEX"}}}},
		"Arista_vEOS_by_name": {steps: []testStep{{config: datasourceDeviceProfile{name: "Arista vEOS"}}}},
		"Cisco_NXOSv_by_name": {steps: []testStep{{config: datasourceDeviceProfile{name: "Cisco NXOSv"}}}},
		"Juniper_vEX_by_name": {steps: []testStep{{config: datasourceDeviceProfile{name: "Juniper vEX"}}}},
	}

	resourceType := tfapstra.DatasourceName(ctx, &tfapstra.DataSourceDeviceProfile)

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

				//var checkLog string
				//var checkFuncs checks.checks
				//for _, checkList := range checks {
				//	checkLog = checkLog + checkList.string(len(checkFuncs))
				//	checkFuncs = append(checkFuncs, checkList.checks...)
				//}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, checks.string(), stepName)

				preApplyPlanChecks := make([]plancheck.PlanCheck, len(step.preApplyResourceActions))
				for i, preApplyPlanCheck := range step.preApplyResourceActions {
					preApplyPlanChecks[i] = plancheck.ExpectResourceAction(resourceType+"."+tName, preApplyPlanCheck)
				}

				steps = append(steps, resource.TestStep{
					Config:           insecureProviderConfigHCL + config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: preApplyPlanChecks},
					Check:            resource.ComposeAggregateTestCheckFunc(checks.checks...),
				})
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
