//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"maps"
	"math/rand/v2"
	"net"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

const resourceDataCenterGenericSystemLinkHCL = `
    {
      tags                          = %s
      lag_mode                      = %s
      target_switch_id              = %q
      target_switch_if_name         = %q
      target_switch_if_transform_id = %d
      group_label                   = %s
    },`

type resourceDataCenterGenericSystemLink struct {
	tags           []string
	lagMode        apstra.RackLinkLagMode
	targetSwitchId apstra.ObjectId
	targetSwitchIf string
	targetSwitchTf int
	groupLabel     string
}

func (o *resourceDataCenterGenericSystemLink) render() string {
	return fmt.Sprintf(resourceDataCenterGenericSystemLinkHCL,
		stringSliceOrNull(o.tags),
		stringOrNull(o.lagMode.String()),
		o.targetSwitchId,
		o.targetSwitchIf,
		o.targetSwitchTf,
		stringOrNull(o.groupLabel),
	)
}

func (o *resourceDataCenterGenericSystemLink) addTestChecks(t testing.TB, testChecks *testChecks) {
	m := map[string]string{
		"target_switch_id":              o.targetSwitchId.String(),
		"target_switch_if_name":         o.targetSwitchIf,
		"target_switch_if_transform_id": strconv.Itoa(o.targetSwitchTf),
	}
	if len(o.tags) > 0 {
		m["tags.#"] = strconv.Itoa(len(o.tags))
		// todo - each tag, somehow? not critical, the extra plan stage will catch it
	}
	if o.lagMode != apstra.RackLinkLagModeNone {
		m["lag_mode"] = rosetta.StringersToFriendlyString(o.lagMode)
	}
	if o.groupLabel != "" {
		m["group_label"] = o.groupLabel
	}
	testChecks.appendSetNestedCheck(t, "links.*", m)
}

const resourceDataCenterGenericSystemHCL = `
resource %q %q {
  blueprint_id         = %q
  name                 = %s
  hostname             = %s
  asn                  = %s
  loopback_ipv4        = %s
  loopback_ipv6        = %s
  tags                 = %s
  deploy_mode          = %s
  port_channel_id_min  = %s
  port_channel_id_max  = %s
  clear_cts_on_destroy = %s
  links                = %s
}
`

type resourceDataCenterGenericSystem struct {
	name              string
	hostname          string
	asn               *int
	loopback4         *net.IPNet
	loopback6         *net.IPNet
	tags              []string
	deployMode        *string
	portChannelIdMin  *int
	portChannelIdMax  *int
	clearCtsOnDestroy *bool
	links             []resourceDataCenterGenericSystemLink
}

func (o resourceDataCenterGenericSystem) render(rType, rName string, bpId apstra.ObjectId) string {
	var loopback_ipv4, loopback_ipv6 string
	if o.loopback4 != nil {
		loopback_ipv4 = o.loopback4.String()
	}
	if o.loopback6 != nil {
		loopback_ipv6 = o.loopback6.String()
	}

	links := new(strings.Builder)
	links.WriteString("[")
	for _, link := range o.links {
		links.WriteString(link.render())
	}
	links.WriteString("\n  ]")

	return fmt.Sprintf(resourceDataCenterGenericSystemHCL,
		rType, rName,
		bpId,
		stringOrNull(o.name),
		stringOrNull(o.hostname),
		intPtrOrNull(o.asn),
		stringOrNull(loopback_ipv4),
		stringOrNull(loopback_ipv6),
		stringSliceOrNull(o.tags),
		stringPtrOrNull(o.deployMode),
		intPtrOrNull(o.portChannelIdMin),
		intPtrOrNull(o.portChannelIdMax),
		boolPtrOrNull(o.clearCtsOnDestroy),
		links.String(),
	)
}

