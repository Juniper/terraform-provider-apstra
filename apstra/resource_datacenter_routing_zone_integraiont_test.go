//go:build integration

package tfapstra

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"
)

const (
	resourceDataCenterRoutingZoneHCL = `
resource "apstra_datacenter_routing_zone" "test" {
  blueprint_id         = "%s" 
  name                 = "%s"
  vlan_id              = %s
  vni                  = %s
  dhcp_servers         = %s
  routing_policy_id    = %s
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

func TestResourceDatacenterRoutingZone(t *testing.T) {
	ctx := context.Background()

	vniPool := testutils.VniPool(t, ctx, 4096, 4096)
	bpClient := testutils.BlueprintA(t, ctx)

	rgn := apstra.ResourceGroupNameEvpnL3Vni
	err := bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: rgn.Type(),
			Name: rgn,
		},
		PoolIds: []apstra.ObjectId{vniPool.Id},
	})
	require.NoError(t, err)

	policyId1, err := bpClient.CreateRoutingPolicy(ctx, &apstra.DcRoutingPolicyData{
		Label:        acctest.RandString(5),
		PolicyType:   apstra.DcRoutingPolicyTypeUser,
		ImportPolicy: apstra.DcRoutingPolicyImportPolicyDefaultOnly,
	})
	require.NoError(t, err)

	policyId2, err := bpClient.CreateRoutingPolicy(ctx, &apstra.DcRoutingPolicyData{
		Label:        acctest.RandString(5),
		PolicyType:   apstra.DcRoutingPolicyTypeUser,
		ImportPolicy: apstra.DcRoutingPolicyImportPolicyDefaultOnly,
	})
	require.NoError(t, err)

	rtsA := routeTargets{}
	rtsA.init(1)
	rtsB := routeTargets{}
	rtsB.init(2)

	type config struct {
		name          string
		vlan          *int
		vni           *int
		dhcpServers   []string
		routingPolicy string
		importRTs     routeTargets
		exportRTs     routeTargets
	}

	type step struct {
		config config
		checks []resource.TestCheckFunc
	}

	type testCase struct {
		steps []step
	}

	render := func(cfg config) string {
		vlan := "null"
		if cfg.vlan != nil {
			vlan = fmt.Sprintf("%d", *cfg.vlan)
		}

		vni := "null"
		if cfg.vni != nil {
			vni = fmt.Sprintf("%d", *cfg.vni)
		}

		dhcpServers := "null"
		if len(cfg.dhcpServers) > 0 {
			dhcpServers = `["` + strings.Join(cfg.dhcpServers, `", "`) + `"]`
		}

		routingPolicy := "null"
		if len(cfg.routingPolicy) != 0 {
			routingPolicy = fmt.Sprintf("%q", cfg.routingPolicy)
		}

		return fmt.Sprintf(resourceDataCenterRoutingZoneHCL,
			bpClient.Id(),
			cfg.name,
			vlan,
			vni,
			dhcpServers,
			routingPolicy,
			cfg.importRTs.String(),
			cfg.exportRTs.String(),
		)
	}

	testCases := map[string]testCase{
		"create_minimal": {
			steps: []step{
				{
					config: config{
						name: "a1",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", "a1"),
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "vlan_id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "4096"),
						resource.TestCheckNoResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers"),
					},
				},
				{
					config: config{
						name:          "a2",
						vlan:          utils.ToPtr(10),
						vni:           utils.ToPtr(10010),
						dhcpServers:   []string{"1.1.1.1", "2.2.2.2"},
						routingPolicy: policyId1.String(),
						importRTs:     rtsA,
						exportRTs:     rtsB,
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", "a2"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", "10"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "10010"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.#", "2"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", "1.1.1.1"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", "2.2.2.2"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "routing_policy_id", policyId1.String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "import_route_targets.#", strconv.Itoa(len(rtsA.rts))),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "import_route_targets.*", rtsA.rts[0]),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "export_route_targets.#", strconv.Itoa(len(rtsB.rts))),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "export_route_targets.*", rtsB.rts[0]),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "export_route_targets.*", rtsB.rts[1]),
					},
				},
				{
					config: config{
						name:          "a3",
						vlan:          utils.ToPtr(20),
						vni:           utils.ToPtr(10020),
						dhcpServers:   []string{"3.3.3.3", "4.4.4.4", "5.5.5.5"},
						routingPolicy: policyId2.String(),
						importRTs:     rtsB,
						exportRTs:     rtsA,
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", "a3"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", "20"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "10020"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.#", "3"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", "3.3.3.3"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", "4.4.4.4"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", "5.5.5.5"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "import_route_targets.*", rtsB.rts[0]),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "import_route_targets.*", rtsB.rts[1]),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "routing_policy_id", policyId2.String()),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "export_route_targets.*", rtsA.rts[0]),
					},
				},
				{
					config: config{
						name: "a4",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", "a4"),
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "vlan_id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "4096"),
						resource.TestCheckNoResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers"),
					},
				},
			},
		},
		"create_maximal": {
			steps: []step{
				{
					config: config{
						name:          "b1",
						vlan:          utils.ToPtr(30),
						vni:           utils.ToPtr(10030),
						dhcpServers:   []string{"6.6.6.6", "7.7.7.7"},
						routingPolicy: policyId1.String(),
						importRTs:     rtsA,
						exportRTs:     rtsB,
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", "b1"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", "30"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "10030"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.#", "2"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", "6.6.6.6"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", "7.7.7.7"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "routing_policy_id", policyId1.String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "import_route_targets.#", strconv.Itoa(len(rtsA.rts))),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "import_route_targets.*", rtsA.rts[0]),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "export_route_targets.#", strconv.Itoa(len(rtsB.rts))),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "export_route_targets.*", rtsB.rts[0]),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "export_route_targets.*", rtsB.rts[1]),
					},
				},
				{
					config: config{
						name: "b2",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", "b2"),
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "vlan_id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "4096"),
						resource.TestCheckNoResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers"),
					},
				},
				{
					config: config{
						name:          "b3",
						vlan:          utils.ToPtr(20),
						vni:           utils.ToPtr(20010),
						dhcpServers:   []string{"8.8.8.8"},
						routingPolicy: policyId2.String(),
						importRTs:     rtsB,
						exportRTs:     rtsA,
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_routing_zone.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "name", "b3"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vlan_id", "20"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "vni", "20010"),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.#", "1"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "dhcp_servers.*", "8.8.8.8"),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "import_route_targets.*", rtsB.rts[0]),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "import_route_targets.*", rtsB.rts[1]),
						resource.TestCheckResourceAttr("apstra_datacenter_routing_zone.test", "routing_policy_id", policyId2.String()),
						resource.TestCheckTypeSetElemAttr("apstra_datacenter_routing_zone.test", "export_route_targets.*", rtsA.rts[0]),
					},
				},
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() don't run in parallel due to VNI pool constraint

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + render(step.config),
					Check:  resource.ComposeAggregateTestCheckFunc(step.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
