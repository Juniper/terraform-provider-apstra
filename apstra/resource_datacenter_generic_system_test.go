package tfapstra

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceDataCenterGenericSystemHCL = `
resource "apstra_datacenter_generic_system" "test" {
  blueprint_id         = %s
  name                 = %s
  hostname             = %s
  asn                  = %s
  loopback_ipv4        = %s
  loopback_ipv6        = %s
  tags                 = %s
  deploy_mode          = %s
  port_channel_id_min  = %s
  port_channel_id_max  = %s
  clear_cts_on_destroy = %t
  links                = [
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

	err := bpClient.SetFabricSettings(ctx, &apstra.FabricSettings{Ipv6Enabled: utils.ToPtr(true)})
	require.NoError(t, err)

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
		targetSwitchId apstra.ObjectId
		targetSwitchIf string
		targetSwitchTf int
		groupLabel     string
	}
	renderLink := func(in link) string {
		return fmt.Sprintf(resourceDataCenterGenericSystemLinkHCL,
			renderTags(in.tags),
			stringOrNull(in.lagMode.String()),
			stringOrNull(in.targetSwitchId.String()),
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
		bpId              string
		name              string
		hostname          string
		asn               *int
		loopback4         *net.IPNet
		loopback6         *net.IPNet
		tags              tagSlice
		deployMode        string
		portChannelIdMin  int
		portChannelIdMax  int
		clearCtsOnDestroy bool
		links             []link
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
			in.clearCtsOnDestroy,
			renderLinks(in.links),
		)
	}

	type systemNode struct {
		Label string          `json:"label"`
		Id    apstra.ObjectId `json:"id"`
		Role  string          `json:"role"`
	}
	response := struct {
		Nodes map[string]systemNode `json:"nodes"`
	}{}

	err = bpClient.Client().GetNodes(ctx, bpClient.Id(), apstra.NodeTypeSystem, &response)
	require.NoError(t, err)

	var leafIds []apstra.ObjectId
	for _, system := range response.Nodes {
		if system.Role == "leaf" {
			leafIds = append(leafIds, system.Id)
		}
	}
	if len(leafIds) == 0 {
		t.Fatal("no leafs found")
	}

	// assign the leaf switch interface map
	err = bpClient.SetInterfaceMapAssignments(ctx, apstra.SystemIdToInterfaceMapAssignment{
		leafIds[0].String(): "Juniper_QFX5100-48T_Junos__AOS-48x10_6x40-1",
	})
	require.NoError(t, err)

	// discover the routing zones
	szs, err := bpClient.GetAllSecurityZones(ctx)
	require.NoError(t, err)

	// create a connectivity template
	ct := apstra.ConnectivityTemplate{
		Label: acctest.RandString(5),
		Subpolicies: []*apstra.ConnectivityTemplatePrimitive{
			{
				Label: "",
				Attributes: &apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink{
					SecurityZone:       &szs[0].Id,
					IPv4AddressingType: apstra.CtPrimitiveIPv4AddressingTypeNumbered,
				},
			},
		},
	}
	err = ct.SetIds()
	require.NoError(t, err)

	err = ct.SetUserData()
	require.NoError(t, err)

	err = bpClient.CreateConnectivityTemplate(ctx, &ct)
	require.NoError(t, err)
	if err != nil {
		t.Fatal(err)
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

	attachCtToPort := func(portName string) {
		query := new(apstra.PathQuery).
			SetBlueprintId(bpClient.Id()).
			SetClient(bpClient.Client()).
			Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(leafIds[0])}}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeInterface.QEEAttribute(),
				{Key: "if_name", Value: apstra.QEStringVal(portName)},
				{Key: "name", Value: apstra.QEStringVal("n_interface")},
			})
		var response struct {
			Items []struct {
				Interface struct {
					Id apstra.ObjectId `json:"id"`
				} `json:"n_interface"`
			} `json:"items"`
		}
		err := query.Do(context.Background(), &response)
		require.NoError(t, err)

		err = bpClient.SetApplicationPointConnectivityTemplates(context.Background(), response.Items[0].Interface.Id, []apstra.ObjectId{*ct.Id})
		require.NoError(t, err)
	}

	type testStep struct {
		genericSystem genericSystem
		testCheckFunc resource.TestCheckFunc
		preConfig     func()
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"lots_of_changes": {
			versionConstraints: version.MustConstraints(version.NewConstraint(">" + apiversions.Apstra411)), // tags allowed in 4.1.1 only
			steps: []testStep{
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
						resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0].String()),
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
						resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0].String()),
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
						resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0].String()),
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
						deployMode: apstra.DeployModeReady.String(),
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
						resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "deploy_mode", apstra.DeployModeReady.String()),
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
						deployMode: apstra.DeployModeDeploy.String(),
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
						resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "deploy_mode", apstra.DeployModeDeploy.String()),
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
			},
		},
		"destroy_with_attached_ct": {
			steps: []testStep{
				{
					genericSystem: genericSystem{
						clearCtsOnDestroy: true,
						links: []link{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/7",
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					preConfig: func() {
						attachCtToPort("xe-0/0/7")
					},
					genericSystem: genericSystem{
						clearCtsOnDestroy: true,
						links: []link{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/7",
								targetSwitchTf: 1,
							},
						},
					},
				},
			},
		},
		"remove_link_with_attached_ct": {
			steps: []testStep{
				{
					genericSystem: genericSystem{
						clearCtsOnDestroy: true,
						links: []link{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/8",
								targetSwitchTf: 1,
							},
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/9",
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					preConfig: func() {
						attachCtToPort("xe-0/0/8")
					},
					genericSystem: genericSystem{
						clearCtsOnDestroy: true,
						links: []link{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/9",
								targetSwitchTf: 1,
							},
						},
					},
				},
			},
		},
		"exercise_deploy_mode": {
			steps: []testStep{
				{
					genericSystem: genericSystem{
						links: []link{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/10",
								targetSwitchTf: 1,
							},
						},
					},
					testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "deploy_mode", "deploy"),
					}...),
				},
				{
					genericSystem: genericSystem{
						deployMode: "not_set",
						links: []link{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/10",
								targetSwitchTf: 1,
							},
						},
					},
					testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "deploy_mode", "not_set"),
					}...),
				},
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bpClient.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				step.genericSystem.bpId = bpClient.Id().String()
				steps[i] = resource.TestStep{
					Config:    insecureProviderConfigHCL + renderGenericSystem(step.genericSystem),
					Check:     step.testCheckFunc,
					PreConfig: step.preConfig,
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
