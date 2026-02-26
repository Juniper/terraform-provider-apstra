package tfapstra_test

import (
	"context"
	"fmt"
	"maps"
	"math"
	"math/rand"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	fftestobj "github.com/Juniper/terraform-provider-apstra/internal/test_utils/freeform_test_objects"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const datasourceFreeformAggregateLinkHCL = `
data %q %q {
  blueprint_id = %q
  id           = %s
  name         = %s
}
`

const resourceFreeformAggregateLinkHCL = `resource %q %q {
  blueprint_id    = %q
  name            = %q
  member_link_ids = %s
  tags            = %s
  endpoint_group_1 = %s
  endpoint_group_2 = %s
}
`

type resourceFreeformAggregateLink struct {
	blueprintID    string
	name           string
	memberLinkIDs  []string
	tags           []string
	endpointGroup1 resourceFreeformAggregateLinkEndpointGroup
	endpointGroup2 resourceFreeformAggregateLinkEndpointGroup
}

func (o resourceFreeformAggregateLink) render(rType, rName string) string {
	resourceBlock := fmt.Sprintf(resourceFreeformAggregateLinkHCL, rType, rName,
		o.blueprintID,
		stringOrNull(o.name),
		stringSliceOrNull(o.memberLinkIDs),
		stringSliceOrNull(o.tags),
		o.endpointGroup1.render(),
		o.endpointGroup2.render(),
	)

	datasourceBlockByID := fmt.Sprintf(datasourceFreeformAggregateLinkHCL, rType, rName+"_by_id", o.blueprintID, fmt.Sprintf("%s.%s.id", rType, rName), "null")
	datasourceBlockByName := fmt.Sprintf(datasourceFreeformAggregateLinkHCL, rType, rName+"_by_name", o.blueprintID, "null", fmt.Sprintf("%s.%s.name", rType, rName))

	return resourceBlock + datasourceBlockByID + datasourceBlockByName
}

func (o resourceFreeformAggregateLink) testChecks(t testing.TB, rType, rName string) []testChecks {
	resourceChecks := newTestChecks(rType + "." + rName)
	dataByIDChecks := newTestChecks("data." + rType + "." + rName + "_by_id")
	dataByNameChecks := newTestChecks("data." + rType + "." + rName + "_by_name")

	// required and computed attributes can always be checked
	resourceChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	resourceChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	if o.name != "" {
		resourceChecks.append(t, "TestCheckResourceAttr", "name", o.name)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "name", o.name)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "name")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "name")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "name")
	}

	o.endpointGroup1.testChecks(t, "endpoint_group_1", &resourceChecks)
	o.endpointGroup1.testChecks(t, "endpoint_group_1", &dataByIDChecks)
	o.endpointGroup1.testChecks(t, "endpoint_group_1", &dataByNameChecks)
	o.endpointGroup2.testChecks(t, "endpoint_group_2", &resourceChecks)
	o.endpointGroup2.testChecks(t, "endpoint_group_2", &dataByIDChecks)
	o.endpointGroup2.testChecks(t, "endpoint_group_2", &dataByNameChecks)

	resourceChecks.append(t, "TestCheckResourceAttr", "member_link_ids.#", strconv.Itoa(len(o.memberLinkIDs)))
	dataByIDChecks.append(t, "TestCheckResourceAttr", "member_link_ids.#", strconv.Itoa(len(o.memberLinkIDs)))
	dataByNameChecks.append(t, "TestCheckResourceAttr", "member_link_ids.#", strconv.Itoa(len(o.memberLinkIDs)))
	for _, memberLinkID := range o.memberLinkIDs {
		resourceChecks.append(t, "TestCheckTypeSetElemAttr", "member_link_ids.*", memberLinkID)
		dataByIDChecks.append(t, "TestCheckTypeSetElemAttr", "member_link_ids.*", memberLinkID)
		dataByNameChecks.append(t, "TestCheckTypeSetElemAttr", "member_link_ids.*", memberLinkID)
	}

	resourceChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByIDChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByNameChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	for _, tag := range o.tags {
		resourceChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByIDChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByNameChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
	}

	return []testChecks{resourceChecks, dataByIDChecks, dataByNameChecks}
}

