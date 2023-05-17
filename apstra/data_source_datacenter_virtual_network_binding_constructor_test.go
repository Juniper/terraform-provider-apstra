package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"regexp"
	"strings"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"testing"
)

const (
	dataSourceDataCenterVirtualNetworkBindingConstructorHCL = `
data "apstra_datacenter_virtual_network_binding_constructor" "test" {
  blueprint_id = "%s"
  switch_ids = %s
  vlan_id = %s
}
`
)

func TestAccDataSourceVirtualNetworkBindingConstructor_A(t *testing.T) {
	ctx := context.Background()
	client, deleteFunc, err := testutils.BlueprintF(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = deleteFunc(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	type node struct {
		Label string `json:"label"`
		Id    string `json:"id"`
	}
	response := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}
	err = client.Client().GetNodes(ctx, client.Id(), apstra.NodeTypeSystem, response)
	if err != nil {
		t.Fatal(err)
	}

	shortLabelToNodeId := make(map[string]string)
	populateShortLabelToNodeId := func(shortLabel string) {
		parts := strings.Split(shortLabel, "__")
		if len(parts) != 2 {
			t.Fatalf("expected 2 parts, got %d", len(parts))
		}

		var found bool
		for _, n := range response.Nodes {
			found, err = regexp.Match(fmt.Sprintf("^%s_.*_%s$", parts[0], parts[1]), []byte(n.Label))
			if err != nil {
				t.Fatal(err)
			}
			if found {
				shortLabelToNodeId[shortLabel] = n.Id
				break
			}
		}
		if !found {
			t.Fatalf("node matching short label %q not found", shortLabel)
		}
	}

	populateShortLabelToNodeId("f__001_leaf1")
	populateShortLabelToNodeId("f__001_access1")

	generateHCL := func(vlan_id string, ids ...string) string {
		hclIds := fmt.Sprintf("[\"%s\"]", strings.Join(ids, "\",\""))
		return fmt.Sprintf(dataSourceDataCenterVirtualNetworkBindingConstructorHCL, client.Id(), hclIds, vlan_id)
	}

	vlanId := "null"
	dataSourceHcl := generateHCL(vlanId,
		shortLabelToNodeId["f__001_leaf1"],
		shortLabelToNodeId["f__001_access1"],
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: insecureProviderConfigHCL + dataSourceHcl,
				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
					resource.TestCheckResourceAttr("data.apstra_datacenter_virtual_network_binding_constructor.test", "vlan_id", vlanId),
				}...),
			},
		},
	})

}
