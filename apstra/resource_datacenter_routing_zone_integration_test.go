//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceDataCenterRoutingZoneHCL = `resource %q %q {
  blueprint_id         = %q // required attribute
  name                 = %q // required attribute
  vlan_id              = %s
  vni                  = %s
  dhcp_servers         = %s
  routing_policy_id    = %s
  import_route_targets = %s
  export_route_targets = %s
  junos_evpn_irb_mode  = %s
  ip_addressing_type   = %s
  disable_ipv4         = %s
  tags                 = %s
}
`

	datasourceDatacenterRoutingZoneHCL = `
data %q %q {
  blueprint_id = %q
  id           = %s
  name         = %s
  vrf_name     = %s
}
`
)

type testRoutingZone struct {
	name             string
	vlan             *int
	vni              *int
	dhcpServers      []string
	routingPolicy    string
	importRTs        []string
	exportRTs        []string
	irbMode          string
	ipAddressingType string
	disableIPv4      *bool
	tags             []string
}

func (o testRoutingZone) render(bpId apstra.ObjectId, rType, rName string) string {
	resource := fmt.Sprintf(resourceDataCenterRoutingZoneHCL,
		rType, rName,
		bpId,
		o.name,

		intPtrOrNull(o.vlan),
		intPtrOrNull(o.vni),
		stringSliceOrNull(o.dhcpServers),
		stringOrNull(o.routingPolicy),
		stringSliceOrNull(o.importRTs),
		stringSliceOrNull(o.exportRTs),
		stringOrNull(o.irbMode),
		stringOrNull(o.ipAddressingType),
		boolPtrOrNull(o.disableIPv4),
		stringSliceOrNull(o.tags),
	)

	datasourceByID := fmt.Sprintf(datasourceDatacenterRoutingZoneHCL, rType, rName+"_by_id", bpId, fmt.Sprintf("%s.%s.id", rType, rName), "null", "null")
	datasourceByName := fmt.Sprintf(datasourceDatacenterRoutingZoneHCL, rType, rName+"_by_name", bpId, "null", fmt.Sprintf("%s.%s.name", rType, rName), "null")
	datasourceByVRFName := fmt.Sprintf(datasourceDatacenterRoutingZoneHCL, rType, rName+"_by_vrf_name", bpId, "null", "null", fmt.Sprintf("%s.%s.vrf_name", rType, rName))

	return resource + datasourceByID + datasourceByName + datasourceByVRFName
}

func (o testRoutingZone) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string, immutableAttributes map[string]string) []testChecks {
	resourceChecks := newTestChecks(rType + "." + rName)
	dataByIDChecks := newTestChecks("data." + rType + "." + rName + "_by_id")
	dataByNameChecks := newTestChecks("data." + rType + "." + rName + "_by_name")
	dataByVRFNameChecks := newTestChecks("data." + rType + "." + rName + "_by_vrf_name")

	// stash immutable attributes in map for reuse
	if _, ok := immutableAttributes["vrf_name"]; !ok {
		immutableAttributes["vrf_name"] = o.name
	}

	// required and computed attributes can always be checked
	resourceChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByVRFNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	resourceChecks.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	dataByIDChecks.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	dataByNameChecks.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	resourceChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "name", o.name)

	// immutable attributes are checked using map values
	resourceChecks.append(t, "TestCheckResourceAttr", "vrf_name", immutableAttributes["vrf_name"])
	dataByIDChecks.append(t, "TestCheckResourceAttr", "vrf_name", immutableAttributes["vrf_name"])
	dataByNameChecks.append(t, "TestCheckResourceAttr", "vrf_name", immutableAttributes["vrf_name"])
	dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "vrf_name", immutableAttributes["vrf_name"])

	if o.vlan != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "vlan_id", strconv.Itoa(*o.vlan))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "vlan_id", strconv.Itoa(*o.vlan))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "vlan_id", strconv.Itoa(*o.vlan))
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "vlan_id", strconv.Itoa(*o.vlan))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "vlan_id")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "vlan_id")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "vlan_id")
		dataByVRFNameChecks.append(t, "TestCheckResourceAttrSet", "vlan_id")
	}

	if o.vni != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "vni", strconv.Itoa(*o.vni))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "vni", strconv.Itoa(*o.vni))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "vni", strconv.Itoa(*o.vni))
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "vni", strconv.Itoa(*o.vni))
	}

	if len(o.dhcpServers) > 0 {
		resourceChecks.append(t, "TestCheckResourceAttr", "dhcp_servers.#", strconv.Itoa(len(o.dhcpServers)))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "dhcp_servers.#", strconv.Itoa(len(o.dhcpServers)))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "dhcp_servers.#", strconv.Itoa(len(o.dhcpServers)))
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "dhcp_servers.#", strconv.Itoa(len(o.dhcpServers)))
		for _, dhcpServer := range o.dhcpServers {
			resourceChecks.append(t, "TestCheckTypeSetElemAttr", "dhcp_servers.*", dhcpServer)
			dataByIDChecks.append(t, "TestCheckTypeSetElemAttr", "dhcp_servers.*", dhcpServer)
			dataByNameChecks.append(t, "TestCheckTypeSetElemAttr", "dhcp_servers.*", dhcpServer)
			dataByVRFNameChecks.append(t, "TestCheckTypeSetElemAttr", "dhcp_servers.*", dhcpServer)
		}
	}

	if o.routingPolicy != "" {
		resourceChecks.append(t, "TestCheckResourceAttr", "routing_policy_id", o.routingPolicy)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "routing_policy_id", o.routingPolicy)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "routing_policy_id", o.routingPolicy)
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "routing_policy_id", o.routingPolicy)
	}

	if len(o.importRTs) > 0 {
		resourceChecks.append(t, "TestCheckResourceAttr", "import_route_targets.#", strconv.Itoa(len(o.importRTs)))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "import_route_targets.#", strconv.Itoa(len(o.importRTs)))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "import_route_targets.#", strconv.Itoa(len(o.importRTs)))
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "import_route_targets.#", strconv.Itoa(len(o.importRTs)))
		for _, importRT := range o.importRTs {
			resourceChecks.append(t, "TestCheckTypeSetElemAttr", "import_route_targets.*", importRT)
			dataByIDChecks.append(t, "TestCheckTypeSetElemAttr", "import_route_targets.*", importRT)
			dataByNameChecks.append(t, "TestCheckTypeSetElemAttr", "import_route_targets.*", importRT)
			dataByVRFNameChecks.append(t, "TestCheckTypeSetElemAttr", "import_route_targets.*", importRT)
		}
	}

	if len(o.exportRTs) > 0 {
		resourceChecks.append(t, "TestCheckResourceAttr", "export_route_targets.#", strconv.Itoa(len(o.exportRTs)))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "export_route_targets.#", strconv.Itoa(len(o.exportRTs)))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "export_route_targets.#", strconv.Itoa(len(o.exportRTs)))
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "export_route_targets.#", strconv.Itoa(len(o.exportRTs)))
		for _, exportRT := range o.exportRTs {
			resourceChecks.append(t, "TestCheckTypeSetElemAttr", "export_route_targets.*", exportRT)
			dataByIDChecks.append(t, "TestCheckTypeSetElemAttr", "export_route_targets.*", exportRT)
			dataByNameChecks.append(t, "TestCheckTypeSetElemAttr", "export_route_targets.*", exportRT)
			dataByVRFNameChecks.append(t, "TestCheckTypeSetElemAttr", "export_route_targets.*", exportRT)
		}
	}

	if o.irbMode != "" {
		resourceChecks.append(t, "TestCheckResourceAttr", "junos_evpn_irb_mode", o.irbMode)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "junos_evpn_irb_mode", o.irbMode)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "junos_evpn_irb_mode", o.irbMode)
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "junos_evpn_irb_mode", o.irbMode)
	}

	if o.ipAddressingType != "" {
		resourceChecks.append(t, "TestCheckResourceAttr", "ip_addressing_type", o.ipAddressingType)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "ip_addressing_type", o.ipAddressingType)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "ip_addressing_type", o.ipAddressingType)
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "ip_addressing_type", o.ipAddressingType)
	}

	if o.disableIPv4 != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "disable_ipv4", strconv.FormatBool(*o.disableIPv4))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "disable_ipv4", strconv.FormatBool(*o.disableIPv4))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "disable_ipv4", strconv.FormatBool(*o.disableIPv4))
		dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "disable_ipv4", strconv.FormatBool(*o.disableIPv4))
	}

	resourceChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByIDChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByNameChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByVRFNameChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	for _, tag := range o.tags {
		resourceChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByIDChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByNameChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByVRFNameChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
	}

	return []testChecks{resourceChecks, dataByIDChecks, dataByNameChecks, dataByVRFNameChecks}
}

func TestResourceDatacenterRoutingZone(t *testing.T) {
	ctx := context.Background()

	// create a blueprint
	bpClient := testutils.BlueprintA(t, ctx)

	attachVniPool := func(t testing.TB, ctx context.Context, min, max uint32) {
		// create a VNI pool
		vniPool := testutils.VniPool(t, ctx, min, max, true)

		// link the VNI pool to the blueprint
		rgn := apstra.ResourceGroupNameEvpnL3Vni
		err := bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
			ResourceGroup: apstra.ResourceGroup{
				Type: rgn.Type(),
				Name: rgn,
			},
			PoolIds: []apstra.ObjectId{vniPool.Id},
		})
		require.NoError(t, err)
	}

	policyIds := make([]apstra.ObjectId, rand.Intn(3)+2)
	for i := range policyIds {
		id, err := bpClient.CreateRoutingPolicy(ctx, &apstra.DcRoutingPolicyData{
			Label:        acctest.RandString(5),
			PolicyType:   apstra.DcRoutingPolicyTypeUser,
			ImportPolicy: apstra.DcRoutingPolicyImportPolicyDefaultOnly,
		})
		require.NoError(t, err)

		policyIds[i] = id
	}

	type extraCheck struct {
		testFuncName string
		testFuncArgs []string
	}

	type testStep struct {
		config      testRoutingZone
		preConfig   func(testing.TB)
		extraChecks []extraCheck
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"create_minimal": {
			steps: []testStep{
				{
					preConfig: func(t testing.TB) { attachVniPool(t, ctx, 5000, 5099) },
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
					extraChecks: []extraCheck{
						{
							testFuncName: "TestCheckResourceAttrPair",
							testFuncArgs: []string{"name", "vrf_name"},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vni", strconv.Itoa(5000), strconv.Itoa(5099)},
						},
					},
				},
				{
					config: testRoutingZone{
						name:          acctest.RandString(6),
						vlan:          pointer.To(10),
						vni:           pointer.To(10010),
						dhcpServers:   randomIPs(t, rand.Intn(3)+2, "10.0.0.0/8", "2001:db8::/65"),
						routingPolicy: policyIds[rand.Intn(len(policyIds))].String(),
						importRTs:     randomRTs(t, 1, 3),
						exportRTs:     randomRTs(t, 1, 3),
					},
				},
				{
					config: testRoutingZone{
						name:          acctest.RandString(6),
						vlan:          pointer.To(20),
						vni:           pointer.To(10020),
						dhcpServers:   randomIPs(t, rand.Intn(3)+2, "10.0.0.0/8", "2001:db8::/65"),
						routingPolicy: policyIds[rand.Intn(len(policyIds))].String(),
						importRTs:     randomRTs(t, 1, 3),
						exportRTs:     randomRTs(t, 1, 3),
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
				},
			},
		},
		"create_maximal": {
			steps: []testStep{
				{
					config: testRoutingZone{
						name:          acctest.RandString(6),
						vlan:          pointer.To(30),
						vni:           pointer.To(10030),
						dhcpServers:   randomIPs(t, rand.Intn(3)+2, "10.0.0.0/8", "2001:db8::/65"),
						routingPolicy: policyIds[rand.Intn(len(policyIds))].String(),
						importRTs:     randomRTs(t, 1, 3),
						exportRTs:     randomRTs(t, 1, 3),
					},
					extraChecks: []extraCheck{
						{
							testFuncName: "TestCheckResourceAttrPair",
							testFuncArgs: []string{"name", "vrf_name"},
						},
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
				},
				{
					config: testRoutingZone{
						name:          acctest.RandString(6),
						vlan:          pointer.To(30),
						vni:           pointer.To(10030),
						dhcpServers:   randomIPs(t, rand.Intn(3)+2, "10.0.0.0/8", "2001:db8::/65"),
						routingPolicy: policyIds[rand.Intn(len(policyIds))].String(),
						importRTs:     randomRTs(t, 1, 3),
						exportRTs:     randomRTs(t, 1, 3),
					},
				},
			},
		},
		"prior_begin_empty": {
			steps: []testStep{
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
					preConfig: func(t testing.TB) { attachVniPool(t, ctx, 6100, 6199) },
					extraChecks: []extraCheck{
						{
							testFuncName: "TestCheckResourceAttrPair",
							testFuncArgs: []string{"name", "vrf_name"},
						},
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vni_config", "false"},
						},
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vlan_id_config", "false"},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vlan_id", strconv.Itoa(2), strconv.Itoa(100)},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vni", strconv.Itoa(6100), strconv.Itoa(6199)},
						},
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
						vlan: pointer.To(rand.Intn(100) + 100),
						vni:  pointer.To(rand.Intn(100) + 6200),
					},
					extraChecks: []extraCheck{
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vni_config", "true"},
						},
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vlan_id_config", "true"},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vlan_id", strconv.Itoa(100), strconv.Itoa(199)},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vni", strconv.Itoa(6200), strconv.Itoa(6299)},
						},
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
					extraChecks: []extraCheck{
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vni_config", "false"},
						},
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vlan_id_config", "false"},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vlan_id", strconv.Itoa(2), strconv.Itoa(100)},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vni", strconv.Itoa(6100), strconv.Itoa(6199)},
						},
					},
				},
			},
		},
		"prior_begin_populated": {
			steps: []testStep{
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
						vlan: pointer.To(rand.Intn(100) + 300),
						vni:  pointer.To(rand.Intn(100) + 6300),
					},
					extraChecks: []extraCheck{
						{
							testFuncName: "TestCheckResourceAttrPair",
							testFuncArgs: []string{"name", "vrf_name"},
						},
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vni_config", "true"},
						},
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vlan_id_config", "true"},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vlan_id", strconv.Itoa(300), strconv.Itoa(399)},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vni", strconv.Itoa(6300), strconv.Itoa(6399)},
						},
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
					preConfig: func(t testing.TB) { attachVniPool(t, ctx, 6400, 6499) },
					extraChecks: []extraCheck{
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vni_config", "false"},
						},
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vlan_id_config", "false"},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vlan_id", strconv.Itoa(2), strconv.Itoa(100)},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vni", strconv.Itoa(6400), strconv.Itoa(6499)},
						},
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
						vlan: pointer.To(rand.Intn(100) + 500),
						vni:  pointer.To(rand.Intn(100) + 6500),
					},
					extraChecks: []extraCheck{
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vni_config", "true"},
						},
						{
							testFuncName: "TestCheckResourceAttr",
							testFuncArgs: []string{"had_prior_vlan_id_config", "true"},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vlan_id", strconv.Itoa(500), strconv.Itoa(599)},
						},
						{
							testFuncName: "TestCheckResourceInt64AttrBetween",
							testFuncArgs: []string{"vni", strconv.Itoa(6500), strconv.Itoa(6599)},
						},
					},
				},
			},
		},
		"ipv4_to_ipv6_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4",
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv6",
					},
				},
			},
		},
		"ipv6_to_ipv46_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv6",
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4_ipv6",
					},
				},
			},
		},
		"ipv46_to_ipv4_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4_ipv6",
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4",
					},
				},
			},
		},
		"ipv6_to_ipv4_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv6",
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4",
					},
				},
			},
		},
		"ipv46_to_ipv6_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4_ipv6",
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv6",
					},
				},
			},
		},
		"ipv4_to_ipv46_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4",
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4_ipv6",
					},
				},
			},
		},
		"ipv4_to_ipv6_only_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4",
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv6",
						disableIPv4:      pointer.To(true),
					},
				},
			},
		},
		"ipv6_only_to_ipv46_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv6",
						disableIPv4:      pointer.To(true),
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4_ipv6",
					},
				},
			},
		},
		"ipv6_only_to_ipv4_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv6",
						disableIPv4:      pointer.To(true),
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4",
					},
				},
			},
		},
		"ipv46_to_ipv6_only_with_apstra610_or_later": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv4_ipv6",
					},
				},
				{
					config: testRoutingZone{
						name:             acctest.RandString(6),
						ipAddressingType: "ipv6",
						disableIPv4:      pointer.To(true),
					},
				},
			},
		},
		"tags_create_minimal": {
			versionConstraints: compatibility.RoutingZoneTagsOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
						tags: randomStrings(rand.Intn(5)+1, 6),
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
						tags: randomStrings(rand.Intn(5)+1, 6),
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
				},
			},
		},
		"tags_create_maximal": {
			versionConstraints: compatibility.RoutingZoneTagsOK.Constraints,
			steps: []testStep{
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
						tags: randomStrings(rand.Intn(5)+1, 6),
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
					},
				},
				{
					config: testRoutingZone{
						name: acctest.RandString(6),
						tags: randomStrings(rand.Intn(5)+1, 6),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterRoutingZone)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() don't run in parallel due to VNI pool constraint

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bpClient.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			immutableAttributes := make(map[string]string) // for attributes which are: immutable, computed, predictable
			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(bpClient.Id(), resourceType, tName)
				checks := step.config.testChecks(t, bpClient.Id(), resourceType, tName, immutableAttributes)

				var checkLog string
				var checkFuncs []resource.TestCheckFunc

				// add extra checks
				for _, ec := range step.extraChecks {
					checks[0].append(t, ec.testFuncName, ec.testFuncArgs...)
				}

				for _, checkList := range checks {
					checkLog = checkLog + checkList.string(len(checkFuncs))
					checkFuncs = append(checkFuncs, checkList.checks...)
				}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, checkLog, stepName)

				var preconfig func()
				if step.preConfig != nil {
					t, f := t, step.preConfig
					preconfig = func() {
						f(t)
					}
				}

				steps[i] = resource.TestStep{
					PreConfig: preconfig,
					Config:    insecureProviderConfigHCL + config,
					Check:     resource.ComposeAggregateTestCheckFunc(checkFuncs...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