const resourceFreeformAggregateLinkEndpointGroupHCL = `{
    label     = %s
    tags      = %s
    endpoints = [
      %s
    ]
  }`

type resourceFreeformAggregateLinkEndpointGroup struct {
	label     string
	tags      []string
	endpoints []resourceFreeformAggregateLinkEndpoint
}

func (o resourceFreeformAggregateLinkEndpointGroup) render() string {
	endpointHCL := make([]string, len(o.endpoints))
	for i, endpoint := range o.endpoints {
		endpointHCL[i] = endpoint.render()
	}

	return fmt.Sprintf(resourceFreeformAggregateLinkEndpointGroupHCL,
		stringOrNull(o.label),
		stringSliceOrNull(o.tags),
		strings.Join(endpointHCL, ",\n"),
	)
}

func (o *resourceFreeformAggregateLinkEndpointGroup) testChecks(t testing.TB, label string, testChecks *testChecks) {
	testChecks.append(t, "TestCheckResourceAttrSet", label+".id")

	if o.label != "" {
		testChecks.append(t, "TestCheckResourceAttr", label+".label", o.label)
	} else {
		testChecks.append(t, "TestCheckNoResourceAttr", label+".label")
	}

	testChecks.append(t, "TestCheckResourceAttr", label+".tags.#", strconv.Itoa(len(o.tags)))
	for _, tag := range o.tags {
		testChecks.append(t, "TestCheckTypeSetElemAttr", label+".tags.*", tag)
	}
	testChecks.append(t, "TestCheckResourceAttr", label+".endpoints.#", strconv.Itoa(len(o.endpoints)))
	for i, endpoint := range o.endpoints {
		endpoint.testChecks(t, fmt.Sprintf("%s.endpoints.%d", label, i), testChecks)
	}
}

const resourceFreeformAggregateLinkEndpointHCL = `{
          system_id       = %q
          if_name         = %s
		  ipv4_address    = %s
		  ipv6_address    = %s
		  port_channel_id = %d
		  lag_mode        = %q
		  tags            = %s
        }`

type resourceFreeformAggregateLinkEndpoint struct {
	systemID      string
	ifName        string
	ipv4Address   *netip.Prefix
	ipv6Address   *netip.Prefix
	portChannelID int
	lagMode       enum.LAGMode
	tags          []string
}

func (o resourceFreeformAggregateLinkEndpoint) render() string {
	return fmt.Sprintf(resourceFreeformAggregateLinkEndpointHCL,
		o.systemID,
		stringOrNull(o.ifName),
		stringerOrNull(o.ipv4Address),
		stringerOrNull(o.ipv6Address),
		o.portChannelID,
		o.lagMode.String(),
		stringSliceOrNull(o.tags),
	)
}

func (o *resourceFreeformAggregateLinkEndpoint) testChecks(t testing.TB, label string, testChecks *testChecks) {
	testChecks.append(t, "TestCheckResourceAttr", label+".system_id", o.systemID)

	if o.ifName != "" {
		testChecks.append(t, "TestCheckResourceAttr", label+".if_name", o.ifName)
	} else {
		testChecks.append(t, "TestCheckResourceAttrSet", label+".if_name")
	}

	if o.ipv4Address != nil {
		testChecks.append(t, "TestCheckResourceAttr", label+".ipv4_address", o.ipv4Address.String())
	} else {
		testChecks.append(t, "TestCheckNoResourceAttr", label+".ipv4_address")
	}

	if o.ipv6Address != nil {
		testChecks.append(t, "TestCheckResourceAttr", label+".ipv6_address", o.ipv6Address.String())
	} else {
		testChecks.append(t, "TestCheckNoResourceAttr", label+".ipv6_address")
	}

	testChecks.append(t, "TestCheckResourceAttr", label+".port_channel_id", strconv.Itoa(o.portChannelID))
	testChecks.append(t, "TestCheckResourceAttr", label+".lag_mode", o.lagMode.String())

	testChecks.append(t, "TestCheckResourceAttr", label+".tags.#", strconv.Itoa(len(o.tags)))
	for _, tag := range o.tags {
		testChecks.append(t, "TestCheckTypeSetElemAttr", label+".tags.*", tag)
	}
}

