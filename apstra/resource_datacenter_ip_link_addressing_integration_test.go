//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"net"
	"testing"

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
	resourceDataCenterIpLinkAddressingHCL = `resource %q %q {
  blueprint_id              = %q // required attribute
  link_id                   = %q // required attribute
  switch_ipv4_address_type  = %s
  switch_ipv4_address       = %s
  switch_ipv6_address_type  = %s
  switch_ipv6_address       = %s
  generic_ipv4_address_type = %s
  generic_ipv4_address      = %s
  generic_ipv6_address_type = %s
  generic_ipv6_address      = %s
}
`
)

type resourceDataCenterIpLinkAddressing struct {
	blueprintId            apstra.ObjectId
	linkId                 apstra.ObjectId
	switchIpv4AddressType  string
	switchIpv4Address      string
	switchIpv6AddressType  string
	switchIpv6Address      string
	genericIpv4AddressType string
	genericIpv4Address     string
	genericIpv6AddressType string
	genericIpv6Address     string
}

func (o resourceDataCenterIpLinkAddressing) render(rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterIpLinkAddressingHCL,
		rType, rName,
		o.blueprintId,
		o.linkId,
		stringOrNull(o.switchIpv4AddressType),
		stringOrNull(o.switchIpv4Address),
		stringOrNull(o.switchIpv6AddressType),
		stringOrNull(o.switchIpv6Address),
		stringOrNull(o.genericIpv4AddressType),
		stringOrNull(o.genericIpv4Address),
		stringOrNull(o.genericIpv6AddressType),
		stringOrNull(o.genericIpv6Address),
	)
}

func (o resourceDataCenterIpLinkAddressing) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "link_id", o.linkId.String())

	if o.switchIpv4AddressType == "" {
		result.append(t, "TestCheckResourceAttr", "switch_ipv4_address_type", "none")
	} else {
		result.append(t, "TestCheckResourceAttr", "switch_ipv4_address_type", o.switchIpv4AddressType)
	}
	if o.switchIpv4Address != "" {
		result.append(t, "TestCheckResourceAttr", "switch_ipv4_address", o.switchIpv4Address)
	}

	if o.switchIpv6AddressType == "" {
		result.append(t, "TestCheckResourceAttr", "switch_ipv6_address_type", "none")
	} else {
		result.append(t, "TestCheckResourceAttr", "switch_ipv6_address_type", o.switchIpv6AddressType)
	}
	if o.switchIpv6Address != "" {
		result.append(t, "TestCheckResourceAttr", "switch_ipv6_address", o.switchIpv6Address)
	}

	if o.genericIpv4AddressType == "" {
		result.append(t, "TestCheckResourceAttr", "generic_ipv4_address_type", "none")
	} else {
		result.append(t, "TestCheckResourceAttr", "generic_ipv4_address_type", o.genericIpv4AddressType)
	}
	if o.genericIpv4Address != "" {
		result.append(t, "TestCheckResourceAttr", "generic_ipv4_address", o.genericIpv4Address)
	}

	if o.genericIpv6AddressType == "" {
		result.append(t, "TestCheckResourceAttr", "generic_ipv6_address_type", "none")
	} else {
		result.append(t, "TestCheckResourceAttr", "generic_ipv6_address_type", o.genericIpv6AddressType)
	}
	if o.genericIpv6Address != "" {
		result.append(t, "TestCheckResourceAttr", "generic_ipv6_address", o.genericIpv6Address)
	}

	return result
}

