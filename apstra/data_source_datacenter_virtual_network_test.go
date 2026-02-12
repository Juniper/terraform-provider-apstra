//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	zoneId, err := bp.CreateSecurityZone(ctx, apstra.SecurityZone{
		Type:    enum.SecurityZoneTypeEVPN,
		VRFName: name,
		Label:   name,
	})
	if err != nil {
		t.Fatal(err)
	}

	// grab some data we'll need when creating virtual networks
	leafIdStrings := systemIds(ctx, t, bp, "leaf")
	vnBindings := make([]apstra.VnBinding, len(leafIdStrings))
	for i, id := range leafIdStrings {
		vnBindings[i] = apstra.VnBinding{SystemId: apstra.ObjectId(id)}
	}

	// specify virtual networks we want to create (and ultimately test the data source against)
	virtualNetworks := []apstra.VirtualNetwork{
		{
			Data: &apstra.VirtualNetworkData{
				Ipv4Enabled:    true,
				Ipv4Subnet:     randIpNetMust(t, "10.0.0.0/16"),
				Label:          acctest.RandString(5),
				SecurityZoneId: apstra.ObjectId(zoneId),
				VnType:         enum.VnTypeVxlan,
				VnBindings:     vnBindings,
			},
		},
		{
			Data: &apstra.VirtualNetworkData{
				Ipv4Enabled:    true,
				Ipv4Subnet:     randIpNetMust(t, "10.1.0.0/16"),
				Label:          acctest.RandString(5),
				SecurityZoneId: apstra.ObjectId(zoneId),
				VnType:         enum.VnTypeVlan,
				VnBindings:     []apstra.VnBinding{{SystemId: apstra.ObjectId(leafIdStrings[0])}},
			},
		},
	}

	// create the test virtual networks
	for i := range virtualNetworks {
		virtualNetworks[i].Id, err = bp.CreateVirtualNetwork(ctx, virtualNetworks[i].Data)
		if err != nil {
			t.Fatal(err)
		}
	}

	genTestCheckFuncs := func(vn apstra.VirtualNetwork) []resource.TestCheckFunc {
		result := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "id", vn.Id.String()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "blueprint_id", bp.Id().String()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "name", vn.Data.Label),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "type", vn.Data.VnType.String()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "ipv4_connectivity_enabled", fmt.Sprintf("%t", vn.Data.Ipv4Enabled)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "ipv4_virtual_gateway_enabled", fmt.Sprintf("%t", vn.Data.VirtualGatewayIpv4Enabled)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "ipv6_connectivity_enabled", fmt.Sprintf("%t", vn.Data.Ipv6Enabled)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "ipv6_virtual_gateway_enabled", fmt.Sprintf("%t", vn.Data.VirtualGatewayIpv6Enabled)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network.test", "bindings.%", fmt.Sprintf("%d", len(vn.Data.VnBindings))),
		}
		return result
	}

	testCheckFuncsByVnId := make(map[apstra.ObjectId][]resource.TestCheckFunc, len(virtualNetworks))
	for _, virtualNetwork := range virtualNetworks {
		testCheckFuncsByVnId[virtualNetwork.Id] = genTestCheckFuncs(virtualNetwork)
	}

	stepsById := make([]resource.TestStep, len(virtualNetworks))
	for i, virtualNetwork := range virtualNetworks {
		stepsById[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDataCenterVirtualNetworkByIdHcl, bp.Id(), virtualNetwork.Id),
			Check:  resource.ComposeAggregateTestCheckFunc(testCheckFuncsByVnId[virtualNetwork.Id]...),
		}
	}

	stepsByName := make([]resource.TestStep, len(virtualNetworks))
	for i, virtualNetwork := range virtualNetworks {
		stepsByName[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDataCenterVirtualNetworkByNameHcl, bp.Id(), virtualNetwork.Data.Label),
			Check:  resource.ComposeAggregateTestCheckFunc(genTestCheckFuncs(virtualNetwork)...),
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    append(stepsById, stepsByName...),
	})
}
