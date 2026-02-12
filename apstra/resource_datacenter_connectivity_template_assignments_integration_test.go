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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const resourceDatacenterConnectivityTemplateAssignmentsTemplateHCL = `
resource %q %q {
  blueprint_id             = %q
  connectivity_template_id = %q
  application_point_ids    = %s
  fetch_ip_link_ids        = %s
}`

type resourceDatacenterConnectivityTemplateAssignments struct {
	blueprintId            apstra.ObjectId
	connectivityTemplateId apstra.ObjectId
	applicationPointIds    []apstra.ObjectId
	fetchIpLinkIds         *bool
}

func (o resourceDatacenterConnectivityTemplateAssignments) render(rType, rName string) string {
	return fmt.Sprintf(resourceDatacenterConnectivityTemplateAssignmentsTemplateHCL,
		rType, rName,
		o.blueprintId,
		o.connectivityTemplateId,
		stringSliceOrNull(o.applicationPointIds),
		boolPtrOrNull(o.fetchIpLinkIds),
	)
}

func (o resourceDatacenterConnectivityTemplateAssignments) testChecks(t testing.TB, ctx context.Context, rType, rName string, bp *apstra.TwoStageL3ClosClient) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId.String())
	result.append(t, "TestCheckResourceAttr", "connectivity_template_id", o.connectivityTemplateId.String())
	result.append(t, "TestCheckResourceAttr", "application_point_ids.#", strconv.Itoa(len(o.applicationPointIds)))

	for _, applicationPointId := range o.applicationPointIds {
		result.append(t, "TestCheckTypeSetElemAttr", "application_point_ids.*", applicationPointId.String())
	}

	if o.fetchIpLinkIds == nil {
		result.append(t, "TestCheckNoResourceAttr", "fetch_ip_link_ids")
	} else {
		result.append(t, "TestCheckResourceAttr", "fetch_ip_link_ids", strconv.FormatBool(*o.fetchIpLinkIds))

		ctVlans := testutils.DatacenterIpLinkConnectivityTemplateVlans(t, ctx, bp, o.connectivityTemplateId)
		for _, applicationPointId := range o.applicationPointIds {
			for _, ctVlan := range ctVlans {
				result.append(t, "TestCheckResourceAttrSet", fmt.Sprintf(`ip_link_ids.%s.%d`, applicationPointId, ctVlan))
			}
		}
	}

	return result
}

func TestAccDatacenterConnectivityTemplateAssignments(t *testing.T) {
	ctx := context.Background()

	ctCount := 5

	// Create blueprint, routing zones and connectivity templates
	bp := testutils.BlueprintA(t, ctx)
	ctIds := make([]apstra.ObjectId, ctCount)
	for i := range ctIds {
		szID := testutils.SecurityZoneA(t, ctx, bp, true)
		ctId := testutils.DatacenterConnectivityTemplateA(t, ctx, bp, szID, 101+i)
		ctIds[i] = ctId
	}

	applicationPointIds := testutils.LeafSwitchGenericSystemInterfaces(t, ctx, bp)
	require.Equal(t, 8, len(applicationPointIds)) // BlueprintA should have 8 single-homed generic systems

	type testCase struct {
		steps []resourceDatacenterConnectivityTemplateAssignments
	}

	testCases := map[string]testCase{
		"single": {
			steps: []resourceDatacenterConnectivityTemplateAssignments{
				{
					blueprintId:            bp.Id(),
					connectivityTemplateId: ctIds[0],
					applicationPointIds:    []apstra.ObjectId{applicationPointIds[0]},
				},
			},
		},
		"single_with_fetch": {
			steps: []resourceDatacenterConnectivityTemplateAssignments{
				{
					blueprintId:            bp.Id(),
					connectivityTemplateId: ctIds[0],
					applicationPointIds:    []apstra.ObjectId{applicationPointIds[0]},
					fetchIpLinkIds:         pointer.To(true),
				},
			},
		},
		"simple": {
			steps: []resourceDatacenterConnectivityTemplateAssignments{
				{
					blueprintId:            bp.Id(),
					connectivityTemplateId: ctIds[0],
					applicationPointIds:    []apstra.ObjectId{applicationPointIds[0]},
				},
				{
					blueprintId:            bp.Id(),
					connectivityTemplateId: ctIds[0],
					applicationPointIds:    []apstra.ObjectId{applicationPointIds[1], applicationPointIds[2]},
				},
				{
					blueprintId:            bp.Id(),
					connectivityTemplateId: ctIds[0],
					applicationPointIds:    []apstra.ObjectId{applicationPointIds[3]},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterConnectivityTemplateAssignments)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() // do not use - all test data works with limited set of application points and connectivity templates

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.render(resourceType, tName)
				checks := step.testChecks(t, ctx, resourceType, tName, bp)

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