func TestResourceDatacenterIpLinkAddressing(t *testing.T) {
	ctx := context.Background()

	// create a blueprint
	bp := testutils.BlueprintA(t, ctx)

	// enable ipv6
	fs, err := bp.GetFabricSettings(ctx)
	require.NoError(t, err)
	fs.Ipv6Enabled = utils.ToPtr(true)
	require.NoError(t, bp.SetFabricSettings(ctx, fs))

	// create routing zones
	rzCount := 2
	rzIds := make([]apstra.ObjectId, rzCount)
	for i := range rzIds {
		name := acctest.RandString(6)
		rzIds[i], err = bp.CreateSecurityZone(ctx, &apstra.SecurityZoneData{
			Label:   name,
			VrfName: name,
			SzType:  apstra.SecurityZoneTypeEVPN,
		})
		require.NoError(t, err)
	}

	// discover IPv4 and IPv6 pools
	ip4PoolIds, err := bp.Client().ListIp4PoolIds(ctx)
	require.NoError(t, err)
	require.Greater(t, len(ip4PoolIds), 0)
	ip6PoolIds, err := bp.Client().ListIp6PoolIds(ctx)
	require.NoError(t, err)
	require.Greater(t, len(ip6PoolIds), 0)

	// assign IPv4 and IPv6 pools to routing zones
	for _, rzId := range rzIds {
		rzId := rzId
		require.NoError(t, bp.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
			ResourceGroup: apstra.ResourceGroup{
				Type:           apstra.ResourceTypeIp4Pool,
				Name:           apstra.ResourceGroupNameToGenericLinkIpv4,
				SecurityZoneId: &rzId,
			},
			PoolIds: ip4PoolIds,
		}))
		require.NoError(t, bp.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
			ResourceGroup: apstra.ResourceGroup{
				Type:           apstra.ResourceTypeIp6Pool,
				Name:           apstra.ResourceGroupNameToGenericLinkIpv6,
				SecurityZoneId: &rzId,
			},
			PoolIds: ip6PoolIds,
		}))
	}

	// prep Connectivity Template subpolicies
	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(rzIds))
	for i, rzId := range rzIds {
		subpolicies[i] = &apstra.ConnectivityTemplatePrimitive{
			Attributes: &apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink{
				SecurityZone: utils.ToPtr(rzId),
				Tagged:       true,
				Vlan:         utils.ToPtr(apstra.Vlan(101 + i)),
			},
		}
	}

	// create Connectivity Template
	ct := apstra.ConnectivityTemplate{
		Label:       "all",
		Subpolicies: subpolicies,
	}
	require.NoError(t, ct.SetIds())
	require.NoError(t, ct.SetUserData())
	require.NoError(t, bp.CreateConnectivityTemplate(ctx, &ct))

	query := new(apstra.PathQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal(apstra.SystemTypeSwitch.String())},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_interface")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeLink.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal(apstra.SystemTypeServer.String())},
		})

	var response struct {
		Items []struct {
			Interface struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_interface"`
		} `json:"items"`
	}

	require.NoError(t, query.Do(context.Background(), &response))
	require.Greater(t, len(response.Items), 0)

	// attach CT to interface
	require.NoError(t, bp.SetApplicationPointConnectivityTemplates(ctx, response.Items[0].Interface.Id, []apstra.ObjectId{*ct.Id}))

	links, err := bp.GetAllSubinterfaceLinks(ctx)
	require.NoError(t, err)
	require.Len(t, links, rzCount)

	linkId := links[0].Id

	type testStep struct {
		config resourceDataCenterIpLinkAddressing
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	var slash31otherEnd net.IPNet
	getSlash31Endpoint := func() *net.IPNet {
		if slash31otherEnd.IP != nil {
			var result net.IPNet
			result.IP = make([]byte, len(slash31otherEnd.IP))
			result.Mask = make([]byte, len(slash31otherEnd.Mask))
			copy(result.IP, slash31otherEnd.IP)
			copy(result.Mask, slash31otherEnd.Mask)
			slash31otherEnd = net.IPNet{}
			return &result
		}

		result := randomSlash31(t, "192.0.2.0/24")
		slash31otherEnd.IP = make([]byte, len(result.IP))
		slash31otherEnd.Mask = make([]byte, len(result.Mask))
		copy(slash31otherEnd.IP, result.IP)
		copy(slash31otherEnd.Mask, result.Mask)
		slash31otherEnd.IP[3]++
		return &result
	}

	var slash127otherEnd net.IPNet
	getSlash127Endpoint := func() *net.IPNet {
		if slash127otherEnd.IP != nil {
			var result net.IPNet
			result.IP = make([]byte, len(slash127otherEnd.IP))
			result.Mask = make([]byte, len(slash127otherEnd.Mask))
			copy(result.IP, slash127otherEnd.IP)
			copy(result.Mask, slash127otherEnd.Mask)
			slash127otherEnd = net.IPNet{}
			return &result
		}

		result := randomSlash127(t, "2001:db8::/32")
		slash127otherEnd.IP = make([]byte, len(result.IP))
		slash127otherEnd.Mask = make([]byte, len(result.Mask))
		copy(slash127otherEnd.IP, result.IP)
		copy(slash127otherEnd.Mask, result.Mask)
		slash127otherEnd.IP[15]++
		return &result
	}

	testCases := map[string]testCase{
		"empty-all_numbered-empty": {
			steps: []testStep{
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId: bp.Id(),
						linkId:      linkId,
					},
				},
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId:            bp.Id(),
						linkId:                 linkId,
						switchIpv4AddressType:  "numbered",
						switchIpv4Address:      getSlash31Endpoint().String(),
						switchIpv6AddressType:  "numbered",
						switchIpv6Address:      getSlash127Endpoint().String(),
						genericIpv4AddressType: "numbered",
						genericIpv4Address:     getSlash31Endpoint().String(),
						genericIpv6AddressType: "numbered",
						genericIpv6Address:     getSlash127Endpoint().String(),
					},
				},
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId: bp.Id(),
						linkId:      linkId,
					},
				},
			},
		},
		"all_numbered-empty-all_numbered": {
			steps: []testStep{
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId:            bp.Id(),
						linkId:                 linkId,
						switchIpv4AddressType:  "numbered",
						switchIpv4Address:      getSlash31Endpoint().String(),
						switchIpv6AddressType:  "numbered",
						switchIpv6Address:      getSlash127Endpoint().String(),
						genericIpv4AddressType: "numbered",
						genericIpv4Address:     getSlash31Endpoint().String(),
						genericIpv6AddressType: "numbered",
						genericIpv6Address:     getSlash127Endpoint().String(),
					},
				},
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId: bp.Id(),
						linkId:      linkId,
					},
				},
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId:            bp.Id(),
						linkId:                 linkId,
						switchIpv4AddressType:  "numbered",
						switchIpv4Address:      getSlash31Endpoint().String(),
						switchIpv6AddressType:  "numbered",
						switchIpv6Address:      getSlash127Endpoint().String(),
						genericIpv4AddressType: "numbered",
						genericIpv4Address:     getSlash31Endpoint().String(),
						genericIpv6AddressType: "numbered",
						genericIpv6Address:     getSlash127Endpoint().String(),
					},
				},
			},
		},
		"empty-link_local-empty": {
			steps: []testStep{
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId: bp.Id(),
						linkId:      linkId,
					},
				},
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId:            bp.Id(),
						linkId:                 linkId,
						switchIpv6AddressType:  "link_local",
						genericIpv6AddressType: "link_local",
					},
				},
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId: bp.Id(),
						linkId:      linkId,
					},
				},
			},
		},
		"link_local-empty-link_local": {
			steps: []testStep{
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId:            bp.Id(),
						linkId:                 linkId,
						switchIpv6AddressType:  "link_local",
						genericIpv6AddressType: "link_local",
					},
				},
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId: bp.Id(),
						linkId:      linkId,
					},
				},
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId:            bp.Id(),
						linkId:                 linkId,
						switchIpv6AddressType:  "link_local",
						genericIpv6AddressType: "link_local",
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterIpLinkAddressing)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() // don't use parallel because test cases all work on the same link

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
