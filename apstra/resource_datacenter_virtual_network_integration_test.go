//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	versionconstraints "github.com/chrismarget-j/version-constraints"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

const (
	resourceDatacenterVirtualNetworkTemplateHCL = `
resource %q %q {
  blueprint_id              = %q
  name                      = %q
  description               = %s
  type                      = %s
  vni                       = %s
  routing_zone_id           = %s
  l3_mtu                    = %s
  bindings                  = %s
  reserve_vlan              = %s
  reserved_vlan_id          = %s
  tags                      = %s
  dhcp_service_enabled      = %s
  ipv4_connectivity_enabled = %s
  ipv6_connectivity_enabled = %s
}
`
	resourceDatacenterVirtualNetworkTemplateBindingHCL = `
    %q = {
      vlan_id    = %s
      access_ids = %s
    },
`
)

type resourceDatacenterVirtualNetworkTemplate struct {
	blueprintId             apstra.ObjectId
	name                    string
	description             string
	vnType                  string
	vni                     *int
	routingZoneId           apstra.ObjectId
	l3Mtu                   *int
	bindings                []resourceDatacenterVirtualNetworkTemplateBinding
	reserveVlan             *bool
	reservedVlanId          *int
	tags                    []string
	dhcpEnabled             *bool
	ipv4ConnectivityEnabled *bool
	ipv6ConnectivityEnabled *bool
}

func (o resourceDatacenterVirtualNetworkTemplate) render(rType, rName string) string {
	bindings := new(strings.Builder)
	if o.bindings == nil {
		bindings.WriteString("null")
	} else {
		bindings.WriteString("{")
		for _, binding := range o.bindings {
			r := binding.render()
			bindings.WriteString(r)
		}
		bindings.WriteString("  }")
	}

	return fmt.Sprintf(resourceDatacenterVirtualNetworkTemplateHCL,
		rType, rName,
		o.blueprintId,
		o.name,
		stringOrNull(o.description),
		stringOrNull(o.vnType),
		intPtrOrNull(o.vni),
		stringOrNull(o.routingZoneId.String()),
		intPtrOrNull(o.l3Mtu),
		bindings.String(),
		boolPtrOrNull(o.reserveVlan),
		intPtrOrNull(o.reservedVlanId),
		stringSliceOrNull(o.tags),
		boolPtrOrNull(o.dhcpEnabled),
		boolPtrOrNull(o.ipv4ConnectivityEnabled),
		boolPtrOrNull(o.ipv6ConnectivityEnabled),
	)
}

func (o resourceDatacenterVirtualNetworkTemplate) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	if o.vnType != "" {
		result.append(t, "TestCheckResourceAttr", "type", o.vnType)
	} else {
		result.append(t, "TestCheckResourceAttr", "type", enum.VnTypeVxlan.String())
	}

	if o.vni != nil {
		result.append(t, "TestCheckResourceAttr", "vni", strconv.Itoa(*o.vni))
	}

	if o.routingZoneId != "" {
		result.append(t, "TestCheckResourceAttr", "routing_zone_id", o.routingZoneId.String())
	}

	if o.l3Mtu != nil {
		result.append(t, "TestCheckResourceAttr", "l3_mtu", strconv.Itoa(*o.l3Mtu))
	}

	if o.bindings != nil {
		result.append(t, "TestCheckResourceAttr", "bindings.%", strconv.Itoa(len(o.bindings)))
		for _, binding := range o.bindings {
			binding.addTestChecks(t, &result)
		}
	} else {
		result.append(t, "TestCheckNoResourceAttr", "bindings")
	}

	result.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	for _, tag := range o.tags {
		result.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
	}

	if o.dhcpEnabled == nil {
		result.append(t, "TestCheckResourceAttr", "dhcp_service_enabled", strconv.FormatBool(false))
	} else {
		result.append(t, "TestCheckResourceAttr", "dhcp_service_enabled", strconv.FormatBool(*o.dhcpEnabled))
	}

	if o.ipv4ConnectivityEnabled == nil {
		result.append(t, "TestCheckResourceAttr", "ipv4_connectivity_enabled", strconv.FormatBool(true))
	} else {
		if *o.ipv4ConnectivityEnabled {
			result.append(t, "TestCheckResourceAttr", "ipv4_connectivity_enabled", strconv.FormatBool(true))
		} else {
			result.append(t, "TestCheckResourceAttr", "ipv4_connectivity_enabled", strconv.FormatBool(false))
		}
	}

	if o.ipv6ConnectivityEnabled == nil {
		result.append(t, "TestCheckResourceAttr", "ipv6_connectivity_enabled", strconv.FormatBool(false))
	} else {
		if *o.ipv6ConnectivityEnabled {
			result.append(t, "TestCheckResourceAttr", "ipv6_connectivity_enabled", strconv.FormatBool(true))
		} else {
			result.append(t, "TestCheckResourceAttr", "ipv6_connectivity_enabled", strconv.FormatBool(false))
		}
	}

	return result
}

