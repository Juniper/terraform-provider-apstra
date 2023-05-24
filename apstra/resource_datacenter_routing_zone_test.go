package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strconv"
	"strings"
	"sync"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"testing"
)

const (
	resourceDataCenterRoutingZoneHCL = `
resource "apstra_datacenter_routing_zone" "test" {
  blueprint_id = "%s" 
  name         = "%s"
  vlan_id      = %s
  vni          = %s
  dhcp_servers = %s
}`
)

func TestResourceDatacenterRoutingZone_A(t *testing.T) {
	ctx := context.Background()

	// BlueprintB returns a bpClient and the template from which the blueprint was created
	bpClient, bpDelete, err := testutils.BlueprintA(ctx)
	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}
	bpWg := new(sync.WaitGroup)
	defer func() {
		err := bpDelete(ctx)
		bpWg.Done()
		if err != nil {
			t.Error(err)
		}
	}()
	bpWg.Add(1)

	vniPoolId, err := bpClient.Client().CreateVniPool(ctx, &apstra.VniPoolRequest{
		DisplayName: acctest.RandString(5),
		Ranges: []apstra.IntfIntRange{
			apstra.IntRange{
				First: 4096,
				Last:  4096,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		go func() {
			bpWg.Wait()
			err := bpClient.Client().DeleteVniPool(ctx, vniPoolId)
			if err != nil {
				t.Error(err)
			}
		}()
	}()

	rgn := apstra.ResourceGroupNameEvpnL3Vni
	err = bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: rgn.Type(),
			Name: rgn,
		},
		PoolIds: []apstra.ObjectId{vniPoolId},
	})
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		name          string
		vlan          string
		vni           string
		dhcpServers   []string
		testCheckFunc resource.TestCheckFunc
	}

	render := func(tc testCase) string {
		if tc.vlan == "" {
			tc.vlan = "null"
		}
		if tc.vni == "" {
			tc.vni = "null"
		}
		dhcpServers := "null"
		if len(tc.dhcpServers) != 0 {
			dhcpServers = fmt.Sprintf(`["%s"]`, strings.Join(tc.dhcpServers, `","`))
		}
		return insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRoutingZoneHCL,
			bpClient.Id(), tc.name, tc.vlan, tc.vni, dhcpServers)
	}

	nameA := acctest.RandString(5)
	nameB := acctest.RandString(5)
	vlanA := strconv.Itoa(acctest.RandIntRange(2, 4093))
	vniA := strconv.Itoa(acctest.RandIntRange(4097, 16777213))
	dhcpServerA := "1.1.1.1"
	dhcpServerB := "2.2.2.2"

	testCases := []testCase{
		{
			name: nameA,
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", nameA),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", "2"), // first available vlan ID
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "4096"),  // first available vni
				resource.TestCheckNoResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers"),
			}...),
		},
		{
			name: nameB,
			vlan: "1",
			vni:  "4096",
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", nameB),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", "1"), // first available vlan ID
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "4096"),  // first available vni
				resource.TestCheckNoResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers"),
			}...),
		},
		{
			name:        nameA,
			vlan:        vlanA,
			vni:         vniA,
			dhcpServers: []string{dhcpServerA, dhcpServerB},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", nameA),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", vlanA), // first available vlan ID
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", vniA),      // first available vni
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.#", "2"),
				resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", dhcpServerA),
				resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", dhcpServerB),
			}...),
		},
	}

	steps := make([]resource.TestStep, len(testCases))
	for i, tc := range testCases {
		steps[i] = resource.TestStep{
			Config: render(tc),
			Check:  tc.testCheckFunc,
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}