func (o resourceDataCenterGenericSystem) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttrSet", "id")
	if o.name == "" {
		result.append(t, "TestMatchResourceAttr", "name", "^.+$")
	} else {
		result.append(t, "TestCheckResourceAttr", "name", o.name)
	}
	if o.hostname == "" {
		result.append(t, "TestMatchResourceAttr", "hostname", "^.+$")
	} else {
		result.append(t, "TestCheckResourceAttr", "hostname", o.hostname)
	}
	if o.asn != nil {
		result.append(t, "TestCheckResourceAttr", "asn", strconv.Itoa(*o.asn))
	}
	if o.loopback4 != nil {
		result.append(t, "TestCheckResourceAttr", "loopback_ipv4", o.loopback4.String())
	}
	if o.loopback6 != nil {
		result.append(t, "TestCheckResourceAttr", "loopback_ipv6", o.loopback6.String())
	}
	if len(o.tags) == 0 {
		result.append(t, "TestCheckNoResourceAttr", "tags")
	} else {
		result.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	}
	for _, tag := range o.tags {
		result.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
	}
	if o.deployMode == nil {
		result.append(t, "TestCheckResourceAttr", "deploy_mode", rosetta.StringersToFriendlyString(enum.DeployModeDeploy))
	} else {
		result.append(t, "TestCheckResourceAttr", "deploy_mode", *o.deployMode)
	}
	if o.portChannelIdMin == nil {
		result.append(t, "TestCheckResourceAttr", "port_channel_id_min", "0")
	} else {
		result.append(t, "TestCheckResourceAttr", "port_channel_id_min", strconv.Itoa(*o.portChannelIdMin))
	}
	if o.portChannelIdMax == nil {
		result.append(t, "TestCheckResourceAttr", "port_channel_id_max", "0")
	} else {
		result.append(t, "TestCheckResourceAttr", "port_channel_id_max", strconv.Itoa(*o.portChannelIdMax))
	}
	if o.clearCtsOnDestroy == nil {
		result.append(t, "TestCheckNoResourceAttr", "clear_cts_on_destroy")
	} else {
		result.append(t, "TestCheckResourceAttr", "clear_cts_on_destroy", strconv.FormatBool(*o.clearCtsOnDestroy))
	}
	result.append(t, "TestCheckResourceAttr", "links.#", strconv.Itoa(len(o.links)))

	var unknownGroupLabelCount int
	knownGroupLabels := make(map[string]struct{})

	tags := make(map[string]struct{})
	tagsToUnknownGroupLabelCount := make(map[string]int)
	tagsToKnownGroupLabels := make(map[string]map[string]struct{})

	for _, link := range o.links {
		link.addTestChecks(t, &result)

		if link.groupLabel == "" {
			unknownGroupLabelCount++
		} else {
			knownGroupLabels[link.groupLabel] = struct{}{}
		}

		for _, tag := range link.tags {
			tags[tag] = struct{}{} // keep track of every tag we see for total map size

			if link.groupLabel == "" {
				tagsToUnknownGroupLabelCount[tag]++ // links with unknown group labels are non-lag - we count them
			} else {
				if _, ok := tagsToKnownGroupLabels[tag]; !ok {
					tagsToKnownGroupLabels[tag] = make(map[string]struct{})
				}
				tagsToKnownGroupLabels[tag][link.groupLabel] = struct{}{}
			}
		}
	}

	result.append(t, "TestCheckResourceAttr", "link_application_point_ids_by_group_label.%", strconv.Itoa(unknownGroupLabelCount+len(knownGroupLabels)))
	for groupLabel := range knownGroupLabels {
		result.append(t, "TestCheckResourceAttrSet", fmt.Sprintf("link_application_point_ids_by_group_label.%s", groupLabel))
	}

	result.append(t, "TestCheckResourceAttr", "link_application_point_ids_by_tag.%", strconv.Itoa(len(tags)))
	for tag := range tags {
		result.append(t, "TestCheckResourceAttr", fmt.Sprintf("link_application_point_ids_by_tag.%s.#", tag), strconv.Itoa(tagsToUnknownGroupLabelCount[tag]+len(tagsToKnownGroupLabels[tag])))
	}

	return result
}

