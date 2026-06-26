//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/apstra-go-sdk/enum"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	dataSourceDataCenterVirtualNetworkByIdHcl = `
data "apstra_datacenter_virtual_network" "test" {
  blueprint_id = "%s"
  id           = "%s"
}
`

	dataSourceDataCenterVirtualNetworkByNameHcl = `
data "apstra_datacenter_virtual_network" "test" {
  blueprint_id = "%s"
  name         = "%s"
}
`
)

func TestDatacenterVirtualNetwork(t *testing.T) {
	ctx := context.Background()

	// create a test blueprint
	bp := testutils.BlueprintA(t, ctx)

	// create a security zone within the blueprint
	name := acctest.RandString(5)
	zoneId, err := bp.CreateSecurityZone(ctx, datacenter.SecurityZone{
		Type:    enum.SecurityZoneTypeEVPN,
		VRFName: name,
		Label:   name,
	})
	if err != nil {
		t.Fatal(err)
	}

	// grab some data we'll need when creating virtual networks
	leafIdStrings := systemIds(ctx, t, bp, "leaf")
	vnBindings := make([]datacenter.VNBinding, len(leafIdStrings))
	for i, id := range leafIdStrings {
		vnBindings[i] = datacenter.VNBinding{SystemID: id}
	}

	// specify virtual networks we want to create (and ultimately test the data source against)
	virtualNetworks := []datacenter.VirtualNetwork{
		{
			IPv4Enabled:    true,
			IPv4Subnet:     randIpNetMust(t, "10.0.0.0/16"),
			Label:          acctest.RandString(5),
			SecurityZoneID: zoneId,
			Type:           enum.VnTypeVxlan,
			Bindings:       vnBindings,
		},
		{
			IPv4Enabled:    true,
			IPv4Subnet:     randIpNetMust(t, "10.1.0.0/16"),
			Label:          acctest.RandString(5),
			SecurityZoneID: zoneId,
			Type:           enum.VnTypeVlan,
			Bindings:       []datacenter.VNBinding{{SystemID: leafIdStrings[0]}},
		},
	}

	// create the test virtual networks
	for i := range virtualNetworks {
		id, err := bp.CreateVirtualNetwork(ctx, virtualNetworks[i])
		require.NoError(t, err)
		require.NoError(t, virtualNetworks[i].SetID(id))
	}

	genTestCheckFuncs := func(vn datacenter.VirtualNetwork) []resource.TestCheckFunc {
		result := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "id", *vn.ID()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "blueprint_id", bp.Id().String()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "name", vn.Label),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "type", vn.Type.String()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "ipv4_connectivity_enabled", fmt.Sprintf("%t", vn.IPv4Enabled)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "ipv4_virtual_gateway_enabled", fmt.Sprintf("%t", vn.VirtualGatewayIPv4Enabled)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "ipv6_connectivity_enabled", fmt.Sprintf("%t", vn.IPv6Enabled)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "ipv6_virtual_gateway_enabled", fmt.Sprintf("%t", vn.VirtualGatewayIPv6Enabled)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "bindings.%", fmt.Sprintf("%d", len(vn.Bindings))),
		}
		return result
	}

	testCheckFuncsByVnId := make(map[string][]resource.TestCheckFunc, len(virtualNetworks))
	for _, virtualNetwork := range virtualNetworks {
		testCheckFuncsByVnId[*virtualNetwork.ID()] = genTestCheckFuncs(virtualNetwork)
	}

	stepsById := make([]resource.TestStep, len(virtualNetworks))
	for i, virtualNetwork := range virtualNetworks {
		stepsById[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDataCenterVirtualNetworkByIdHcl, bp.Id(), *virtualNetwork.ID()),
			Check:  resource.ComposeAggregateTestCheckFunc(testCheckFuncsByVnId[*virtualNetwork.ID()]...),
		}
	}

	stepsByName := make([]resource.TestStep, len(virtualNetworks))
	for i, virtualNetwork := range virtualNetworks {
		stepsByName[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDataCenterVirtualNetworkByNameHcl, bp.Id(), virtualNetwork.Label),
			Check:  resource.ComposeAggregateTestCheckFunc(genTestCheckFuncs(virtualNetwork)...),
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    append(stepsById, stepsByName...),
	})
}
