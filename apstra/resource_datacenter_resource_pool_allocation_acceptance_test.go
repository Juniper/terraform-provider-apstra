//go:build integration

package tfapstra_test

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	dctestobj "github.com/Juniper/terraform-provider-apstra/internal/test_utils/datacenter_test_objects"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	resourcetestobj "github.com/Juniper/terraform-provider-apstra/internal/test_utils/resource_test_objects"
	versionconstraints "github.com/chrismarget-j/version-constraints"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceDatacenterResourcePoolAllocationHCL = `resource %q %q {
  blueprint_id    = %q // required attribute
  role            = %q // required attribute
  pool_ids        = %s // required attribute
  routing_zone_id = %s // optional attribute
}
`
)

type resourceDatacenterResourcePoolAllocation struct {
	role          enum.ResourceGroup
	poolIDs       []string
	routingZoneID string
}

func (o resourceDatacenterResourcePoolAllocation) render(rType, rName, bpID string) string {
	return fmt.Sprintf(resourceDatacenterResourcePoolAllocationHCL, rType, rName,
		bpID,
		rosetta.StringersToFriendlyString(o.role),
		stringSliceOrNull(o.poolIDs),
		stringOrNull(o.routingZoneID),
	)
}

func (o resourceDatacenterResourcePoolAllocation) testChecks(t testing.TB, rType, rName, bpID string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpID)
	result.append(t, "TestCheckResourceAttr", "role", rosetta.StringersToFriendlyString(o.role))
	result.append(t, "TestCheckResourceAttr", "pool_ids.#", strconv.Itoa(len(o.poolIDs)))
	for _, poolID := range o.poolIDs {
		result.append(t, "TestCheckTypeSetElemAttr", "pool_ids.*", poolID)
	}

	if o.routingZoneID == "" {
		result.append(t, "TestCheckNoResourceAttr", "routing_zone_id")
	} else {
		result.append(t, "TestCheckResourceAttr", "routing_zone_id", o.routingZoneID)
	}

	return result
}

func TestAccDatacenterResourcePoolAllocation(t *testing.T) {
	rzCount := 3
	asnPoolCount := 3
	ipv4PoolCount := 3
	ipv6PoolCount := 3
	vniPoolCount := 3

	ctx := t.Context()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	asnPoolIDs := make([]string, asnPoolCount)
	ipv4PoolIDs := make([]string, ipv4PoolCount)
	ipv6PoolIDs := make([]string, ipv6PoolCount)
	vniPoolIDs := make([]string, vniPoolCount)

	wait := new(sync.WaitGroup)
	wait.Add(1)
	go func() {
		for i := range asnPoolIDs {
			asnPoolIDs[i] = resourcetestobj.RandomASNPool(t, ctx, client, 100+(i*100), 199+(i*100))
		}
		for i := range ipv4PoolIDs {
			ipv4PoolIDs[i] = resourcetestobj.RandomIPv4Pool(t, ctx, client, fmt.Sprintf("10.%d.0.0/16", i), 24, 3)
		}
		for i := range ipv6PoolIDs {
			ipv6PoolIDs[i] = resourcetestobj.RandomIPv6Pool(t, ctx, client, fmt.Sprintf("fc00:0:%d::/48", i), 64, 3)
		}
		for i := range vniPoolIDs {
			vniPoolIDs[i] = resourcetestobj.RandomVNIPool(t, ctx, client, 100000+(i*100000), 199999+(i*100000))
		}
		wait.Done()
	}()

	bp := dctestobj.BlueprintA(t, ctx, client)

	rzIDs := make([]string, rzCount)
	for i := range rzIDs {
		rzIDs[i] = dctestobj.RoutingZoneA(t, ctx, bp, false)
		_ = dctestobj.VirtualNetworkB(t, ctx, bp, rzIDs[i])
	}

	wait.Wait()

	type testCase struct {
		apiVersionConstraints []versionconstraints.Constraints
		steps                 []resourceDatacenterResourcePoolAllocation
	}

	testCases := map[string]testCase{
		"leaf_asn": {
			steps: []resourceDatacenterResourcePoolAllocation{
				{
					role:    enum.ResourceGroupLeafASN,
					poolIDs: random.SomeOf(asnPoolIDs, 1, uint16(len(asnPoolIDs))),
				},
				{
					role:    enum.ResourceGroupLeafASN,
					poolIDs: random.SomeOf(asnPoolIDs, 1, uint16(len(asnPoolIDs))),
				},
			},
		},
		"leaf_ipv4": {
			steps: []resourceDatacenterResourcePoolAllocation{
				{
					role:    enum.ResourceGroupLeafIPv4,
					poolIDs: random.SomeOf(ipv4PoolIDs, 1, uint16(len(ipv4PoolIDs))),
				},
				{
					role:    enum.ResourceGroupLeafIPv4,
					poolIDs: random.SomeOf(ipv4PoolIDs, 1, uint16(len(ipv4PoolIDs))),
				},
			},
		},
		"leaf_ipv6": {
			steps: []resourceDatacenterResourcePoolAllocation{
				{
					role:    enum.ResourceGroupLeafIPv6,
					poolIDs: random.SomeOf(ipv6PoolIDs, 1, uint16(len(ipv6PoolIDs))),
				},
				{
					role:    enum.ResourceGroupLeafIPv6,
					poolIDs: random.SomeOf(ipv6PoolIDs, 1, uint16(len(ipv6PoolIDs))),
				},
			},
		},
		"vrf_leaf_ipv4": { // same RZ ID in both cases, but different pool IDs
			steps: []resourceDatacenterResourcePoolAllocation{
				{
					role:          enum.ResourceGroupLeafIPv4,
					routingZoneID: rzIDs[random.PersistentIntn("vrf_leaf_ipv4", len(rzIDs))],
					poolIDs:       random.SomeOf(ipv4PoolIDs, 1, uint16(len(ipv4PoolIDs))),
				},
				{
					role:          enum.ResourceGroupLeafIPv4,
					routingZoneID: rzIDs[random.PersistentIntn("vrf_leaf_ipv4", len(rzIDs))],
					poolIDs:       random.SomeOf(ipv4PoolIDs, 1, uint16(len(ipv4PoolIDs))),
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterResourcePoolAllocation)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() // Not using parallelism because the test cases may use the same resource groups.

			for _, constraint := range tCase.apiVersionConstraints {
				if !constraint.Check(apiVersion) {
					t.Skipf("test case %s requires Apstra %s", tName, constraint)
				}
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.render(resourceType, tName, bp.Id().String())
				checks := step.testChecks(t, resourceType, tName, bp.Id().String())

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
	log.Println("done")
}