func TestResourceFreeformAggregateLink(t *testing.T) {
	ctx := context.Background()
	systemCountGroup0 := 2
	systemCountGroup1 := 2
	meshLinkCount := 3

	// create the blueprint
	client := testutils.GetTestClient(t, ctx)
	require.NotNil(t, client)
	bp, linkIDToSysIDs := fftestobj.TestBlueprintB(t, ctx, *client, systemCountGroup0, systemCountGroup1, meshLinkCount)
	_ = linkIDToSysIDs
	_ = bp

	type testCase struct {
		steps []resourceFreeformAggregateLink
	}

	linkIDs := slices.Collect(maps.Keys(linkIDToSysIDs))
	rand.Shuffle(len(linkIDs), func(i, j int) {
		linkIDs[i], linkIDs[j] = linkIDs[j], linkIDs[i]
	})

	testCases := map[string]testCase{
		"one_link_one_step": {
			steps: []resourceFreeformAggregateLink{
				{
					blueprintID:   bp.Id().String(),
					name:          acctest.RandString(6),
					memberLinkIDs: []string{linkIDs[0]},
					tags:          randomStrings(3, 6),
					endpointGroup1: resourceFreeformAggregateLinkEndpointGroup{
						label: acctest.RandString(6),
						tags:  randomStrings(3, 6),
						endpoints: []resourceFreeformAggregateLinkEndpoint{
							{
								systemID:      linkIDToSysIDs[linkIDs[0]][0],
								ifName:        fmt.Sprintf("ae%d", 1+rand.Intn(math.MaxInt8)),
								ipv4Address:   pointer.To(netip.MustParsePrefix(randIpvAddressMust(t, "192.0.2.0/24").String() + "/24")),
								ipv6Address:   pointer.To(netip.MustParsePrefix(randIpvAddressMust(t, "3fff::/64").String() + "/64")),
								portChannelID: 1 + rand.Intn(math.MaxInt8),
								lagMode:       enum.LAGModeActiveLACP,
								tags:          randomStrings(3, 6),
							},
						},
					},
					endpointGroup2: resourceFreeformAggregateLinkEndpointGroup{
						label: acctest.RandString(6),
						tags:  randomStrings(3, 6),
						endpoints: []resourceFreeformAggregateLinkEndpoint{
							{
								systemID:      linkIDToSysIDs[linkIDs[0]][1],
								ifName:        fmt.Sprintf("ae%d", 1+rand.Intn(math.MaxInt8)),
								ipv4Address:   pointer.To(netip.MustParsePrefix(randIpvAddressMust(t, "192.0.2.0/24").String() + "/24")),
								ipv6Address:   pointer.To(netip.MustParsePrefix(randIpvAddressMust(t, "3fff::/64").String() + "/64")),
								portChannelID: 1 + rand.Intn(math.MaxInt8),
								lagMode:       enum.LAGModePassiveLACP,
								tags:          randomStrings(3, 6),
							},
						},
					},
				},
			},
		},
		"one_link_start_minimal": {
			steps: []resourceFreeformAggregateLink{
				{
					blueprintID:   bp.Id().String(),
					memberLinkIDs: []string{linkIDs[0]},
					endpointGroup1: resourceFreeformAggregateLinkEndpointGroup{
						endpoints: []resourceFreeformAggregateLinkEndpoint{
							{
								systemID:      linkIDToSysIDs[linkIDs[0]][0],
								portChannelID: 1 + rand.Intn(math.MaxInt8),
								lagMode:       enum.LAGModeActiveLACP,
							},
						},
					},
					endpointGroup2: resourceFreeformAggregateLinkEndpointGroup{
						endpoints: []resourceFreeformAggregateLinkEndpoint{
							{
								systemID:      linkIDToSysIDs[linkIDs[0]][1],
								portChannelID: 1 + rand.Intn(math.MaxInt8),
								lagMode:       enum.LAGModePassiveLACP,
							},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformAggregateLink)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() don't run in parallel -- test cases are based on limited links

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.render(resourceType, tName)
				checks := step.testChecks(t, resourceType, tName)

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
