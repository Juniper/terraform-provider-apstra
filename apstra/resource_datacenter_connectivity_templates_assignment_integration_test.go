//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const resourceDatacenterConnectivityTemplatesAssignmentTemplateHCL = `
resource %q %q {
  blueprint_id              = %q
  application_point_id      = %q
  connectivity_template_ids = %s
  fetch_ip_link_ids         = %s
}`

type resourceDatacenterConnectivityTemplatesAssignment struct {
	blueprintId             apstra.ObjectId
	applicationPointId      apstra.ObjectId
	connectivityTemplateIds []apstra.ObjectId
	fetchIpLinkIds          *bool
}

func (o resourceDatacenterConnectivityTemplatesAssignment) render(rType, rName string) string {
	return fmt.Sprintf(resourceDatacenterConnectivityTemplatesAssignmentTemplateHCL,
		rType, rName,
		o.blueprintId,
		o.applicationPointId,
		stringSliceOrNull(o.connectivityTemplateIds),
		boolPtrOrNull(o.fetchIpLinkIds),
	)
}

func (o resourceDatacenterConnectivityTemplatesAssignment) testChecks(t testing.TB, ctx context.Context, rType, rName string, bp *apstra.TwoStageL3ClosClient) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId.String())
	result.append(t, "TestCheckResourceAttr", "application_point_id", o.applicationPointId.String())
	result.append(t, "TestCheckResourceAttr", "connectivity_template_ids.#", strconv.Itoa(len(o.connectivityTemplateIds)))

	for _, connectivityTemplateId := range o.connectivityTemplateIds {
		result.append(t, "TestCheckTypeSetElemAttr", "connectivity_template_ids.*", connectivityTemplateId.String())
	}

	if o.fetchIpLinkIds == nil {
		result.append(t, "TestCheckNoResourceAttr", "fetch_ip_link_ids")
	} else {
		result.append(t, "TestCheckResourceAttr", "fetch_ip_link_ids", strconv.FormatBool(*o.fetchIpLinkIds))

		for _, connectivityTemplateId := range o.connectivityTemplateIds {
			ctVlans := testutils.DatacenterIpLinkConnectivityTemplateVlans(t, ctx, bp, connectivityTemplateId)
			for _, ctVlan := range ctVlans {
				result.append(t, "TestCheckResourceAttrSet", fmt.Sprintf(`ip_link_ids.%s.%d`, connectivityTemplateId, ctVlan))
			}
		}
	}

	return result
}

func TestAccDatacenterConnectivityTemplatesAssignment(t *testing.T) {
	ctx := context.Background()

	ifCTCount := 5
	systemCTCount := 4 // 4 leaf switches

	// Create blueprint, routing zones and connectivity templates
	bp := testutils.BlueprintA(t, ctx)
	interfaceCTs := make([]apstra.ObjectId, ifCTCount)
	for i := range interfaceCTs {
		szId := testutils.SecurityZoneA(t, ctx, bp, true)
		ctId := testutils.DatacenterConnectivityTemplateIPLink(t, ctx, bp, szId, 101+i)
		interfaceCTs[i] = ctId
	}
	systemCTs := make([]apstra.ObjectId, systemCTCount)
	for i := range systemCTs {
		szId := testutils.SecurityZoneA(t, ctx, bp, true)
		ctid := testutils.DatacenterConnectivityTemplateCustomStaticRoute(t, ctx, bp, szId)
		systemCTs[i] = ctid
	}

	interfaceAPIDs := testutils.LeafSwitchGenericSystemInterfaces(t, ctx, bp)
	require.Equal(t, 8, len(interfaceAPIDs)) // BlueprintA should have 8 single-homed generic systems

	systemAPIDmap := testutils.GetSystemIds(t, ctx, bp, "leaf")
	systemAPIDs := slices.Collect(maps.Values(systemAPIDmap))
	require.Equal(t, 4, len(systemAPIDs))

	type testCase struct {
		steps []resourceDatacenterConnectivityTemplatesAssignment
	}

	testCases := map[string]testCase{
		"single_csr_one_step": {
			steps: []resourceDatacenterConnectivityTemplatesAssignment{
				{
					blueprintId:             bp.Id(),
					applicationPointId:      systemAPIDs[0],
					connectivityTemplateIds: []apstra.ObjectId{systemCTs[0]},
				},
			},
		},
		"single_if_one_step": {
			steps: []resourceDatacenterConnectivityTemplatesAssignment{
				{
					blueprintId:             bp.Id(),
					connectivityTemplateIds: []apstra.ObjectId{interfaceCTs[0]},
					applicationPointId:      interfaceAPIDs[0],
				},
			},
		},
		"multiple_if_one_step": {
			steps: []resourceDatacenterConnectivityTemplatesAssignment{
				{
					blueprintId:             bp.Id(),
					connectivityTemplateIds: interfaceCTs,
					applicationPointId:      interfaceAPIDs[0],
				},
			},
		},
		"single_if_with_fetch": {
			steps: []resourceDatacenterConnectivityTemplatesAssignment{
				{
					blueprintId:             bp.Id(),
					connectivityTemplateIds: interfaceCTs,
					applicationPointId:      interfaceAPIDs[0],
					fetchIpLinkIds:          pointer.To(true),
				},
			},
		},
		"simple_if": {
			steps: []resourceDatacenterConnectivityTemplatesAssignment{
				{
					blueprintId:             bp.Id(),
					connectivityTemplateIds: interfaceCTs[0:1],
					applicationPointId:      interfaceAPIDs[1],
				},
				{
					blueprintId:             bp.Id(),
					connectivityTemplateIds: interfaceCTs[1:3],
					applicationPointId:      interfaceAPIDs[1],
				},
				{
					blueprintId:             bp.Id(),
					connectivityTemplateIds: interfaceCTs[3:4],
					applicationPointId:      interfaceAPIDs[1],
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterConnectivityTemplatesAssignment)

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
