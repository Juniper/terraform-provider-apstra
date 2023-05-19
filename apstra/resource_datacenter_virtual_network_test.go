package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strings"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"testing"
)

const (
	resourceDatacenterVirtualNetworkTemplateHCL = `
resource "apstra_datacenter_virtual_network" "test" {
  name            = "%s"
  blueprint_id    = "%s"
  type            = "%s"
  vni             = %s
  routing_zone_id = "%s"
  bindings        = {%s}
}
`
	bindingTemplateHCL = `
    %q = {
      vlan_id    = %s
      access_ids = [%s]
    },
`
)

func TestAccDatacenterVirtualNetwork_A(t *testing.T) {
	ctx := context.Background()
	bp, deleteBlueprint, err := testutils.BlueprintC(ctx)
	if err != nil {
		t.Fatal()
	}
	defer func() {
		err = deleteBlueprint(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// security zone will evaporate when blueprint is deleted
	szId, _, err := testutils.SecurityZoneA(ctx, bp)

	type node struct {
		Label string `json:"label"`
		Id    string `json:"id"`
	}

	systemNodesResponse := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}
	err = bp.GetNodes(ctx, apstra.NodeTypeSystem, systemNodesResponse)
	if err != nil {
		t.Fatal(err)
	}

	redundancyGroupNodesResponse := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}
	err = bp.GetNodes(ctx, apstra.NodeTypeRedundancyGroup, redundancyGroupNodesResponse)
	if err != nil {
		t.Fatal(err)
	}

	// l2_one_access_001_leaf1	l2_one_access_001_access1
	// l2_one_access_002_leaf1	l2_one_access_002_access1
	// l2_one_access_003_leaf1	l2_one_access_002_access1
	// l2_esi_acs_dual_001_leaf1	l2_esi_acs_dual_001_access1
	// l2_esi_acs_dual_001_leaf2	l2_esi_acs_dual_001_access2
	// l2_esi_acs_dual_002_leaf1	l2_esi_acs_dual_002_access1	l2_esi_acs_dual_002_access_pair1
	// l2_esi_acs_dual_002_leaf2	l2_esi_acs_dual_002_access2

	labelToNodeId := make(map[string]string)
	for k, v := range systemNodesResponse.Nodes {
		labelToNodeId[v.Label] = k
	}
	for k, v := range redundancyGroupNodesResponse.Nodes {
		labelToNodeId[v.Label] = k
	}

	type bindingParams struct {
		leafId    string
		vlanId    string
		accessIds []string
	}
	type blueprintParams struct {
		name          string
		blueprintId   string
		vnType        string
		vni           string
		routingZoneId string
		bindings      []bindingParams
	}

	params := []blueprintParams{
		{
			name:          acctest.RandString(10),
			blueprintId:   bp.Id().String(),
			vnType:        apstra.VnTypeVlan.String(),
			vni:           "null",
			routingZoneId: szId.String(),
			bindings: []bindingParams{
				{
					leafId: labelToNodeId["l2_one_access_001_leaf1"],
					vlanId: "null",
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
					leafId: labelToNodeId["l2_one_access_001_leaf1"],
					vlanId: "null",
				},
			},
		},
	}

	render := func(p blueprintParams) string {
		b := strings.Builder{}
		for _, binding := range p.bindings {
			quotedAccessIds := make([]string, len(binding.accessIds))
			for i := range binding.accessIds {
				quotedAccessIds[i] = fmt.Sprintf("%q", binding.accessIds[i])
			}
			b.WriteString(fmt.Sprintf(bindingTemplateHCL,
				binding.leafId, binding.vlanId, strings.Join(quotedAccessIds, ","),
			))
		}
		return fmt.Sprintf(resourceDatacenterVirtualNetworkTemplateHCL,
			p.name,
			p.blueprintId,
			p.vnType,
			p.vni,
			p.routingZoneId,
			b.String())
	}

	resourceHCL := make([]string, len(params))
	for i := range params {
		resourceHCL[i] = render(params[i])
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + resourceHCL[0],
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "name", params[0].name),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "blueprint_id", params[0].blueprintId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "type", params[0].vnType),
					resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "vni"),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "routing_zone_id", params[0].routingZoneId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "bindings.%", fmt.Sprintf("%d", len(params[0].bindings))),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + resourceHCL[0],
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "name", params[0].name),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "blueprint_id", params[0].blueprintId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "type", params[0].vnType),
					resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "vni"),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "routing_zone_id", params[0].routingZoneId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "bindings.%", fmt.Sprintf("%d", len(params[0].bindings))),
				),
			},
		},
	})

}
