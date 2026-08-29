//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	dctestobj "github.com/Juniper/terraform-provider-apstra/internal/test_utils/datacenter_test_objects"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceDataCenterInterconnectDomainL3PolicyHCL = `resource %q %q {
  blueprint_id           = %q // required attribute
  interconnect_domain_id = %q // required attribute
  routing_zone_id        = %q // required attribute
  enabled_for_type_5     = %t // required attribute
  routing_policy_id      = %q // required attribute
  route_target           = %q // required attribute
}
`

	datasourceDataCenterInterconnectDomainL3PolicyHCL = `data %q %q {
  blueprint_id           = %q        // required attribute
  interconnect_domain_id = %q        // required attribute
  routing_zone_id        = %q        // required attribute
  depends_on             = [ %s.%s ] // ensure the data source runs after resource Create() or Update()
}
`
)

type resourceDataCenterInterconnectDomainL3Policy struct {
	InterconnectDomainID string
	RoutingZoneID        string
	EnabledForType5      bool
	RoutingPolicyID      string
	RouteTarget          string
}

func (l3p resourceDataCenterInterconnectDomainL3Policy) render(rType, rName, bpID string) string {
	r := fmt.Sprintf(resourceDataCenterInterconnectDomainL3PolicyHCL, rType, rName,
		bpID,
		l3p.InterconnectDomainID,
		l3p.RoutingZoneID,
		l3p.EnabledForType5,
		l3p.RoutingPolicyID,
		l3p.RouteTarget,
	)

	d := fmt.Sprintf(datasourceDataCenterInterconnectDomainL3PolicyHCL, rType, rName,
		bpID,
		l3p.InterconnectDomainID,
		l3p.RoutingZoneID,
		rType, rName,
	)

	return r + d
}

func (l3p resourceDataCenterInterconnectDomainL3Policy) testChecks(t testing.TB, rType, rName, bpID string) []testChecks {
	rChecks := newTestChecks(rType + "." + rName)
	dChecks := newTestChecks("data." + rType + "." + rName)

	// required and computed attributes can always be checked
	rChecks.append(t, "TestCheckResourceAttr", "blueprint_id", bpID)
	dChecks.append(t, "TestCheckResourceAttr", "blueprint_id", bpID)
	rChecks.append(t, "TestCheckResourceAttr", "interconnect_domain_id", l3p.InterconnectDomainID)
	dChecks.append(t, "TestCheckResourceAttr", "interconnect_domain_id", l3p.InterconnectDomainID)
	rChecks.append(t, "TestCheckResourceAttr", "routing_zone_id", l3p.RoutingZoneID)
	dChecks.append(t, "TestCheckResourceAttr", "routing_zone_id", l3p.RoutingZoneID)
	rChecks.append(t, "TestCheckResourceAttr", "enabled_for_type_5", strconv.FormatBool(l3p.EnabledForType5))
	dChecks.append(t, "TestCheckResourceAttr", "enabled_for_type_5", strconv.FormatBool(l3p.EnabledForType5))
	rChecks.append(t, "TestCheckResourceAttr", "routing_policy_id", l3p.RoutingPolicyID)
	dChecks.append(t, "TestCheckResourceAttr", "routing_policy_id", l3p.RoutingPolicyID)
	rChecks.append(t, "TestCheckResourceAttr", "route_target", l3p.RouteTarget)
	dChecks.append(t, "TestCheckResourceAttr", "route_target", l3p.RouteTarget)

	return []testChecks{rChecks, dChecks}
}

func TestACCResourceDatacenterInterconnectDomainL3Policy(t *testing.T) {
	ctx := context.Background()

	// Create a Blueprint.
	bp := testutils.BlueprintA(t, ctx)

	// Create an Interconnect Domain within the Blueprint.
	dciID, err := bp.CreateEVPNInterconnectGroup(ctx, apstra.EVPNInterconnectGroup{
		Label:       pointer.To(acctest.RandString(6)),
		RouteTarget: pointer.To(random.RouteTarget(t)),
	})
	require.NoError(t, err)

	type testStep struct {
		config resourceDataCenterInterconnectDomainL3Policy
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"random_x3": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      random.OneOf(true, false),
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      random.OneOf(true, false),
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      random.OneOf(true, false),
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
			},
		},
		"values_011_111_011": { // false, value, value - true, value, value - false, value, value
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      false,
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      true,
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      false,
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
			},
		},
		"values_111_011_111": { // true, value, value - false, value, value - true, value, value
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      true,
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      false,
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainL3Policy{
						InterconnectDomainID: dciID,
						RoutingZoneID:        dctestobj.RoutingZoneA(t, ctx, bp, false),
						EnabledForType5:      true,
						RoutingPolicyID:      dctestobj.RoutingPolicyRandom(t, ctx, bp),
						RouteTarget:          random.RouteTarget(t),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterInterconnectDomainL3Policy)
	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName, string(bp.Id()))
				checks := step.config.testChecks(t, resourceType, tName, string(bp.Id()))

				var checkLog string
				var checkFuncs []resource.TestCheckFunc

				for _, checkList := range checks {
					checkLog = checkLog + checkList.string(len(checkFuncs))
					checkFuncs = append(checkFuncs, checkList.checks...)
				}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, checkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checkFuncs...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})

		})
	}
}
