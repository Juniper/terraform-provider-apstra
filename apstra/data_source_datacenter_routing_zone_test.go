package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"net"
	testutils "terraform-provider-apstra/apstra/test_utils"
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

	szId, szDelete, err := testutils.SecurityZoneA(ctx, bpClient)
	if err != nil {
		t.Fatal(errors.Join(err, szDelete(ctx)))
	}
	defer func() {
		err = szDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	sz, err := bpClient.GetSecurityZone(ctx, szId)
	if err != nil {
		t.Fatal(err)
	}

	ip1 := net.ParseIP("1.1.1.1")
	ip2 := net.ParseIP("2.2.2.2")
	bpClient.SetSecurityZoneDhcpServers(ctx, sz.Id, []net.IP{ip1, ip2})

	rp, err := bpClient.GetDefaultRoutingPolicy(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// generate the terraform config
	dataSourceHCL := fmt.Sprintf(dataSourceDataCenterRoutingZoneHCL, bpClient.Id(), szId)

	// test check functions

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
