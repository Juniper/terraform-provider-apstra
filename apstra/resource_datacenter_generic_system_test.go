package tfapstra_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceDataCenterGenericSystemHCL = `
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
  clear_cts_on_destroy = %t
  links                = %s
}`

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

type testGenericSystemLink struct {
	tags           []string
	lagMode        apstra.RackLinkLagMode
	targetSwitchId string
	targetSwitchIf string
	targetSwitchTf int
	groupLabel     string
}

func (o testGenericSystemLink) render() string {
	return fmt.Sprintf(resourceDataCenterGenericSystemLinkHCL,
		stringSetOrNull(o.tags),
		stringOrNull(o.lagMode.String()),
		stringOrNull(o.targetSwitchId),
		stringOrNull(o.targetSwitchIf),
		o.targetSwitchTf,
		stringOrNull(o.groupLabel),
	)
}

func (o testGenericSystemLink) checks(rName string) []resource.TestCheckFunc {
	nested := map[string]string{
		"target_switch_id":              o.targetSwitchId,
		"target_switch_if_name":         o.targetSwitchIf,
		"target_switch_if_transform_id": strconv.Itoa(o.targetSwitchTf),
	}

	if o.lagMode > 0 {
		nested["lag_mode"] = o.lagMode.String()
	}

	if o.groupLabel != "" {
		nested["group_label"] = o.groupLabel
	}

	if len(o.tags) > 0 {
		nested["tags.#"] = strconv.Itoa(len(o.tags))
	}

	result := []resource.TestCheckFunc{
		resource.TestCheckTypeSetElemNestedAttrs(rName, "links.*", nested),
	}

	resource.TestCheckTypeSetElemNestedAttrs(
		"resource_name",
		"parent_object_set.*",
		map[string]string{"child_nested_strings.#": "3"},
	)

	// check that *some* link has each expected tag
	for _, tag := range o.tags {
		result = append(result, resource.TestCheckTypeSetElemAttr(rName, "links.*.tags.*", tag))
	}

	return result
}

type testGenericSystem struct {
	name              string
	hostname          string
	asn               *int
	loopback4         *net.IPNet
	loopback6         *net.IPNet
	tags              []string
	deployMode        string
	portChannelIdMin  int
	portChannelIdMax  int
	clearCtsOnDestroy bool
	links             []testGenericSystemLink
}

func (o testGenericSystem) render(bpId apstra.ObjectId, blockLabel0, blockLabel1 string) string {
	links := new(strings.Builder)
	for i, link := range o.links {
		if i == 0 {
			links.WriteString("[\n")
		}
		links.WriteString(link.render())
	}
	links.WriteString("  ]")

	return fmt.Sprintf(resourceDataCenterGenericSystemHCL,
		blockLabel0,
		blockLabel1,
		bpId,
		stringOrNull(o.name),
		stringOrNull(o.hostname),
		intPtrOrNull(o.asn),
		cidrOrNull(o.loopback4),
		cidrOrNull(o.loopback6),
		stringSetOrNull(o.tags),
		stringOrNull(o.deployMode),
		intZeroAsNull(o.portChannelIdMin),
		intZeroAsNull(o.portChannelIdMax),
		o.clearCtsOnDestroy,
		links,
	)
}

