package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"net"
	"strconv"
	"strings"
	"testing"
)

const (
	resourceDataCenterGenericSystemHCL = `
resource "apstra_datacenter_generic_system" "test" {
  blueprint_id        = %s
  name                = %s
  hostname            = %s
  asn                 = %s
  loopback_ipv4       = %s
  loopback_ipv6       = %s
  tags                = %s
  deploy_mode         = %s
  port_channel_id_min = %s
  port_channel_id_max = %s
  links               = [
                       %s  ]
}
`

	resourceDataCenterGenericSystemLinkHCL = `    {
      tags                          = %s
      lag_mode                      = %s
      target_switch_id              = %s
      target_switch_if_name         = %s
      target_switch_if_transform_id = %d
      group_label                   = %s
    },
`
)

func TestResourceDatacenterGenericSystem_A(t *testing.T) {
	ctx := context.Background()

	bpClient := testutils.BlueprintF(t, ctx)

	err := bpClient.SetFabricAddressingPolicy(ctx, &apstra.TwoStageL3ClosFabricAddressingPolicy{Ipv6Enabled: utils.ToPtr(true)})
	if err != nil {
		t.Fatal(err)
	}

	stringOrNull := func(in string) string {
		if in == "" {
			return "null"
		}
		return `"` + in + `"`
	}

	intOrNull := func(in *int) string {
		if in == nil {
			return "null"
		}
		return strconv.Itoa(*in)
	}

	zeroAsNull := func(in int) string {
		if in == 0 {
			return "null"
		}
		return strconv.Itoa(in)
	}

	ipOrNull := func(in *net.IPNet) string {
		if in == nil {
			return "null"
		}
		return `"` + in.String() + `"`
	}

	type tagSlice []string
	renderTags := func(in tagSlice) string {
		if len(in) == 0 {
			return "null"
		}
		return `["` + strings.Join(in, `","`) + `"]`
	}

	type link struct {
		tags           []string
		lagMode        apstra.RackLinkLagMode
		targetSwitchId string
		targetSwitchIf string
		targetSwitchTf int
		groupLabel     string
	}
	renderLink := func(in link) string {
		return fmt.Sprintf(resourceDataCenterGenericSystemLinkHCL,
			renderTags(in.tags),
			stringOrNull(in.lagMode.String()),
			stringOrNull(in.targetSwitchId),
			stringOrNull(in.targetSwitchIf),
			in.targetSwitchTf,
			stringOrNull(in.groupLabel),
		)
	}
	renderLinks := func(in []link) string {
		sb := strings.Builder{}
		for i := range in {
			sb.WriteString(renderLink(in[i]))
		}
		return sb.String()
	}

	type genericSystem struct {
		bpId             string
		name             string
		hostname         string
		asn              *int
		loopback4        *net.IPNet
		loopback6        *net.IPNet
		tags             tagSlice
		deployMode       string
		portChannelIdMin int
		portChannelIdMax int
		links            []link
	}
	renderGenericSystem := func(in genericSystem) string {
		return fmt.Sprintf(resourceDataCenterGenericSystemHCL,
			stringOrNull(bpClient.Id().String()),
			stringOrNull(in.name),
			stringOrNull(in.hostname),
			intOrNull(in.asn),
			ipOrNull(in.loopback4),
			ipOrNull(in.loopback6),
			renderTags(in.tags),
			stringOrNull(in.deployMode),
			zeroAsNull(in.portChannelIdMin),
			zeroAsNull(in.portChannelIdMax),
			renderLinks(in.links),
		)
	}
	_ = renderGenericSystem

	leafQuery := new(apstra.PathQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(bpClient.Id()).
		SetClient(bpClient.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal("leaf")},
			{Key: "name", Value: apstra.QEStringVal("n_leaf")},
		})
	var leafQueryResult struct {
		Items []struct {
			Leaf struct {
				Id string `json:"id"`
			} `json:"n_leaf"`
		} `json:"items"`
	}
	err = leafQuery.Do(ctx, &leafQueryResult)
	if err != nil {
		t.Fatal(err)
	}
	leafIds := make([]string, len(leafQueryResult.Items))
	for i, item := range leafQueryResult.Items {
		leafIds[i] = item.Leaf.Id
	}

	// assign the leaf switch interface map
	err = bpClient.SetInterfaceMapAssignments(ctx, apstra.SystemIdToInterfaceMapAssignment{
		leafIds[0]: "Juniper_QFX5100-48T_Junos__AOS-48x10_6x40-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		genericSystem genericSystem
		testCheckFunc resource.TestCheckFunc
	}

	// "A" and "B" represent first and second config state in multi-step test
	var asnA, asnB int
	var lo4A, lo4B net.IPNet
	var lo6A, lo6B net.IPNet
	var portChannelIdMinA, portChannelIdMaxA, portChannelIdMinB, portChannelIdMaxB int
	asnA = 5
	asnB = 6
	portChannelIdMinA = 1
	portChannelIdMaxA = 5
	portChannelIdMinB = 6
	portChannelIdMaxB = 10

	lo4A = net.IPNet{
		IP:   net.IP{10, 0, 0, 5},
		Mask: net.CIDRMask(32, 32),
	}
	lo4B = net.IPNet{
		IP:   net.IP{10, 0, 0, 6},
		Mask: net.CIDRMask(24, 32),
	}
	lo6A = net.IPNet{
		IP:   net.IP{0x20, 0x1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5},
		Mask: net.CIDRMask(128, 128),
	}
	lo6B = net.IPNet{
		IP:   net.IP{0x20, 0x1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6},
		Mask: net.CIDRMask(64, 128),
	}

	testCases := []testCase{
		{
			genericSystem: genericSystem{
				// name:     "",
				// hostname: "",
				// tags:     []string{},
				links: []link{
					{
						// lagMode: apstra.RackLinkLagModeNone,
						// groupLabel: "",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						// tags:     []string{},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "name"),
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "hostname"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "tags"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags"),
			}...),
		},
		{
			genericSystem: genericSystem{
				name:             "foo",
				hostname:         "foo.com",
				asn:              &asnA,
				loopback4:        &lo4A,
				loopback6:        &lo6A,
				tags:             []string{"a"},
				portChannelIdMin: portChannelIdMinA,
				portChannelIdMax: portChannelIdMaxA,
				links: []link{
					{
						lagMode:        apstra.RackLinkLagModeActive,
						groupLabel:     "foo",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						tags:           []string{"b"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "name", "foo"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "hostname", "foo.com"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "tags.0", "a"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label", "foo"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode", apstra.RackLinkLagModeActive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.0", "b"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "port_channel_id_min", fmt.Sprint(portChannelIdMinA)),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "port_channel_id_max", fmt.Sprint(portChannelIdMaxA)),
			}...),
		},
		{
			genericSystem: genericSystem{
				// name:     "foo",
				// hostname: "foo.com",
				asn:       &asnB,
				loopback4: &lo4B,
				loopback6: &lo6B,
				// tags:     []string{"a"},
				portChannelIdMin: portChannelIdMinB,
				portChannelIdMax: portChannelIdMaxB,
				links: []link{
					{
						// lagMode:        apstra.RackLinkLagModeActive,
						// groupLabel:     "foo",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,

						// tags:           []string{"b"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "name", "foo"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "hostname", "foo.com"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "asn", strconv.Itoa(asnB)),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "loopback_ipv4", lo4B.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "loopback_ipv6", lo6B.String()),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "tags"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "port_channel_id_min", fmt.Sprint(portChannelIdMinB)),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "port_channel_id_max", fmt.Sprint(portChannelIdMaxB)),
			}...),
		},
		{
			genericSystem: genericSystem{
				// name:     "foo",
				// hostname: "foo.com",
				// tags:     []string{"a"},
				links: []link{
					{
						lagMode:        apstra.RackLinkLagModePassive,
						groupLabel:     "bar",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						tags:           []string{"c"},
					},
					{
						lagMode:        apstra.RackLinkLagModePassive,
						groupLabel:     "bar",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/7",
						targetSwitchTf: 1,
						tags:           []string{"c"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "asn"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "loopback_ipv4"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "loopback_ipv6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "2"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label", "bar"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode", apstra.RackLinkLagModePassive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.0", "c"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.group_label", "bar"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.lag_mode", apstra.RackLinkLagModePassive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_name", "xe-0/0/7"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags.0", "c"),
			}...),
		},
		{
			genericSystem: genericSystem{
				// name:     "foo",
				// hostname: "foo.com",
				// tags:     []string{"a"},
				deployMode: apstra.NodeDeployModeReady.String(),
				links: []link{
					{
						lagMode:        apstra.RackLinkLagModePassive,
						groupLabel:     "bar",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						tags:           []string{"c"},
					},
					{
						lagMode:        apstra.RackLinkLagModePassive,
						groupLabel:     "bar",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/7",
						targetSwitchTf: 1,
						tags:           []string{"c"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "asn"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "loopback_ipv4"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "loopback_ipv6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "deploy_mode", apstra.NodeDeployModeReady.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "2"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label", "bar"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode", apstra.RackLinkLagModePassive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.0", "c"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.group_label", "bar"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.lag_mode", apstra.RackLinkLagModePassive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_name", "xe-0/0/7"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags.0", "c"),
			}...),
		},
		{
			genericSystem: genericSystem{
				// name:     "foo",
				// hostname: "foo.com",
				// tags:     []string{"a"},
				deployMode: apstra.NodeDeployModeDeploy.String(),
				links: []link{
					{
						lagMode:        apstra.RackLinkLagModePassive,
						groupLabel:     "bar",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						tags:           []string{"c"},
					},
					{
						lagMode:        apstra.RackLinkLagModePassive,
						groupLabel:     "bar",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/7",
						targetSwitchTf: 1,
						tags:           []string{"c"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "asn"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "loopback_ipv4"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "loopback_ipv6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "deploy_mode", apstra.NodeDeployModeDeploy.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "2"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label", "bar"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode", apstra.RackLinkLagModePassive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.0", "c"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.group_label", "bar"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.lag_mode", apstra.RackLinkLagModePassive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_name", "xe-0/0/7"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags.0", "c"),
			}...),
		},
		{
			genericSystem: genericSystem{
				// name:     "foo",
				// hostname: "foo.com",
				// tags:     []string{"a"},
				links: []link{
					{
						lagMode:        apstra.RackLinkLagModeStatic,
						groupLabel:     "baz",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						// tags:           []string{"c"},
					},
					{
						lagMode:        apstra.RackLinkLagModeStatic,
						groupLabel:     "baz",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/7",
						targetSwitchTf: 1,
						// tags:           []string{"c"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "2"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label", "baz"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode", apstra.RackLinkLagModeStatic.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.group_label", "baz"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.lag_mode", apstra.RackLinkLagModeStatic.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags"),
			}...),
		},
		{
			genericSystem: genericSystem{
				// name:     "foo",
				// hostname: "foo.com",
				// tags:     []string{"a"},
				links: []link{
					{
						// lagMode:        apstra.RackLinkLagModeStatic,
						// groupLabel:     "baz",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/8",
						targetSwitchTf: 1,
						// tags:           []string{"c"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/8"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags"),
			}...),
		},
	}

	steps := make([]resource.TestStep, len(testCases))
	for i, tc := range testCases {
		tc.genericSystem.bpId = bpClient.Id().String()
		steps[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + renderGenericSystem(tc.genericSystem),
			Check:  tc.testCheckFunc,
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}
