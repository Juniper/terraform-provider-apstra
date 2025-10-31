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
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const resourceDataCenterConnectivityTemplateSviHCL = `resource %q %q {
  blueprint_id             = %q
  name                     = %q
  description              = %s
  tags                     = %s
  bgp_peering_ip_endpoints = %s
  dynamic_bgp_peerings     = %s
}
`

type resourceDataCenterConnectivityTemplateSvi struct {
	blueprintId          string
	name                 string
	description          string
	tags                 []string
	bgpPeeringIpEndoints map[string]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint
	dynamicBgpPeerings   map[string]resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering
}

func (o resourceDataCenterConnectivityTemplateSvi) render(rType, rName string) string {
	bgpPeeringIpEndoints := "null"
	if len(o.bgpPeeringIpEndoints) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.bgpPeeringIpEndoints {
			sb.WriteString(tfapstra.Indent(2, k+" = "+v.render(2)))
		}

		bgpPeeringIpEndoints = "{\n" + sb.String() + "  }"
	}

	dynamicBgpPeerings := "null"

	if len(o.dynamicBgpPeerings) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.dynamicBgpPeerings {
			sb.WriteString(tfapstra.Indent(2, k+" = "+v.render(2)))
		}

		dynamicBgpPeerings = "{\n" + sb.String() + "  }"
	}

	return fmt.Sprintf(resourceDataCenterConnectivityTemplateSviHCL,
		rType, rName,
		o.blueprintId,
		o.name,
		stringOrNull(o.description),
		stringSliceOrNull(o.tags),
		bgpPeeringIpEndoints,
		dynamicBgpPeerings,
	)
}

func (o resourceDataCenterConnectivityTemplateSvi) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
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

	result.append(t, "TestCheckResourceAttr", "bgp_peering_ip_endpoints.%", strconv.Itoa(len(o.bgpPeeringIpEndoints)))
	for k, v := range o.bgpPeeringIpEndoints {
		for _, check := range v.testChecks("bgp_peering_ip_endpoints." + k) {
			result.append(t, check[0], check[1:]...)
		}
	}

	result.append(t, "TestCheckResourceAttr", "dynamic_bgp_peerings.%", strconv.Itoa(len(o.dynamicBgpPeerings)))
	for k, v := range o.dynamicBgpPeerings {
		for _, check := range v.testChecks("dynamic_bgp_peerings." + k) {
			result.append(t, check[0], check[1:]...)
		}
	}

	return result
}

func TestResourceDatacenteConnectivityTemplateSvi(t *testing.T) {
	ctx := context.Background()
	cleanup := true

	// Create a blueprint
	bp := testutils.BlueprintG(t, ctx, cleanup)

	// Enable IPv6
	fs, err := bp.GetFabricSettings(ctx)
	require.NoError(t, err)
	fs.Ipv6Enabled = pointer.To(true)
	err = bp.SetFabricSettings(ctx, fs)
	require.NoError(t, err)

	type testStep struct {
		config resourceDataCenterConnectivityTemplateSvi
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateSvi{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSvi{
						blueprintId:          bp.Id().String(),
						name:                 acctest.RandString(6),
						description:          acctest.RandString(6),
						tags:                 randomStrings(rand.IntN(5)+2, 6),
						bgpPeeringIpEndoints: randomBgpPeeringIpEndpointPrimitives(t, ctx, rand.IntN(3)+2, bp, cleanup),
						dynamicBgpPeerings:   randomDynamicBgpPeeringPrimitives(t, ctx, rand.IntN(3)+2, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSvi{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateSvi{
						blueprintId:          bp.Id().String(),
						name:                 acctest.RandString(6),
						description:          acctest.RandString(6),
						tags:                 randomStrings(rand.IntN(5)+2, 6),
						bgpPeeringIpEndoints: randomBgpPeeringIpEndpointPrimitives(t, ctx, rand.IntN(3)+2, bp, cleanup),
						dynamicBgpPeerings:   randomDynamicBgpPeeringPrimitives(t, ctx, rand.IntN(3)+2, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSvi{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateSvi{
						blueprintId:          bp.Id().String(),
						name:                 acctest.RandString(6),
						description:          acctest.RandString(6),
						tags:                 randomStrings(rand.IntN(5)+2, 6),
						bgpPeeringIpEndoints: randomBgpPeeringIpEndpointPrimitives(t, ctx, rand.IntN(3)+2, bp, cleanup),
						dynamicBgpPeerings:   randomDynamicBgpPeeringPrimitives(t, ctx, rand.IntN(3)+2, bp, cleanup),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterConnectivityTemplateSvi)

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