func (o testGenericSystem) checks(bpId string, blockLabels ...string) []resource.TestCheckFunc {
	rName := strings.Join(blockLabels, ".")

	result := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(rName, "id"),
		resource.TestCheckResourceAttr(rName, "blueprint_id", bpId),
		resource.TestCheckResourceAttr(rName, "port_channel_id_min", strconv.Itoa(o.portChannelIdMin)),
		resource.TestCheckResourceAttr(rName, "port_channel_id_max", strconv.Itoa(o.portChannelIdMax)),
		resource.TestCheckResourceAttr(rName, "clear_cts_on_destroy", strconv.FormatBool(o.clearCtsOnDestroy)),
		resource.TestCheckResourceAttr(rName, "links.#", strconv.Itoa(len(o.links))),
	}

	if o.name != "" {
		result = append(result, resource.TestCheckResourceAttr(rName, "name", o.name))
	} else {
		result = append(result, resource.TestCheckResourceAttrSet(rName, "name"))
	}

	if o.hostname != "" {
		result = append(result, resource.TestCheckResourceAttr(rName, "hostname", o.hostname))
	} else {
		result = append(result, resource.TestCheckResourceAttrSet(rName, "hostname"))
	}

	if o.asn != nil {
		result = append(result, resource.TestCheckResourceAttr(rName, "asn", strconv.Itoa(*o.asn)))
	}

	if o.loopback4 != nil {
		result = append(result, resource.TestCheckResourceAttr(rName, "loopback_ipv4", o.loopback4.String()))
	}

	if o.loopback6 != nil {
		result = append(result, resource.TestCheckResourceAttr(rName, "loopback_ipv6", o.loopback6.String()))
	}

	if len(o.tags) > 0 {
		result = append(result, resource.TestCheckResourceAttr(rName, "tags.#", strconv.Itoa(len(o.tags))))
		for _, tag := range o.tags {
			result = append(result, resource.TestCheckTypeSetElemAttr(rName, "tags.*", tag))
		}
	}

	if o.deployMode != "" {
		result = append(result, resource.TestCheckResourceAttr(rName, "deploy_mode", o.deployMode))
	} else {
		result = append(result, resource.TestCheckResourceAttr(rName, "deploy_mode", "deploy")) // default value
	}

	for _, link := range o.links {
		result = append(result, link.checks(rName)...)
	}

	return result
}

