package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"math/rand"
	"strconv"
	"strings"
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
  l3_mtu          = %s
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
		t.Fatal(err)
	}
	defer func() {
		err = deleteBlueprint(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// security zone will evaporate when blueprint is deleted
	szId := testutils.SecurityZoneA(t, ctx, bp)

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
					leafId: labelToNodeId["l2_one_access_001_leaf1"],
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
					leafId: labelToNodeId["l2_one_access_001_leaf1"],
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
					leafId: labelToNodeId["l2_one_access_001_leaf1"],
					vlanId: "null",
				},
			},
		},
	}

	render := func(p vnParams) string {
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
		mtu := "null"
		if p.l3Mtu != nil {
			mtu = strconv.Itoa(*p.l3Mtu)
		}
		return fmt.Sprintf(resourceDatacenterVirtualNetworkTemplateHCL,
			p.name,
			p.blueprintId,
			p.vnType,
			p.vni,
			p.routingZoneId,
			b.String(),
			mtu)
	}

	apiVersion, err := version.NewVersion(bp.Client().ApiVersion())
	if err != nil {
		t.Fatal(err)
	}

	resourceHCL := make([]string, len(params))
	for i, paramset := range params {
		l3MtuMinVersion, _ := version.NewVersion("4.2.0")
		if apiVersion.GreaterThanOrEqual(l3MtuMinVersion) {
			l3Mtu := 1280 + (2 * rand.Intn(3969)) // 1280 - 9216 even numbers only
			paramset.l3Mtu = &l3Mtu
		}
		resourceHCL[i] = render(paramset)
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
				Config: insecureProviderConfigHCL + resourceHCL[1],
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "name", params[1].name),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "blueprint_id", params[1].blueprintId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "type", params[1].vnType),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "vni", params[1].bindings[0].vlanId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "routing_zone_id", params[1].routingZoneId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "bindings.%", strconv.Itoa(len(params[1].bindings))),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + resourceHCL[2],
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_datacenter_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "name", params[2].name),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "blueprint_id", params[2].blueprintId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "type", params[2].vnType),
					resource.TestCheckNoResourceAttr("apstra_datacenter_virtual_network.test", "vni"),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "routing_zone_id", params[2].routingZoneId),
					resource.TestCheckResourceAttr("apstra_datacenter_virtual_network.test", "bindings.%", fmt.Sprintf("%d", len(params[2].bindings))),
				),
			},
		},
	})
}
