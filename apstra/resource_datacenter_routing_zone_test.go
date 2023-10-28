package tfapstra

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
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
  import_route_targets = %s
  export_route_targets = %s
}`
)

type routeTargets struct {
	rts []string
}

// init loads o.rts with random RT strings
func (o *routeTargets) init(count int) {
	s1 := make([]uint32, count)
	s2 := make([]uint32, count)
	FillWithRandomIntegers(s1)
	FillWithRandomIntegers(s2)

	o.rts = make([]string, count)
	for i := 0; i < count; i++ {
		r := rand.Intn(3)
		switch r {
		case 0: // force to 16-bits:32-bits
			o.rts[i] = fmt.Sprintf("%d:%d", uint16(s1[i]), s2[i])
		case 1: // force to 32-bits:16-bits
			o.rts[i] = fmt.Sprintf("%d:%d", s1[i], uint16(s2[i]))
		case 2: // force to IPv4:16-bits
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, s1[i])
			o.rts[i] = fmt.Sprintf("%s:%d", ip.String(), uint16(s2[i]))
		}
	}
}

func (o routeTargets) String() string {
	if len(o.rts) == 0 {
		return "null"
	}

	return fmt.Sprintf(`["%s"]`, strings.Join(o.rts, `","`))
}

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
		irts          routeTargets
		erts          routeTargets
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
			bpClient.Id(), tc.name, tc.vlan, tc.vni, dhcpServers, tc.irts.String(), tc.erts.String())
	}

	nameA := acctest.RandString(5)
	nameB := acctest.RandString(5)
	vlanA := strconv.Itoa(acctest.RandIntRange(2, 4093))
	vniA := strconv.Itoa(acctest.RandIntRange(4097, 16777213))
	dhcpServerA := "1.1.1.1"
	dhcpServerB := "2.2.2.2"
	rtsA := routeTargets{}
	rtsB := routeTargets{}

	rtsA.init(1)
	rtsB.init(6)

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
		{
			name:        nameA,
			vlan:        vlanA,
			vni:         vniA,
			dhcpServers: []string{dhcpServerA, dhcpServerB},
			irts:        rtsA,
			erts:        rtsB,
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", nameA),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", vlanA), // first available vlan ID
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", vniA),      // first available vni
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.#", "2"),
				resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", dhcpServerA),
				resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", dhcpServerB),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "import_route_targets.#", strconv.Itoa(len(rtsA.rts))),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "import_route_targets.0", rtsA.rts[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "export_route_targets.#", strconv.Itoa(len(rtsB.rts))),
			}...),
		},
		{
			name:        nameA,
			vlan:        vlanA,
			vni:         vniA,
			dhcpServers: []string{dhcpServerA, dhcpServerB},
			irts:        rtsB,
			erts:        rtsA,
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", nameA),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", vlanA), // first available vlan ID
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", vniA),      // first available vni
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.#", "2"),
				resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", dhcpServerA),
				resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", dhcpServerB),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "import_route_targets.#", strconv.Itoa(len(rtsB.rts))),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "export_route_targets.#", strconv.Itoa(len(rtsA.rts))),
				resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "export_route_targets.0", rtsA.rts[0]),
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
