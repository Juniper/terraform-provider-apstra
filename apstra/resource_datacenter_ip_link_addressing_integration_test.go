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
	switchIpv4AddressType  *apstra.InterfaceNumberingIpv4Type
	switchIpv4Address      *net.IPNet
	switchIpv6AddressType  *apstra.InterfaceNumberingIpv6Type
	switchIpv6Address      *net.IPNet
	genericIpv4AddressType *apstra.InterfaceNumberingIpv4Type
	genericIpv4Address     *net.IPNet
	genericIpv6AddressType *apstra.InterfaceNumberingIpv6Type
	genericIpv6Address     *net.IPNet
}

func (o resourceDataCenterIpLinkAddressing) render(rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterIpLinkAddressingHCL,
		rType, rName,
		o.blueprintId,
		o.linkId,
		stringerOrNull(o.switchIpv4AddressType),
		stringerOrNull(o.switchIpv4Address),
		stringerOrNull(o.switchIpv6AddressType),
		stringerOrNull(o.switchIpv6Address),
		stringerOrNull(o.genericIpv4AddressType),
		stringerOrNull(o.genericIpv4Address),
		stringerOrNull(o.genericIpv6AddressType),
		stringerOrNull(o.genericIpv6Address),
	)
}

func (o resourceDataCenterIpLinkAddressing) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "link_id", o.linkId.String())

	if o.switchIpv4AddressType == nil {
		result.append(t, "TestCheckResourceAttr", "switch_ipv4_address_type", "none")
	} else {
		result.append(t, "TestCheckResourceAttr", "switch_ipv4_address_type", o.switchIpv4AddressType.String())
	}
	if o.switchIpv4Address != nil {
		result.append(t, "TestCheckResourceAttr", "switch_ipv4_address", o.switchIpv4Address.String())
	}

	if o.switchIpv6AddressType == nil {
		result.append(t, "TestCheckResourceAttr", "switch_ipv6_address_type", "none")
	} else {
		result.append(t, "TestCheckResourceAttr", "switch_ipv6_address_type", o.switchIpv6AddressType.String())
	}
	if o.switchIpv6Address != nil {
		result.append(t, "TestCheckResourceAttr", "switch_ipv6_address", o.switchIpv6Address.String())
	}

	if o.genericIpv4AddressType == nil {
		result.append(t, "TestCheckResourceAttr", "generic_ipv4_address_type", "none")
	} else {
		result.append(t, "TestCheckResourceAttr", "generic_ipv4_address_type", o.genericIpv4AddressType.String())
	}
	if o.genericIpv4Address != nil {
		result.append(t, "TestCheckResourceAttr", "generic_ipv4_address", o.genericIpv4Address.String())
	}

	if o.genericIpv6AddressType == nil {
		result.append(t, "TestCheckResourceAttr", "generic_ipv6_address_type", "none")
	} else {
		result.append(t, "TestCheckResourceAttr", "generic_ipv6_address_type", o.genericIpv6AddressType.String())
	}
	if o.genericIpv6Address != nil {
		result.append(t, "TestCheckResourceAttr", "generic_ipv6_address", o.genericIpv6Address.String())
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

	testCases := map[string]testCase{
		"create_minimal": {
			steps: []testStep{
				{
					config: resourceDataCenterIpLinkAddressing{
						blueprintId: bp.Id(),
						linkId:      linkId,
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterIpLinkAddressing)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
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
