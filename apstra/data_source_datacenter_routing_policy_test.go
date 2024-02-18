package tfapstra

import (
	"context"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"strconv"
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
	bpClient := testutils.BlueprintA(t, ctx)
	rpId := testutils.RoutingPolicyA(t, ctx, bpClient)

	rp, err := bpClient.GetRoutingPolicy(ctx, rpId)
	require.NoError(t, err)

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
