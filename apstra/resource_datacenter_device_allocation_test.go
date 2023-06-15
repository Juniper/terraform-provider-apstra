package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"log"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"testing"
)

const (
	resourceDatacenterDeviceAllocationTemplateHCL = `
resource "apstra_datacenter_device_allocation" "test" {
  blueprint_id     = %q
  node_name        = %q
  interface_map_id = "Juniper_QFX5100-48S_Junos__AOS-48x10_6x40-1"
}
`
)

func TestAccDatacenterDeviceAllocation_A(t *testing.T) {
	ctx := context.Background()
	bp, deleteBlueprint, err := testutils.BlueprintF(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = deleteBlueprint(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	query := new(apstra.PathQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal("leaf")},
			{Key: "name", Value: apstra.QEStringVal("n_leaf")},
		})

	var queryResult struct {
		Items []struct {
			Leaf struct {
				Id    apstra.ObjectId `json:"id"`
				Label string          `json:"label"`
			} `json:"n_leaf"`
		} `json:"items"`
	}

	err = query.Do(ctx, &queryResult)
	if err != nil {
		t.Fatal(err)
	}
	if len(queryResult.Items) != 1 {
		t.Fatal(fmt.Errorf("query didn't find exactly one leaf: %s", query.String()))
	}

	leafName := queryResult.Items[0].Leaf.Label
	resourceHcl := insecureProviderConfigHCL +
		fmt.Sprintf(resourceDatacenterDeviceAllocationTemplateHCL, bp.Id(), leafName)
	log.Print(resourceHcl)

	//resource.Test(t, resource.TestCase{
	//	ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
	//	Steps: []resource.TestStep{
	//		// Create and Read testing
	//		{
	//			Config: insecureProviderConfigHCL + resourceHcl,
	//			Check: resource.ComposeAggregateTestCheckFunc(
	//				// Verify ID has any value set
	//				resource.TestCheckResourceAttrSet("apstra_datacenter_device_allocation.test", "id"),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "name", params[0].name),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "blueprint_id", params[0].blueprintId),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "type", params[0].vnType),
	//				resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "vni"),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "routing_zone_id", params[0].routingZoneId),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "bindings.%", fmt.Sprintf("%d", len(params[0].bindings))),
	//			),
	//		},
	//		// Update and Read testing
	//		{
	//			Config: insecureProviderConfigHCL + resourceHCL[1],
	//			Check: resource.ComposeAggregateTestCheckFunc(
	//				// Verify ID has any value set
	//				resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "id"),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "name", params[1].name),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "blueprint_id", params[1].blueprintId),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "type", params[1].vnType),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "vni", params[1].bindings[0].vlanId),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "routing_zone_id", params[1].routingZoneId),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "bindings.%", strconv.Itoa(len(params[1].bindings))),
	//			),
	//		},
	//		// Update and Read testing
	//		{
	//			Config: insecureProviderConfigHCL + resourceHCL[2],
	//			Check: resource.ComposeAggregateTestCheckFunc(
	//				// Verify ID has any value set
	//				resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "id"),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "name", params[2].name),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "blueprint_id", params[2].blueprintId),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "type", params[2].vnType),
	//				resource.TestCheckNoResourceAttr("apstra_datacenter_virtual_network.test", "vni"),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "routing_zone_id", params[2].routingZoneId),
	//				resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "bindings.%", fmt.Sprintf("%d", len(params[2].bindings))),
	//			),
	//		},
	//	},
	//})
}
