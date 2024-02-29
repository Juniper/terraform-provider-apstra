package tfapstra

import (
	"context"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

const (
	dataSourceDataCenterRoutingZoneHCL = `
data "apstra_datacenter_routing_zone" "test" {
  blueprint_id = "%s" 
  id = "%s"
}
`
)

func TestDataSourceDatacenterRoutingZone_A(t *testing.T) {
	ctx := context.Background()

	// BlueprintB returns a bpClient and the template from which the blueprint was created
	bpClient := testutils.BlueprintA(t, ctx)

	szId := testutils.SecurityZoneA(t, ctx, bpClient)

	sz, err := bpClient.GetSecurityZone(ctx, szId)
	require.NoError(t, err)

	err = bpClient.SetSecurityZoneDhcpServers(ctx, sz.Id, []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2")})
	require.NoError(t, err)

	rp, err := bpClient.GetDefaultRoutingPolicy(ctx)
	require.NoError(t, err)

	// generate the terraform config
	dataSourceHCL := fmt.Sprintf(dataSourceDataCenterRoutingZoneHCL, bpClient.Id(), szId)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: insecureProviderConfigHCL + dataSourceHCL,
				Check: resource.ComposeAggregateTestCheckFunc(
					[]resource.TestCheckFunc{
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_zone.test", "id", szId.String()),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),

						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_zone.test", "name", sz.Data.Label),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_zone.test", "routing_policy_id", rp.Id.String()),
						resource.TestCheckResourceAttr("data.apstra_datacenter_routing_zone.test", "dhcp_servers.#", "2"),
						resource.TestCheckTypeSetElemAttr("data.apstra_datacenter_routing_zone.test", "dhcp_servers.*", "1.1.1.1"),
						resource.TestCheckTypeSetElemAttr("data.apstra_datacenter_routing_zone.test", "dhcp_servers.*", "2.2.2.2"),
					}...,
				),
			},
		},
	})
}
