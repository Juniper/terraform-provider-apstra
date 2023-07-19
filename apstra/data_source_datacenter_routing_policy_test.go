package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strconv"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"testing"
)

const (
	dataSourceDataCenterRoutingPolicyByIdHCL = `
data "apstra_datacenter_routing_policy" "test" {
  blueprint_id = "%s" 
  id = "%s"
}
`
	dataSourceDataCenterRoutingPolicyByNameHCL = `
data "apstra_datacenter_routing_policy" "test" {
  blueprint_id = "%s" 
  name = "%s"
}
`
)

func TestDataSourceDatacenterRoutingPolicy_A(t *testing.T) {
	ctx := context.Background()

	// BlueprintA returns a bpClient and the template from which the blueprint was created
	bpClient, bpDelete, err := testutils.BlueprintA(ctx)
	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}
	defer func() {
		err = bpDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	rpId, rpDelete, err := testutils.RoutingPolicyA(ctx, bpClient)
	if err != nil {
		t.Fatal(errors.Join(err, rpDelete(ctx)))
	}
	defer func() {
		err = rpDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	rp, err := bpClient.GetRoutingPolicy(ctx, rpId)
	if err != nil {
		t.Fatal(err)
	}

	// generate the terraform config
	dataSourceByIdHCL := fmt.Sprintf(dataSourceDataCenterRoutingPolicyByIdHCL, bpClient.Id(), rpId)
	dataSourceByNameHCL := fmt.Sprintf(dataSourceDataCenterRoutingPolicyByNameHCL, bpClient.Id(), rp.Data.Label)

	// test check functions
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: insecureProviderConfigHCL + dataSourceByIdHCL,
				Check: resource.ComposeAggregateTestCheckFunc(
					[]resource.TestCheckFunc{
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "id", rpId.String()),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "blueprint_id", bpClient.Id().String()),
						//
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "name", rp.Data.Label),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "import_policy", rp.Data.ImportPolicy.String()),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "expect_default_ipv4", strconv.FormatBool(rp.Data.ExpectDefaultIpv4Route)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "expect_default_ipv6", strconv.FormatBool(rp.Data.ExpectDefaultIpv6Route)),
						//
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_l2_edge_subnets", strconv.FormatBool(rp.Data.ExportPolicy.L2EdgeSubnets)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_l3_edge_server_links", strconv.FormatBool(rp.Data.ExportPolicy.L3EdgeServerLinks)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_loopbacks", strconv.FormatBool(rp.Data.ExportPolicy.Loopbacks)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_spine_leaf_links", strconv.FormatBool(rp.Data.ExportPolicy.SpineLeafLinks)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_spine_superspine_links", strconv.FormatBool(rp.Data.ExportPolicy.SpineSuperspineLinks)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_static_routes", strconv.FormatBool(rp.Data.ExportPolicy.StaticRoutes)),
					}...,
				),
			},
			{
				Config: insecureProviderConfigHCL + dataSourceByNameHCL,
				Check: resource.ComposeAggregateTestCheckFunc(
					[]resource.TestCheckFunc{
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "id", rpId.String()),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "blueprint_id", bpClient.Id().String()),
						//
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "name", rp.Data.Label),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "import_policy", rp.Data.ImportPolicy.String()),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "expect_default_ipv4", strconv.FormatBool(rp.Data.ExpectDefaultIpv4Route)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "expect_default_ipv6", strconv.FormatBool(rp.Data.ExpectDefaultIpv6Route)),
						//
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_l2_edge_subnets", strconv.FormatBool(rp.Data.ExportPolicy.L2EdgeSubnets)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_l3_edge_server_links", strconv.FormatBool(rp.Data.ExportPolicy.L3EdgeServerLinks)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_loopbacks", strconv.FormatBool(rp.Data.ExportPolicy.Loopbacks)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_spine_leaf_links", strconv.FormatBool(rp.Data.ExportPolicy.SpineLeafLinks)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_spine_superspine_links", strconv.FormatBool(rp.Data.ExportPolicy.SpineSuperspineLinks)),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_policy.test", "export_policy.export_static_routes", strconv.FormatBool(rp.Data.ExportPolicy.StaticRoutes)),
					}...,
				),
			},
		},
	})
}
