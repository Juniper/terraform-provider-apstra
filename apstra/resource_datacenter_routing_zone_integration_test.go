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
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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
}
`
)

type testRoutingZone struct {
	name          string
	vlan          *int
	vni           *int
	dhcpServers   []string
	routingPolicy string
	importRTs     []string
	exportRTs     []string
	irbMode       string
}

func (o testRoutingZone) render(bpId apstra.ObjectId, rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterRoutingZoneHCL,
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
	)
}

func (o testRoutingZone) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string, immutableAttributes map[string]string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// stash immutable attributes in map for reuse
	if _, ok := immutableAttributes["vrf_name"]; !ok {
		immutableAttributes["vrf_name"] = o.name
	}

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	// immutable attributes are checked using map values
	result.append(t, "TestCheckResourceAttr", "vrf_name", immutableAttributes["vrf_name"])

	if o.vlan != nil {
		result.append(t, "TestCheckResourceAttr", "vlan_id", strconv.Itoa(*o.vlan))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "vlan_id")
	}

	if o.vni != nil {
		result.append(t, "TestCheckResourceAttr", "vni", strconv.Itoa(*o.vni))
	}

	if len(o.dhcpServers) > 0 {
		result.append(t, "TestCheckResourceAttr", "dhcp_servers.#", strconv.Itoa(len(o.dhcpServers)))
		for _, dhcpServer := range o.dhcpServers {
			result.append(t, "TestCheckTypeSetElemAttr", "dhcp_servers.*", dhcpServer)
		}
	}

	if o.routingPolicy != "" {
		result.append(t, "TestCheckResourceAttr", "routing_policy_id", o.routingPolicy)
	}

	if len(o.importRTs) > 0 {
		result.append(t, "TestCheckResourceAttr", "import_route_targets.#", strconv.Itoa(len(o.importRTs)))
		for _, importRT := range o.importRTs {
			result.append(t, "TestCheckTypeSetElemAttr", "import_route_targets.*", importRT)
		}
	}

	if len(o.exportRTs) > 0 {
		result.append(t, "TestCheckResourceAttr", "export_route_targets.#", strconv.Itoa(len(o.exportRTs)))
		for _, exportRT := range o.exportRTs {
			result.append(t, "TestCheckTypeSetElemAttr", "export_route_targets.*", exportRT)
		}
	}

	if o.irbMode != "" {
		result.append(t, "TestCheckResourceAttr", "junos_evpn_irb_mode", o.irbMode)
	}

	return result
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
						vlan:          utils.ToPtr(10),
						vni:           utils.ToPtr(10010),
						dhcpServers:   randomIPs(t, rand.Intn(3)+2, "10.0.0.0/8", "2001:db8::/65"),
						routingPolicy: policyIds[rand.Intn(len(policyIds))].String(),
						importRTs:     randomRTs(t, 1, 3),
						exportRTs:     randomRTs(t, 1, 3),
					},
				},
				{
					config: testRoutingZone{
						name:          acctest.RandString(6),
						vlan:          utils.ToPtr(20),
						vni:           utils.ToPtr(10020),
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
						vlan:          utils.ToPtr(30),
						vni:           utils.ToPtr(10030),
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
						vlan:          utils.ToPtr(30),
						vni:           utils.ToPtr(10030),
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
						vlan: utils.ToPtr(rand.Intn(100) + 100),
						vni:  utils.ToPtr(rand.Intn(100) + 6200),
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
						vlan: utils.ToPtr(rand.Intn(100) + 300),
						vni:  utils.ToPtr(rand.Intn(100) + 6300),
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
						vlan: utils.ToPtr(rand.Intn(100) + 500),
						vni:  utils.ToPtr(rand.Intn(100) + 6500),
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

				// add extra checks
				for _, ec := range step.extraChecks {
					checks.append(t, ec.testFuncName, ec.testFuncArgs...)
				}

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

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
					Check:     resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
