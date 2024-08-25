//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const resourceDataCenterConnectivityTemplateInterfaceHCL = `resource %q %q {
  blueprint_id              = %q
  name                      = %q
  description               = %s
  tags                      = %s
  ip_links                  = %s
  routing_zone_constraints  = %s
  virtual_network_multiples = %s
  virtual_network_singles   = %s
}
`

type resourceDataCenterConnectivityTemplateInterface struct {
	blueprintId             string
	name                    string
	description             string
	tags                    []string
	ipLinks                 []resourceDataCenterConnectivityTemplatePrimitiveIpLink
	routingZoneConstraints  []resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint
	virtualNetworkMultiples []resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple
	virtualNetworkSingles   []resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle
}

func (o resourceDataCenterConnectivityTemplateInterface) render(rType, rName string) string {
	ipLinks := "null"
	if len(o.ipLinks) > 0 {
		sb := new(strings.Builder)
		for _, ipLink := range o.ipLinks {
			sb.WriteString(ipLink.render(2))
		}

		ipLinks = "[\n" + sb.String() + "  ]"
	}

	routingZoneConstraints := "null"
	if len(o.routingZoneConstraints) > 0 {
		sb := new(strings.Builder)
		for _, routingZoneConstraint := range o.routingZoneConstraints {
			sb.WriteString(routingZoneConstraint.render(2))
		}

		routingZoneConstraints = "[\n" + sb.String() + "  ]"
	}

	virtualNetworkMultiples := "null"
	if len(o.virtualNetworkMultiples) > 0 {
		sb := new(strings.Builder)
		for _, virtualNetworkMultiple := range o.virtualNetworkMultiples {
			sb.WriteString(virtualNetworkMultiple.render(2))
		}

		virtualNetworkMultiples = "[\n" + sb.String() + "  ]"
	}

	virtualNetworkSingles := "null"
	if len(o.virtualNetworkSingles) > 0 {
		sb := new(strings.Builder)
		for _, virtualNetworkSingle := range o.virtualNetworkSingles {
			sb.WriteString(virtualNetworkSingle.render(2))
		}

		virtualNetworkSingles = "[\n" + sb.String() + "  ]"
	}

	return fmt.Sprintf(resourceDataCenterConnectivityTemplateInterfaceHCL,
		rType, rName,
		o.blueprintId,
		o.name,
		stringOrNull(o.description),
		stringSetOrNull(o.tags),
		ipLinks,
		routingZoneConstraints,
		virtualNetworkMultiples,
		virtualNetworkSingles,
	)
}

func (o resourceDataCenterConnectivityTemplateInterface) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
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

	result.append(t, "TestCheckResourceAttr", "ip_links.#", strconv.Itoa(len(o.ipLinks)))
	for _, ipLink := range o.ipLinks {
		result.appendSetNestedCheck(t, "ip_links.*", ipLink.valueAsMapForChecks())
	}

	result.append(t, "TestCheckResourceAttr", "routing_zone_constraints.#", strconv.Itoa(len(o.routingZoneConstraints)))
	for _, routingZoneConstraint := range o.routingZoneConstraints {
		result.appendSetNestedCheck(t, "routing_zone_constraints.*", routingZoneConstraint.valueAsMapForChecks())
	}

	result.append(t, "TestCheckResourceAttr", "virtual_network_multiples.#", strconv.Itoa(len(o.virtualNetworkMultiples)))
	for _, virtualNetworkMultiple := range o.virtualNetworkMultiples {
		result.appendSetNestedCheck(t, "virtual_network_multiples.*", virtualNetworkMultiple.valueAsMapForChecks())
	}

	result.append(t, "TestCheckResourceAttr", "virtual_network_singles.#", strconv.Itoa(len(o.virtualNetworkSingles)))
	for _, virtualNetworkSingle := range o.virtualNetworkSingles {
		result.appendSetNestedCheck(t, "virtual_network_singles.*", virtualNetworkSingle.valueAsMapForChecks())
	}

	return result
}

func TestResourceDatacenteConnectivityTemplateInterface(t *testing.T) {
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
		config resourceDataCenterConnectivityTemplateInterface
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateInterface{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateInterface{
						blueprintId:             bp.Id().String(),
						name:                    acctest.RandString(6),
						description:             acctest.RandString(32),
						tags:                    randomStrings(3, 6),
						ipLinks:                 randomIpLinks(t, ctx, 3, bp, cleanup),
						routingZoneConstraints:  randomRoutingZoneConstraints(t, ctx, 3, bp, cleanup),
						virtualNetworkMultiples: randomVirtualNetworkMultiples(t, ctx, 3, bp, cleanup),
						virtualNetworkSingles:   randomVirtualNetworkSingles(t, ctx, 3, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateInterface{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: resourceDataCenterConnectivityTemplateInterface{
						blueprintId:             bp.Id().String(),
						name:                    acctest.RandString(6),
						description:             acctest.RandString(32),
						tags:                    randomStrings(3, 6),
						ipLinks:                 randomIpLinks(t, ctx, 3, bp, cleanup),
						routingZoneConstraints:  randomRoutingZoneConstraints(t, ctx, 3, bp, cleanup),
						virtualNetworkMultiples: randomVirtualNetworkMultiples(t, ctx, 3, bp, cleanup),
						virtualNetworkSingles:   randomVirtualNetworkSingles(t, ctx, 3, bp, cleanup),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateInterface{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterConnectivityTemplateInterface{
						blueprintId:             bp.Id().String(),
						name:                    acctest.RandString(6),
						description:             acctest.RandString(32),
						tags:                    randomStrings(3, 6),
						ipLinks:                 randomIpLinks(t, ctx, 3, bp, cleanup),
						routingZoneConstraints:  randomRoutingZoneConstraints(t, ctx, 3, bp, cleanup),
						virtualNetworkMultiples: randomVirtualNetworkMultiples(t, ctx, 3, bp, cleanup),
						virtualNetworkSingles:   randomVirtualNetworkSingles(t, ctx, 3, bp, cleanup),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterConnectivityTemplateInterface)

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