type resourceDatacenterVirtualNetworkTemplateBinding struct {
	leafId    string
	vlanId    *int
	accessIds []string
}

func (o resourceDatacenterVirtualNetworkTemplateBinding) render() string {
	return fmt.Sprintf(
		resourceDatacenterVirtualNetworkTemplateBindingHCL,
		o.leafId,
		intPtrOrNull(o.vlanId),
		stringSliceOrNull(o.accessIds),
	)
}

func (o resourceDatacenterVirtualNetworkTemplateBinding) addTestChecks(t testing.TB, testChecks *testChecks) {
	if o.vlanId != nil {
		testChecks.append(t, "TestCheckResourceAttr", "bindings."+o.leafId+".vlan_id", strconv.Itoa(*o.vlanId))
		testChecks.append(t, "TestCheckResourceAttr", "bindings."+o.leafId+".access_ids.#", strconv.Itoa(len(o.accessIds)))
		for _, access := range o.accessIds {
			testChecks.append(t, "TestCheckTypeSetElemAttr", "bindings."+o.leafId+".access_ids.*", access)
		}
	}
}

func TestAccDatacenterVirtualNetwork(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// The action taken when RZ ID is changed depends on the Apstra version.
	// Prior to 5.0.0 the VN would need to be recreated.
	var rzChangeResourceActionType plancheck.ResourceActionType
	if apiVersion.LessThan(version.Must(version.NewVersion("5.0.0"))) {
		rzChangeResourceActionType = plancheck.ResourceActionDestroyBeforeCreate
	} else {
		rzChangeResourceActionType = plancheck.ResourceActionUpdate
	}

	// Create blueprint and routing zone
	bp := testutils.BlueprintC(t, ctx)
	szId := testutils.SecurityZoneA(t, ctx, bp, true)

	// struct used for both system nodes and redundancy group nodes
	type node struct {
		Id         string `json:"id"`
		Label      string `json:"label"`
		SystemType string `json:"system_type"` // only applies to system nodes
		Role       string `json:"role"`        // only applies to system nodes
	}

	var systemNodesResponse struct {
		Nodes map[string]node `json:"nodes"`
	}
	err := bp.GetNodes(ctx, apstra.NodeTypeSystem, &systemNodesResponse)
	require.NoError(t, err)

	var redundancyGroupNodesResponse struct {
		Nodes map[string]node `json:"nodes"`
	}
	err = bp.GetNodes(ctx, apstra.NodeTypeRedundancyGroup, &redundancyGroupNodesResponse)
	require.NoError(t, err)

	nodesByLabel := make(map[string]string)
	for k, v := range systemNodesResponse.Nodes {
		if v.SystemType == "switch" && (v.Role == "leaf" || v.Role == "access") {
			nodesByLabel[v.Label] = k
		}
	}
	for k, v := range redundancyGroupNodesResponse.Nodes {
		nodesByLabel[v.Label] = k
	}

	type testStep struct {
		config                     resourceDatacenterVirtualNetworkTemplate
		expectError                *regexp.Regexp
		expectNonEmptyPlan         bool
		preApplyResourceActionType plancheck.ResourceActionType
	}

	type testCase struct {
		apiVersionConstraints versionconstraints.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"invalid_dhcp_without_ip": {
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						routingZoneId: szId,
						dhcpEnabled:   utils.ToPtr(true),
						bindings:      []resourceDatacenterVirtualNetworkTemplateBinding{{leafId: nodesByLabel["l2_one_access_001_leaf1"]}},
					},
					expectError: regexp.MustCompile("When `dhcp_service_enabled` is set, at least one"),
				},
			},
		},
		"no_bindings_vlan_start_minimal": {
			apiVersionConstraints: compatibility.VnEmptyBindingsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
					},
				},
			},
		},
		"no_bindings_vlan_start_maximal": {
			apiVersionConstraints: compatibility.VnEmptyBindingsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8900),
					},
				},
			},
		},
		"no_bindings_vxlan_start_minimal": {
			apiVersionConstraints: compatibility.VnEmptyBindingsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						vni:           utils.ToPtr(rand.IntN(10000) + 5000),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
			},
		},
		"no_bindings_vxlan_start_maximal": {
			apiVersionConstraints: compatibility.VnEmptyBindingsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						vni:           utils.ToPtr(rand.IntN(10000) + 5000),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						vni:           utils.ToPtr(rand.IntN(10000) + 5000),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8900),
					},
				},
			},
		},
		"vlan_with_binding_start_minimal": {
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_one_access_001_leaf1"],
							},
						},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8900),
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId:    nodesByLabel["l2_one_access_002_leaf1"],
								vlanId:    utils.ToPtr(51),
								accessIds: []string{nodesByLabel["l2_one_access_002_access1"]},
							},
						},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_one_access_003_leaf1"],
							},
						},
					},
				},
			},
		},
		"vlan_with_binding_start_maximal": {
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId:    nodesByLabel["l2_one_access_001_leaf1"],
								vlanId:    utils.ToPtr(61),
								accessIds: []string{nodesByLabel["l2_one_access_001_access1"]},
							},
						},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_one_access_001_leaf1"],
							},
						},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8900),
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId:    nodesByLabel["l2_one_access_002_leaf1"],
								vlanId:    utils.ToPtr(63),
								accessIds: []string{nodesByLabel["l2_one_access_002_access1"]},
							},
						},
					},
				},
			},
		},
		"vxlan_with_binding_start_minimal": {
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_one_access_001_leaf1"],
							},
						},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						vni:           nil,
						routingZoneId: szId,
						l3Mtu:         nil,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId:    nodesByLabel["l2_one_access_002_leaf1"],
								vlanId:    utils.ToPtr(721),
								accessIds: []string{nodesByLabel["l2_one_access_002_access1"]},
							},
							{
								leafId:    nodesByLabel["l2_one_access_003_leaf1"],
								vlanId:    utils.ToPtr(722),
								accessIds: []string{nodesByLabel["l2_one_access_003_access1"]},
							},
							{
								leafId:    nodesByLabel["l2_esi_acs_dual_001_leaf_pair1"],
								vlanId:    utils.ToPtr(723),
								accessIds: []string{nodesByLabel["l2_esi_acs_dual_001_access_pair1"]},
							},
						},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_esi_acs_dual_002_leaf_pair1"],
							},
						},
					},
				},
			},
		},
		"vxlan_with_binding_start_maximal": {
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						vni:           nil,
						routingZoneId: szId,
						l3Mtu:         nil,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId:    nodesByLabel["l2_one_access_002_leaf1"],
								vlanId:    utils.ToPtr(711),
								accessIds: []string{nodesByLabel["l2_one_access_002_access1"]},
							},
							{
								leafId:    nodesByLabel["l2_one_access_003_leaf1"],
								vlanId:    utils.ToPtr(712),
								accessIds: []string{nodesByLabel["l2_one_access_003_access1"]},
							},
							{
								leafId:    nodesByLabel["l2_esi_acs_dual_001_leaf_pair1"],
								vlanId:    utils.ToPtr(713),
								accessIds: []string{nodesByLabel["l2_esi_acs_dual_001_access_pair1"]},
							},
						},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_esi_acs_dual_002_leaf_pair1"],
							},
						},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						vni:           nil,
						routingZoneId: szId,
						l3Mtu:         nil,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId:    nodesByLabel["l2_one_access_001_leaf1"],
								vlanId:    utils.ToPtr(731),
								accessIds: []string{nodesByLabel["l2_one_access_001_access1"]},
							},
							{
								leafId:    nodesByLabel["l2_esi_acs_dual_002_leaf_pair1"],
								vlanId:    utils.ToPtr(733),
								accessIds: []string{nodesByLabel["l2_esi_acs_dual_002_access_pair1"]},
							},
						},
					},
				},
			},
		},
		"start_no_description": {
			apiVersionConstraints: compatibility.VnDescriptionOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						description:   acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
			},
		},
		"start_with_description": {
			apiVersionConstraints: compatibility.VnDescriptionOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						description:   acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						description:   acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
			},
		},
		"no_bindings_reserved_vlan_id": {
			apiVersionConstraints: compatibility.VnEmptyBindingsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:    bp.Id(),
						name:           acctest.RandString(6),
						vnType:         enum.VnTypeVxlan.String(),
						routingZoneId:  szId,
						reserveVlan:    utils.ToPtr(true),
						reservedVlanId: utils.ToPtr(1100),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:    bp.Id(),
						name:           acctest.RandString(6),
						vnType:         enum.VnTypeVxlan.String(),
						routingZoneId:  szId,
						reserveVlan:    utils.ToPtr(true),
						reservedVlanId: utils.ToPtr(1101),
					},
				},
			},
		},
		"issue_1055": {
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						reserveVlan:   utils.ToPtr(true),
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_one_access_001_leaf1"],
								vlanId: utils.ToPtr(1101),
							},
						},
					},
				},
			},
		},
		"set_clear_set_tags": {
			apiVersionConstraints: compatibility.VnTagsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						tags:          randomStrings(rand.IntN(10)+1, 6),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						tags:          randomStrings(rand.IntN(10)+1, 6),
					},
				},
			},
		},
		"clear_set_clear_tags": {
			apiVersionConstraints: compatibility.VnTagsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						tags:          randomStrings(rand.IntN(10)+1, 6),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
			},
		},
		"change_tags_only": {
			apiVersionConstraints: compatibility.VnTagsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          "change_tags_only",
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						tags:          randomStrings(rand.IntN(10)+1, 6),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          "change_tags_only",
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						tags:          randomStrings(rand.IntN(10)+1, 6),
					},
				},
			},
		},
		"fixed_tags": {
			apiVersionConstraints: compatibility.VnTagsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						tags:          []string{"fixed tag one", "fixed tag two"},
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVxlan.String(),
						routingZoneId: szId,
						tags:          []string{"fixed tag one", "fixed tag two"},
					},
				},
			},
		},
		"change_rz_id": {
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: szId,
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_one_access_001_leaf1"],
							},
						},
					},
				},
				{
					preApplyResourceActionType: rzChangeResourceActionType, // either "Update" or "DestroyBeforeCreate" depending on Apstra version
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        enum.VnTypeVlan.String(),
						routingZoneId: testutils.SecurityZoneA(t, ctx, bp, true), // new routing zone
						bindings: []resourceDatacenterVirtualNetworkTemplateBinding{
							{
								leafId: nodesByLabel["l2_one_access_001_leaf1"],
							},
						},
					},
				},
			},
		},
		"issue_1114_dhcp_with_zero_bindings": {
			apiVersionConstraints: compatibility.VnEmptyBindingsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:             bp.Id(),
						name:                    acctest.RandString(6),
						vnType:                  enum.VnTypeVlan.String(),
						routingZoneId:           szId,
						dhcpEnabled:             utils.ToPtr(true),
						ipv4ConnectivityEnabled: utils.ToPtr(true),
					},
					expectNonEmptyPlan: true,
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterVirtualNetwork)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.apiVersionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config:             insecureProviderConfigHCL + config,
					Check:              resource.ComposeAggregateTestCheckFunc(checks.checks...),
					ExpectError:        step.expectError,
					ExpectNonEmptyPlan: step.expectNonEmptyPlan,
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
