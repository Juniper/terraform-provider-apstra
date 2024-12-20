package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

const (
	resourceDataCenterRoutingZoneConstraintHCL = `resource %q %q {
  blueprint_id                  = %q // required attribute
  name                          = %q // required attribute
  routing_zones_list_constraint = %q // required attribute
  max_count_constraint          = %s
  constraints                   = %s
}

data %[1]q "by_id" {
  blueprint_id = %[3]q
  id           = %[1]s.%[2]s.id
  depends_on   = [%[1]s.%[2]s]
}

data %[1]q "by_name" {
  blueprint_id = %[3]q
  name         = %[1]s.%[2]s.name
  depends_on   = [%[1]s.%[2]s]
}
`
)

type testRoutingZoneConstraint struct {
	name                      string
	MaxCountConstraint        *int
	RoutingZoneListConstraint enum.RoutingZoneConstraintMode
	Constraints               []string
}

func (o testRoutingZoneConstraint) render(bpId apstra.ObjectId, rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterRoutingZoneConstraintHCL,
		rType, rName,
		bpId,
		o.name,
		o.RoutingZoneListConstraint,
		intPtrOrNull(o.MaxCountConstraint),
		stringSliceOrNull(o.Constraints),
	)
}

func (o testRoutingZoneConstraint) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "routing_zones_list_constraint", o.RoutingZoneListConstraint.String())

	if o.MaxCountConstraint == nil {
		result.append(t, "TestCheckNoResourceAttr", "max_count_constraint")
	} else {
		result.append(t, "TestCheckResourceAttr", "max_count_constraint", strconv.Itoa(*o.MaxCountConstraint))
	}

	result.append(t, "TestCheckResourceAttr", "constraints.#", strconv.Itoa(len(o.Constraints)))
	for _, constraint := range o.Constraints {
		result.append(t, "TestCheckTypeSetElemAttr", "constraints.*", constraint)
	}

	return result
}

func TestResourceDatacenterRoutingZoneConstraint(t *testing.T) {
	ctx := context.Background()

	// create a blueprint
	bp := testutils.BlueprintA(t, ctx)

	routingZoneIds := make([]string, acctest.RandIntRange(5, 10))
	for i := range routingZoneIds {
		label := acctest.RandString(6)
		id, err := bp.CreateSecurityZone(ctx, &apstra.SecurityZoneData{
			Label:   label,
			SzType:  apstra.SecurityZoneTypeEVPN,
			VrfName: label,
		})
		require.NoError(t, err)
		routingZoneIds[i] = id.String()
	}

	type testStep struct {
		config testRoutingZoneConstraint
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"start_minimal": {
			steps: []testStep{
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						RoutingZoneListConstraint: enum.RoutingZoneConstraintModeNone,
					},
				},
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						MaxCountConstraint:        utils.ToPtr(acctest.RandIntRange(10, 100)),
						RoutingZoneListConstraint: oneOf(enum.RoutingZoneConstraintModeAllow, enum.RoutingZoneConstraintModeDeny),
						Constraints:               randomSelection(routingZoneIds, len(routingZoneIds)/2),
					},
				},
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						RoutingZoneListConstraint: enum.RoutingZoneConstraintModeNone,
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						MaxCountConstraint:        utils.ToPtr(acctest.RandIntRange(10, 100)),
						RoutingZoneListConstraint: oneOf(enum.RoutingZoneConstraintModeAllow, enum.RoutingZoneConstraintModeDeny),
						Constraints:               randomSelection(routingZoneIds, len(routingZoneIds)/2),
					},
				},
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						MaxCountConstraint:        utils.ToPtr(acctest.RandIntRange(10, 100)),
						RoutingZoneListConstraint: oneOf(enum.RoutingZoneConstraintModeAllow, enum.RoutingZoneConstraintModeDeny),
						Constraints:               randomSelection(routingZoneIds, len(routingZoneIds)/2),
					},
				},
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						RoutingZoneListConstraint: enum.RoutingZoneConstraintModeAllow,
					},
				},
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						RoutingZoneListConstraint: enum.RoutingZoneConstraintModeDeny,
					},
				},
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						RoutingZoneListConstraint: enum.RoutingZoneConstraintModeNone,
					},
				},
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						MaxCountConstraint:        utils.ToPtr(acctest.RandIntRange(10, 100)),
						RoutingZoneListConstraint: oneOf(enum.RoutingZoneConstraintModeAllow, enum.RoutingZoneConstraintModeDeny),
						Constraints:               randomSelection(routingZoneIds, len(routingZoneIds)/2),
					},
				},
				{
					config: testRoutingZoneConstraint{
						name:                      acctest.RandString(6),
						MaxCountConstraint:        utils.ToPtr(acctest.RandIntRange(10, 100)),
						RoutingZoneListConstraint: oneOf(enum.RoutingZoneConstraintModeAllow, enum.RoutingZoneConstraintModeDeny),
						Constraints:               randomSelection(routingZoneIds, len(routingZoneIds)/2),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterRoutingZoneConstraint)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

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
