//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const resourceDataCenterConnectivityTemplateProtocolEndpointHCL = `resource %q %q {
  blueprint_id     = %q
  name             = %q
  description      = %s
  tags             = %s
  routing_policies = %s
}
`

type resourceDataCenterConnectivityTemplateProtocolEndpoint struct {
	blueprintId     string
	name            string
	description     string
	tags            []string
	routingPolicies map[string]resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy
}

func (o resourceDataCenterConnectivityTemplateProtocolEndpoint) render(rType, rName string) string {
	routingPolicies := "null"
	if len(o.routingPolicies) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.routingPolicies {
			sb.WriteString(tfapstra.Indent(2, k+" = "+v.render(2)))
		}

		routingPolicies = "{\n" + sb.String() + "  }"
	}

	return fmt.Sprintf(resourceDataCenterConnectivityTemplateProtocolEndpointHCL,
		rType, rName,
		o.blueprintId,
		o.name,
		stringOrNull(o.description),
		stringSliceOrNull(o.tags),
		routingPolicies,
	)
}

func (o resourceDataCenterConnectivityTemplateProtocolEndpoint) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
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

	result.append(t, "TestCheckResourceAttr", "routing_policies.%", strconv.Itoa(len(o.routingPolicies)))
	for k, v := range o.routingPolicies {
		for _, check := range v.testChecks("routing_policies." + k) {
			result.append(t, check[0], check[1:]...)
		}
	}

	return result
}

func TestResourceDatacenteConnectivityTemplateProtocolEndpoint(t *testing.T) {
	ctx := context.Background()
	cleanup := true

	// create a blueprint
	bp := testutils.BlueprintG(t, ctx, cleanup)

	// enable ipv6
	settings, err := bp.GetFabricSettings(ctx)
	require.NoError(t, err)
	settings.Ipv6Enabled = utils.ToPtr(true)
	require.NoError(t, bp.SetFabricSettings(ctx, settings))

	type testStep struct {
		config resourceDataCenterConnectivityTemplateProtocolEndpoint
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateProtocolEndpoint{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateProtocolEndpoint{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						description:     acctest.RandString(32),
						tags:            randomStrings(3, 6),
						routingPolicies: randomRoutingPolicies(t, ctx, 3, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateProtocolEndpoint{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateProtocolEndpoint{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						description:     acctest.RandString(32),
						tags:            randomStrings(3, 6),
						routingPolicies: randomRoutingPolicies(t, ctx, 1, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateProtocolEndpoint{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateProtocolEndpoint{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						description:     acctest.RandString(32),
						tags:            randomStrings(3, 6),
						routingPolicies: randomRoutingPolicies(t, ctx, 1, bp, cleanup),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterConnectivityTemplateProtocolEndpoint)

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
