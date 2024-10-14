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
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	versionconstraints "github.com/chrismarget-j/version-constraints"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceDatacenterVirtualNetworkTemplateHCL = `
resource %q %q {
  blueprint_id    = %q
  name            = %q
  type            = %s
  vni             = %s
  routing_zone_id = %s
  l3_mtu          = %s
  bindings        = %s
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
	blueprintId   apstra.ObjectId
	name          string
	vnType        string
	vni           *int
	routingZoneId apstra.ObjectId
	l3Mtu         *int
	bindings      []resourceDatacenterVirtualNetworkTemplateBinding
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
		stringOrNull(o.vnType),
		intPtrOrNull(o.vni),
		stringOrNull(o.routingZoneId.String()),
		intPtrOrNull(o.l3Mtu),
		bindings.String(),
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
		result.append(t, "TestCheckResourceAttr", "type", apstra.VnTypeVxlan.String())
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

	// bp, err := client.NewTwoStageL3ClosClient(ctx, "ab4468ce-c007-441c-876c-5cec6566496b")
	// szId := apstra.ObjectId("H0NOxllV2AQ2qRroxA")
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
		config resourceDatacenterVirtualNetworkTemplate
	}

	type testCase struct {
		apiVersionConstraints versionconstraints.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"no_bindings_vlan_start_minimal": {
			apiVersionConstraints: compatibility.VnEmptyBindingsOk,
			steps: []testStep{
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVlan.String(),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVlan.String(),
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
						vnType:        apstra.VnTypeVlan.String(),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVlan.String(),
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
						vnType:        apstra.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVxlan.String(),
						vni:           utils.ToPtr(rand.IntN(10000) + 5000),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVxlan.String(),
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
						vnType:        apstra.VnTypeVxlan.String(),
						vni:           utils.ToPtr(rand.IntN(10000) + 5000),
						routingZoneId: szId,
						l3Mtu:         utils.ToPtr(8800),
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVxlan.String(),
						routingZoneId: szId,
					},
				},
				{
					config: resourceDatacenterVirtualNetworkTemplate{
						blueprintId:   bp.Id(),
						name:          acctest.RandString(6),
						vnType:        apstra.VnTypeVxlan.String(),
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
						vnType:        apstra.VnTypeVlan.String(),
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
						vnType:        apstra.VnTypeVlan.String(),
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
						vnType:        apstra.VnTypeVlan.String(),
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
						vnType:        apstra.VnTypeVlan.String(),
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
						vnType:        apstra.VnTypeVlan.String(),
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
						vnType:        apstra.VnTypeVlan.String(),
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
						vnType:        apstra.VnTypeVxlan.String(),
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
						vnType:        apstra.VnTypeVxlan.String(),
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
						vnType:        apstra.VnTypeVxlan.String(),
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
						vnType:        apstra.VnTypeVxlan.String(),
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
						vnType:        apstra.VnTypeVxlan.String(),
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
						vnType:        apstra.VnTypeVxlan.String(),
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

	type bindingParams struct {
		leafId    string
		vlanId    string
		accessIds []string
	}
	type vnParams struct {
		name          string
		blueprintId   string
		vnType        string
		vni           string
		routingZoneId string
		bindings      []bindingParams
		l3Mtu         *int
	}

	params := []vnParams{
		{
			name:          acctest.RandString(10),
			blueprintId:   bp.Id().String(),
			vnType:        apstra.VnTypeVlan.String(),
			vni:           "null",
			routingZoneId: szId.String(),
			bindings: []bindingParams{
				{
					leafId: nodesByLabel["l2_one_access_001_leaf1"],
					vlanId: "null",
				},
			},
		},
		{
			name:          acctest.RandString(10),
			blueprintId:   bp.Id().String(),
			vnType:        apstra.VnTypeVlan.String(),
			vni:           "null",
			routingZoneId: szId.String(),
			bindings: []bindingParams{
				{
					leafId: nodesByLabel["l2_one_access_001_leaf1"],
					vlanId: "7",
				},
			},
		},
		{
			name:          acctest.RandString(10),
			blueprintId:   bp.Id().String(),
			vnType:        apstra.VnTypeVxlan.String(),
			vni:           "null",
			routingZoneId: szId.String(),
			bindings: []bindingParams{
				{
					leafId: nodesByLabel["l2_one_access_001_leaf1"],
					vlanId: "null",
				},
			},
		},
	}
	_ = params
}