func TestResourceDatacenterGenericSystem_A(t *testing.T) {
	ctx := context.Background()

	bpClient := testutils.BlueprintF(t, ctx)

	err := bpClient.SetFabricSettings(ctx, &apstra.FabricSettings{Ipv6Enabled: utils.ToPtr(true)})
	require.NoError(t, err)

	// assign the leaf switch interface map
	leafIds := systemIds(ctx, t, bpClient, "leaf")
	err = bpClient.SetInterfaceMapAssignments(ctx, apstra.SystemIdToInterfaceMapAssignment{
		leafIds[0]: "Juniper_QFX5100-48T_Junos__AOS-48x10_6x40-1",
	})
	require.NoError(t, err)

	// discover the routing zones
	szs, err := bpClient.GetAllSecurityZones(ctx)
	require.NoError(t, err)
	if len(szs) == 0 {
		t.Fatalf("no security zones found")
	}

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
	_ = attachCtToPort

	type testStep struct {
		genericSystem testGenericSystem
		preConfig     func()
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"lots_of_changes": {
			versionConstraints: apiversions.Ge412, // tags not allowed in 4.1.1
			steps: []testStep{
				{
					genericSystem: testGenericSystem{
						// name:     "",
						// hostname: "",
						// tags:     []string{},
						links: []testGenericSystemLink{
							{
								// lagMode: apstra.RackLinkLagModeNone,
								// groupLabel: "",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/0",
								targetSwitchTf: 1,
								// tags:     []string{},
							},
						},
					},
				},
				{
					genericSystem: testGenericSystem{
						name:             "foo",
						hostname:         "foo.com",
						asn:              utils.ToPtr(acctest.RandIntRange(100, 199)),
						loopback4:        randIpv4NetWithPrefixLen(t, "192.0.2.0/24", 32),
						loopback6:        randIpv6NetWithPrefixLen(t, "2001:db8::/65", 128),
						tags:             []string{"a", "b"},
						portChannelIdMin: acctest.RandIntRange(100, 199),
						portChannelIdMax: acctest.RandIntRange(200, 299),
						links: []testGenericSystemLink{
							{
								lagMode:        apstra.RackLinkLagModeActive,
								groupLabel:     "foo",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/0",
								targetSwitchTf: 1,
								tags:           []string{"c", "d"},
							},
						},
					},
				},
				{
					genericSystem: testGenericSystem{
						// name:     "foo",
						// hostname: "foo.com",
						asn:              utils.ToPtr(acctest.RandIntRange(100, 199)),
						loopback4:        randIpv4NetWithPrefixLen(t, "192.0.2.0/24", 32),
						loopback6:        randIpv6NetWithPrefixLen(t, "2001:db8::/65", 128),
						portChannelIdMin: acctest.RandIntRange(100, 199),
						portChannelIdMax: acctest.RandIntRange(200, 299),
						links: []testGenericSystemLink{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/1",
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					genericSystem: testGenericSystem{
						// name:     "foo",
						// hostname: "foo.com",
						// tags:     []string{"a"},
						links: []testGenericSystemLink{
							{
								lagMode:        apstra.RackLinkLagModePassive,
								groupLabel:     "bar",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/0",
								targetSwitchTf: 1,
								tags:           []string{"c"},
							},
						},
					},
				},
				{
					genericSystem: testGenericSystem{
						// name:     "foo",
						// hostname: "foo.com",
						// tags:     []string{"a"},
						deployMode: apstra.NodeDeployModeReady.String(),
						links: []testGenericSystemLink{
							{
								lagMode:        apstra.RackLinkLagModePassive,
								groupLabel:     "bar",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/0",
								targetSwitchTf: 1,
								tags:           []string{"c"},
							},
						},
					},
				},
				{
					genericSystem: testGenericSystem{
						// name:     "foo",
						// hostname: "foo.com",
						// tags:     []string{"a"},
						deployMode: apstra.NodeDeployModeDeploy.String(),
						links: []testGenericSystemLink{
							{
								lagMode:        apstra.RackLinkLagModePassive,
								groupLabel:     "bar",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/2",
								targetSwitchTf: 1,
								tags:           []string{"c"},
							},
							{
								lagMode:        apstra.RackLinkLagModePassive,
								groupLabel:     "bar",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/3",
								targetSwitchTf: 1,
								tags:           []string{"c"},
							},
						},
					},
				},
				{
					genericSystem: testGenericSystem{
						// name:     "foo",
						// hostname: "foo.com",
						// tags:     []string{"a"},
						links: []testGenericSystemLink{
							{
								lagMode:        apstra.RackLinkLagModeStatic,
								groupLabel:     "baz",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/2",
								targetSwitchTf: 1,
								// tags:           []string{"c"},
							},
							{
								lagMode:        apstra.RackLinkLagModeStatic,
								groupLabel:     "baz",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/3",
								targetSwitchTf: 1,
								// tags:           []string{"c"},
							},
						},
					},
				},
				{
					genericSystem: testGenericSystem{
						// name:     "foo",
						// hostname: "foo.com",
						// tags:     []string{"a"},
						links: []testGenericSystemLink{
							{
								// lagMode:        apstra.RackLinkLagModeStatic,
								// groupLabel:     "baz",
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/2",
								targetSwitchTf: 1,
								// tags:           []string{"c"},
							},
						},
					},
				},
			},
		},
		"destroy_with_attached_ct": {
			steps: []testStep{
				{
					genericSystem: testGenericSystem{
						clearCtsOnDestroy: true,
						links: []testGenericSystemLink{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/4",
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					preConfig: func() {
						attachCtToPort("xe-0/0/4")
					},
					genericSystem: testGenericSystem{
						clearCtsOnDestroy: true,
						links: []testGenericSystemLink{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/4",
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
					genericSystem: testGenericSystem{
						clearCtsOnDestroy: true,
						links: []testGenericSystemLink{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/5",
								targetSwitchTf: 1,
							},
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/5",
								targetSwitchTf: 1,
							},
						},
					},
				},
				{
					preConfig: func() {
						attachCtToPort("xe-0/0/5")
					},
					genericSystem: testGenericSystem{
						clearCtsOnDestroy: true,
						links: []testGenericSystemLink{
							{
								targetSwitchId: leafIds[0],
								targetSwitchIf: "xe-0/0/6",
								targetSwitchTf: 1,
							},
						},
					},
				},
			},
		},
	}

	resourceName := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterGenericSystem)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bpClient.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {

				config := insecureProviderConfigHCL + step.genericSystem.render(bpClient.Id(), resourceName, tName)
				// t.Logf("\n// ----- begin %s step %d -----\n%s\n// -----  end  %s step %d -----\n\n", tName, i, config, tName, i)

				steps[i] = resource.TestStep{
					Config:    config,
					Check:     resource.ComposeAggregateTestCheckFunc(step.genericSystem.checks(bpClient.Id().String(), resourceName, tName)...),
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
