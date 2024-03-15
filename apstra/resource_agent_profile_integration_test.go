//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

const (
	resourceDataCenterAgentProfileHCL = `resource %q %q {
  name         = %q // required attribute
  open_options = %s
  packages     = %s
  platform     = %s
}
`
)

type testAgentProfile struct {
	name        string
	openOptions map[string]string
	packages    map[string]string
	platform    string
}

func (o testAgentProfile) render(rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterAgentProfileHCL,
		rType, rName,
		o.name,
		stringMapOrNull(o.openOptions, 1),
		stringMapOrNull(o.packages, 1),
		stringOrNull(o.platform),
	)
}

func (o testAgentProfile) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttrSet", "has_username")
	result.append(t, "TestCheckResourceAttrSet", "has_password")
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	if o.platform != "" {
		result.append(t, "TestCheckResourceAttr", "platform", o.platform)
	}

	if o.openOptions != nil {
		result.append(t, "TestCheckResourceAttr", "open_options.%", strconv.Itoa(len(o.openOptions)))

		for k, v := range o.openOptions {
			result.append(t, "TestCheckResourceAttr", "open_options."+k, v)
		}
	}

	if o.packages != nil {
		result.append(t, "TestCheckResourceAttr", "packages.%", strconv.Itoa(len(o.packages)))

		for k, v := range o.packages {
			result.append(t, "TestCheckResourceAttr", "packages."+k, v)
		}
	}

	return result
}

func TestResourceAgentProfile(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)

	tNameToApId := make(map[string]string)

	setCredentials := func(t testing.TB, username, password string) {
		t.Helper()

		apId, ok := tNameToApId[t.Name()]
		if !ok {
			t.Fatalf("test %q doesn't have an Agent Profile ID", t.Name())
		}

		err := client.UpdateAgentProfile(ctx, apstra.ObjectId(apId), &apstra.AgentProfileConfig{
			Username: &username,
			Password: &password,
		})
		require.NoError(t, err)
	}
	_ = setCredentials

	type extraCheck struct {
		testFuncName string
		testFuncArgs []string
	}

	type extraction struct {
		attribute string
		target    map[string]string
	}

	type step struct {
		config      testAgentProfile
		extraChecks []extraCheck
		preConfig   func(testing.TB)
		extractions []extraction
	}

	type testCase struct {
		steps []step
	}

	testCases := map[string]testCase{
		"a": {
			steps: []step{
				{
					config: testAgentProfile{
						name: acctest.RandString(6),
					},
					extraChecks: []extraCheck{
						{testFuncName: "TestCheckResourceAttr", testFuncArgs: []string{"has_username", "false"}},
						{testFuncName: "TestCheckResourceAttr", testFuncArgs: []string{"has_password", "false"}},
					},
					extractions: []extraction{
						{
							attribute: "id",
							target:    tNameToApId,
						},
					},
				},
				{
					preConfig: func(t testing.TB) {
						setCredentials(t, acctest.RandString(6), acctest.RandString(6))
					},
					config: testAgentProfile{
						name:     acctest.RandString(6),
						platform: apstra.AgentPlatformJunos.String(),
					},
					extraChecks: []extraCheck{
						{testFuncName: "TestCheckResourceAttr", testFuncArgs: []string{"has_username", "true"}},
						{testFuncName: "TestCheckResourceAttr", testFuncArgs: []string{"has_password", "true"}},
					},
				},
				{
					preConfig: func(t testing.TB) {
						setCredentials(t, "", "")
					},
					config: testAgentProfile{
						name: acctest.RandString(6),
					},
					extraChecks: []extraCheck{
						{testFuncName: "TestCheckResourceAttr", testFuncArgs: []string{"has_username", "false"}},
						{testFuncName: "TestCheckResourceAttr", testFuncArgs: []string{"has_password", "false"}},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceAgentProfile)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				// add extra checks
				for _, ec := range step.extraChecks {
					checks.append(t, ec.testFuncName, ec.testFuncArgs...)
				}

				// add extractions
				for _, ex := range step.extractions {
					checks.extractFromState(t, ex.attribute, tNameToApId)
				}

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				var preconfig func()
				if step.preConfig != nil {
					t, f := t, step.preConfig
					preconfig = func() {
						f(t)
					}
				}

				steps[i] = resource.TestStep{
					PreConfig: preconfig,
					Config:    insecureProviderConfigHCL + config,
					Check:     resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
