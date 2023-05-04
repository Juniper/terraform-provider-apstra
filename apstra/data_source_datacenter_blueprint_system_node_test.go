package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strconv"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"testing"
)

const (
	dataSourceDataCenterBlueprintSystemNodeHCL = `
data "apstra_datacenter_blueprint_system_node" "test" {
  blueprint_id = "%s" 
  id = "%s"
}
`
)

func TestDatacenterBlueprintSystemNode_A(t *testing.T) {
	ctx := context.Background()
	client, err := testutils.GetTestClient()
	if err != nil {
		t.Fatal(err)
	}

	// BlueprintB returns a bpClient and the template from which the blueprint was created
	bpClient, templateId, bpDelete, err := testutils.BlueprintB(ctx)
	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}
	defer func() {
		err := bpDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	// retrieve the origin template details because we need the spine tags
	template, err := client.GetRackBasedTemplate(ctx, templateId)
	if err != nil {
		t.Fatal(err)
	}

	// Get all of the system nodes - we'll compare this data against the dataSource
	type node struct {
		Id         string `json:"id"`
		Hostname   string `json:"hostname"`
		Label      string `json:"label"`
		Role       string `json:"role"`
		SystemId   string `json:"system_id"`
		SystemType string `json:"system_type"`
	}
	nodeResponse := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}
	err = bpClient.GetNodes(ctx, apstra.NodeTypeSystem, nodeResponse)
	if err != nil {
		t.Fatal(err)
	}

	// find spine1
	spine1 := new(node)
	for _, n := range nodeResponse.Nodes {
		if n.Label == "spine1" {
			spine1 = &n
			break
		}
	}
	if spine1 == nil {
		t.Fatalf("spine 1 not found among %d system nodes", len(nodeResponse.Nodes))
	}

	// generate the terraform config
	dataSourceHCL := fmt.Sprintf(dataSourceDataCenterBlueprintSystemNodeHCL, bpClient.Id(), spine1.Id)

	// test check functions
	testCheckFuncs := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint_system_node.test", "id", spine1.Id),
		resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint_system_node.test", "attributes.hostname", spine1.Hostname),
		resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint_system_node.test", "attributes.label", spine1.Label),
		resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint_system_node.test", "attributes.role", spine1.Role),
		resource.TestCheckNoResourceAttr("data.apstra_datacenter_blueprint_system_node.test", "attributes.system_id"),
		resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint_system_node.test", "attributes.system_type", spine1.SystemType),
		resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint_system_node.test", "attributes.tag_ids.#", strconv.Itoa(len(template.Data.Spine.Tags))),
	}
	for i := range template.Data.Spine.Tags {
		testCheckFuncs = append(testCheckFuncs, resource.TestCheckTypeSetElemAttr(
			"data.apstra_datacenter_blueprint_system_node.test",
			"attributes.tag_ids.*",
			template.Data.Spine.Tags[i].Label,
		))
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: insecureProviderConfigHCL + dataSourceHCL,
				Check:  resource.ComposeAggregateTestCheckFunc(testCheckFuncs...),
			},
		},
	})
}