func TestResourceDatacenterGenericSystem(t *testing.T) {
	ctx := context.Background()

	bp := testutils.BlueprintF(t, ctx)

	// get leaf switch IDs, sorted as in web UI
	leafNameToId := testutils.GetSystemIds(t, ctx, bp, "leaf")
	leafNames := slices.Collect(maps.Keys(leafNameToId))
	sort.Strings(leafNames)
	leafSwitchIds := make([]apstra.ObjectId, len(leafNames))
	for i, leafName := range leafNames {
		leafSwitchIds[i] = leafNameToId[leafName]
	}

	// determine routing zone ID so we can create a CT
	sz, err := bp.GetSecurityZoneByVrfName(ctx, "default")
	require.NoError(t, err)

	// create the CT
	ct := apstra.ConnectivityTemplate{
		Label: acctest.RandString(6),
		Subpolicies: []*apstra.ConnectivityTemplatePrimitive{
			{
				Attributes: &apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink{
					SecurityZone:       utils.ToPtr(sz.Id),
					IPv4AddressingType: apstra.CtPrimitiveIPv4AddressingTypeNumbered,
					IPv6AddressingType: apstra.CtPrimitiveIPv6AddressingTypeNone,
				},
			},
		},
	}
	require.NoError(t, ct.SetIds())
	require.NoError(t, ct.SetUserData())
	require.NoError(t, bp.CreateConnectivityTemplate(ctx, &ct))

	attachCtToSingleLink := func(t *testing.T, swId apstra.ObjectId, ifName string) {
		t.Helper()
		ifId, err := blueprint.IfIdFromSwIdAndIfName(ctx, bp, swId, ifName)
		require.NoError(t, err)
		require.NoError(t, bp.SetApplicationPointConnectivityTemplates(ctx, ifId, []apstra.ObjectId{*ct.Id}))
	}

	attachCtToLag := func(t *testing.T, swId apstra.ObjectId, groupName string) {
		t.Helper()
		query := new(apstra.PathQuery).
			SetBlueprintId(bp.Id()).
			SetClient(bp.Client()).
			Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(swId)}}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
			Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOf.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeInterface.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("n_port_channel")},
				{Key: "if_type", Value: apstra.QEStringVal("port_channel")},
			}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.RelationshipTypeLink.QEEAttribute(),
				{Key: "group_label", Value: apstra.QEStringVal(groupName)},
			})
		var result struct {
			Items []struct {
				PortChannel struct {
					Id apstra.ObjectId `json:"id"`
				} `json:"n_port_channel"`
			} `json:"items"`
		}
		require.NoError(t, query.Do(ctx, &result))
		poIdMap := make(map[apstra.ObjectId]struct{})
		for _, item := range result.Items {
			poIdMap[item.PortChannel.Id] = struct{}{}
		}
		require.Equal(t, 1, len(poIdMap))
		for poId := range poIdMap {
			require.NoError(t, bp.SetApplicationPointConnectivityTemplates(ctx, poId, []apstra.ObjectId{*ct.Id}))
		}
	}

	type testStep struct {
		config                     resourceDataCenterGenericSystem
		preconfig                  func(t *testing.T)
		preApplyResourceActionType plancheck.ResourceActionType
		expectError                *regexp.Regexp
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"switch_port_overlap_error": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: "bogus_switch_id",
								targetSwitchIf: "0",
								targetSwitchTf: 1, // *something* must be different between these links, else tf will flatten the set
							},
							{
								targetSwitchId: "bogus_switch_id",
								targetSwitchIf: "0",
								targetSwitchTf: 2, // *something* must be different between these links, else tf will flatten the set
							},
						},
					},
					expectError: regexp.MustCompile("Multiple links use same switch and port"),
				},
			},
		},
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/0", // 0 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((0 * 100) + rand.IntN(50) + 1),  // 0 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((0 * 100) + rand.IntN(50) + 51), // 0 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagMode(rand.IntN(4)),
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/0", // 0 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     acctest.RandString(6),
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagMode(rand.IntN(4)),
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/0", // 0 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     acctest.RandString(6),
							},
						},
					},
				},
				{
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/0", // 0 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
			},
		},
		"start_maximal_fixed_lag_mode": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((1 * 100) + rand.IntN(50) + 1),  // 1 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((1 * 100) + rand.IntN(50) + 51), // 1 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/1", // 1 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     oneOf(acctest.RandString(6), ""),
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/1", // 1 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     oneOf(acctest.RandString(6), ""),
							},
						},
					},
				},
				{
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/1", // 1 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((1 * 100) + rand.IntN(50) + 1),  // 1 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((1 * 100) + rand.IntN(50) + 51), // 1 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModePassive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/1", // 1 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     oneOf(acctest.RandString(6), ""),
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModePassive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/1", // 0 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     oneOf(acctest.RandString(6), ""),
							},
						},
					},
				},
			},
		},
		"delete_with_ct_attached": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/2", // 2 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					preconfig: func(t *testing.T) {
						attachCtToSingleLink(t, leafSwitchIds[1], "xe-0/0/2") // 2 avoids conflict with other test cases
					},
					config: resourceDataCenterGenericSystem{
						clearCtsOnDestroy: utils.ToPtr(true),
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/2", // 2 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
			},
		},
		"change_speed_of_all_links": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/3", // 3 avoids conflict with other test cases
								targetSwitchTf: 2,
							},
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/3", // 3 avoids conflict with other test cases
								targetSwitchTf: 2,
							},
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionUpdate,
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/3", // 3 avoids conflict with other test cases
								targetSwitchTf: 3,
							},
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/3", // 3 avoids conflict with other test cases
								targetSwitchTf: 3,
							},
						},
					},
				},
			},
		},
		"change_speed_of_all_links_with_ct_attached": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						clearCtsOnDestroy: utils.ToPtr(true),
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/4", // 4 avoids conflict with other test cases
								targetSwitchTf: 2,
							},
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/4", // 4 avoids conflict with other test cases
								targetSwitchTf: 2,
							},
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionUpdate,
					preconfig: func(t *testing.T) {
						attachCtToSingleLink(t, leafSwitchIds[0], "ge-0/0/4") // 4 avoids conflict with other test cases
						attachCtToSingleLink(t, leafSwitchIds[1], "ge-0/0/4") // 4 avoids conflict with other test cases
					},
					config: resourceDataCenterGenericSystem{
						clearCtsOnDestroy: utils.ToPtr(true),
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/4", // 4 avoids conflict with other test cases
								targetSwitchTf: 3,
							},
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/4", // 4 avoids conflict with other test cases
								targetSwitchTf: 3,
							},
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionNoop,
					preconfig: func(t *testing.T) { // ensure CT is reattached in previous step
						// get interface (application point) IDs
						ifId1, err := blueprint.IfIdFromSwIdAndIfName(ctx, bp, leafSwitchIds[0], "ge-0/0/4") // 4 avoids conflict with other test cases
						require.NoError(t, err)
						ifId2, err := blueprint.IfIdFromSwIdAndIfName(ctx, bp, leafSwitchIds[1], "ge-0/0/4") // 4 avoids conflict with other test cases
						require.NoError(t, err)

						// determine attached CTs
						ctMap, err := bp.GetConnectivityTemplatesByApplicationPoints(ctx, []apstra.ObjectId{ifId1, ifId2})
						require.NoError(t, err)
						require.Len(t, ctMap, 2) // CT map should include both interfaces
						require.Contains(t, ctMap, ifId1)
						require.Contains(t, ctMap, ifId2)

						// count CTs assigned to each interface
						for _, ctAssigned := range ctMap {
							var count int
							for _, assigned := range ctAssigned {
								if assigned {
									count++
								}
							}
							require.Equal(t, 1, count) // each interface should have a single CT assigned
						}
					},
					config: resourceDataCenterGenericSystem{
						clearCtsOnDestroy: utils.ToPtr(true),
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/4", // 4 avoids conflict with other test cases
								targetSwitchTf: 3,
							},
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/4", // 4 avoids conflict with other test cases
								targetSwitchTf: 3,
							},
						},
					},
				},
			},
		},
		"lots_of_changes": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((5 * 100) + rand.IntN(50) + 1),  // 5 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((5 * 100) + rand.IntN(50) + 51), // 5 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/5", // 5 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bond50",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/5", // 5 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bond50",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModePassive,
								targetSwitchId: leafSwitchIds[2],
								targetSwitchIf: "xe-0/0/5", // 5 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bond51",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModePassive,
								targetSwitchId: leafSwitchIds[3],
								targetSwitchIf: "xe-0/0/5", // 5 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bond51",
							},
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionUpdate,
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((5 * 100) + rand.IntN(50) + 1),  // 5 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((5 * 100) + rand.IntN(50) + 51), // 5 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModePassive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/5", // 5 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bond50",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModePassive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/5", // 5 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bond50",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[2],
								targetSwitchIf: "xe-0/0/5", // 5 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bond51",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[3],
								targetSwitchIf: "xe-0/0/5", // 5 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bond51",
							},
						},
					},
				},
			},
		},
		"change_lagmode_and_transform": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((6 * 100) + rand.IntN(50) + 1),  // 6 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((6 * 100) + rand.IntN(50) + 51), // 6 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/6", // 6 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond60",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/6", // 6 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond60",
							},
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionDestroyBeforeCreate,
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((6 * 100) + rand.IntN(50) + 1),  // 6 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((6 * 100) + rand.IntN(50) + 51), // 6 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModePassive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/6", // 6 avoids conflict with other test cases
								targetSwitchTf: 3,
								groupLabel:     "bond60",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModePassive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/6", // 6 avoids conflict with other test cases
								targetSwitchTf: 3,
								groupLabel:     "bond60",
							},
						},
					},
				},
			},
		},
		"swap_link_between_lags": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((7 * 100) + rand.IntN(50) + 1),  // 7 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((8 * 100) + rand.IntN(50) + 51), // 8 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/7", // 7 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond70",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/7", // 7 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond71",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/8", // 8 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond70",
							},
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionUpdate,
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((7 * 100) + rand.IntN(50) + 1),  // 7 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((8 * 100) + rand.IntN(50) + 51), // 8 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/7", // 7 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond70",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/7", // 7 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond71",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/8", // 8 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond71",
							},
						},
					},
				},
			},
		},
		"change_lag_member_speed": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((9 * 100) + rand.IntN(50) + 1),  // 9 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((9 * 100) + rand.IntN(50) + 51), // 9 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/9", // 9 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond90",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/9", // 9 avoids conflict with other test cases
								targetSwitchTf: 2,
								groupLabel:     "bond90",
							},
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionUpdate,
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((9 * 100) + rand.IntN(50) + 1),  // 9 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((9 * 100) + rand.IntN(50) + 51), // 9 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "ge-0/0/9", // 9 avoids conflict with other test cases
								targetSwitchTf: 3,
								groupLabel:     "bond90",
							},
							{
								tags:           oneOf(randomStrings(3, 3), nil),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "ge-0/0/9", // 9 avoids conflict with other test cases
								targetSwitchTf: 3,
								groupLabel:     "bond90",
							},
						},
					},
				},
			},
		},
		"exercise_deploy_mode": {
			steps: []testStep{
				{ // check for deploy_mode = deploy
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/10", // 10 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
				{ // check for deploy_mode = not_set
					preApplyResourceActionType: plancheck.ResourceActionUpdate,
					config: resourceDataCenterGenericSystem{
						deployMode: utils.ToPtr(rosetta.StringersToFriendlyString(enum.DeployModeNone)),
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/10", // 10 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
				{ // check for deploy_mode = deploy
					preApplyResourceActionType: plancheck.ResourceActionUpdate,
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/10", // 10 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
			},
		},
		"orphan_lag_with_attached_ct": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						clearCtsOnDestroy: utils.ToPtr(true),
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/11", // 11 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/12", // 12 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "foo",
								lagMode:        apstra.RackLinkLagModeActive,
							},
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/13", // 13 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "bar",
								lagMode:        apstra.RackLinkLagModeActive,
							},
						},
					},
				},
				{
					preconfig: func(t *testing.T) {
						attachCtToLag(t, leafSwitchIds[0], "bar") // 2 avoids conflict with other test cases
					},
					config: resourceDataCenterGenericSystem{
						clearCtsOnDestroy: utils.ToPtr(true),
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/11", // 11 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/12", // 12 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "foo",
								lagMode:        apstra.RackLinkLagModeActive,
							},
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/13", // 13 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "foo",
								lagMode:        apstra.RackLinkLagModeActive,
							},
						},
					},
				},
			},
		},
		"replace_all_links": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/14", // 14 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/14", // 14 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionDestroyBeforeCreate,
					config: resourceDataCenterGenericSystem{
						clearCtsOnDestroy: utils.ToPtr(true),
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/15", // 15 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/15", // 15 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
			},
		},
		"overlapping_tags": {
			steps: []testStep{
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((16 * 100) + rand.IntN(50) + 1),  // 16 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((16 * 100) + rand.IntN(50) + 51), // 16 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           append(oneOf(randomStrings(3, 3), nil), "aaaaaa", "bbbbbb"),
								lagMode:        apstra.RackLinkLagMode(rand.IntN(4)),
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/16", // 16 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     acctest.RandString(6),
							},
							{
								tags:           append(oneOf(randomStrings(3, 3), nil), "aaaaaa", "bbbbbb"),
								lagMode:        apstra.RackLinkLagMode(rand.IntN(4)),
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/16", // 16 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     acctest.RandString(6),
							},
							{
								tags:           append(oneOf(randomStrings(3, 3), nil), "aaaaaa", "bbbbbb"),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/17", // 17 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "esi-lag",
							},
							{
								tags:           append(oneOf(randomStrings(3, 3), nil), "aaaaaa", "bbbbbb"),
								lagMode:        apstra.RackLinkLagModeActive,
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/17", // 17 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     "esi-lag",
							},
						},
					},
				},
				{
					config: resourceDataCenterGenericSystem{
						links: []resourceDataCenterGenericSystemLink{
							{
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/16", // 16 avoids conflict with other test cases
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					config: resourceDataCenterGenericSystem{
						name:              acctest.RandString(6),
						hostname:          acctest.RandString(6),
						asn:               utils.ToPtr(10),
						loopback4:         utils.ToPtr(randomPrefix(t, "192.0.2.0/24", 32)),
						loopback6:         utils.ToPtr(randomPrefix(t, "3fff::/20", 128)),
						tags:              oneOf(randomStrings(3, 3), nil),
						deployMode:        utils.ToPtr(oneOf(utils.AllNodeDeployModes()...)),
						portChannelIdMin:  utils.ToPtr((16 * 100) + rand.IntN(50) + 1),  // 16 avoids conflict with other test cases
						portChannelIdMax:  utils.ToPtr((16 * 100) + rand.IntN(50) + 51), // 16 avoids conflict with other test cases
						clearCtsOnDestroy: oneOf(utils.ToPtr(true), utils.ToPtr(true), nil),
						links: []resourceDataCenterGenericSystemLink{
							{
								tags:           append(oneOf(randomStrings(3, 3), nil), "aaaaaa", "bbbbbb"),
								lagMode:        apstra.RackLinkLagMode(rand.IntN(4)),
								targetSwitchId: leafSwitchIds[0],
								targetSwitchIf: "xe-0/0/16", // 16 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     acctest.RandString(6),
							},
							{
								tags:           append(oneOf(randomStrings(3, 3), nil), "aaaaaa", "bbbbbb"),
								lagMode:        apstra.RackLinkLagMode(rand.IntN(4)),
								targetSwitchId: leafSwitchIds[1],
								targetSwitchIf: "xe-0/0/16", // 16 avoids conflict with other test cases
								targetSwitchTf: 1,
								groupLabel:     acctest.RandString(6),
							},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterGenericSystem)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName, bp.Id())
				checks := step.config.testChecks(t, bp.Id(), resourceType, tName)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config:      insecureProviderConfigHCL + config,
					Check:       resource.ComposeAggregateTestCheckFunc(checks.checks...),
					ExpectError: step.expectError,
				}
				if step.preconfig != nil {
					steps[i].PreConfig = func() { step.preconfig(t) }
				}
				if step.preApplyResourceActionType != "" {
					steps[i].ConfigPlanChecks = resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(resourceType+"."+tName, step.preApplyResourceActionType),
						},
					}
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
