//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceFreeformAllocGroupHcl = `
resource %q %q {
  blueprint_id        = %q
  name                = %q
  type                = %q
  pool_ids			  = %s
 }
`
)

type resourceAllocGroup struct {
	blueprintId string
	name        string
	groupType   enum.ResourcePoolType
	poolIds     []string
}

func (o resourceAllocGroup) render(rType, rName string) string {
	return fmt.Sprintf(resourceFreeformAllocGroupHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		utils.StringersToFriendlyString(o.groupType),
		stringSliceOrNull(o.poolIds),
	)
}

func (o resourceAllocGroup) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "testCheckResourceAttr", "type", utils.StringersToFriendlyString(o.groupType))
	return result
}

func TestResourceAllocGroup(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint
	bp := testutils.FfBlueprintA(t, ctx)

	newAsnPool := func(t testing.TB) string {
		t.Helper()

		asnRange := []apstra.IntfIntRange{
			apstra.IntRangeRequest{First: 65535, Last: uint32(acctest.RandIntRange(65536, 67000))},
		}

		asnPoolId, err := bp.Client().CreateAsnPool(ctx, &apstra.AsnPoolRequest{
			DisplayName: acctest.RandString(6),
			Ranges:      asnRange,
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteAsnPool(ctx, asnPoolId)) })
		return asnPoolId.String()
	}

	newIntPool := func(t testing.TB) string {
		t.Helper()

		intRange := []apstra.IntfIntRange{
			apstra.IntRangeRequest{First: 1, Last: uint32(acctest.RandIntRange(2, 3000))},
		}

		intPoolId, err := bp.Client().CreateIntegerPool(ctx, &apstra.IntPoolRequest{
			DisplayName: acctest.RandString(6),
			Ranges:      intRange,
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteIntegerPool(ctx, intPoolId)) })
		return intPoolId.String()
	}

	newVniPool := func(t testing.TB) string {
		t.Helper()

		vniRange := []apstra.IntfIntRange{
			apstra.IntRangeRequest{First: 4100, Last: uint32(acctest.RandIntRange(4100, 10000))},
		}

		intPoolId, err := bp.Client().CreateVniPool(ctx, &apstra.VniPoolRequest{
			DisplayName: acctest.RandString(6),
			Ranges:      vniRange,
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteVniPool(ctx, intPoolId)) })
		return intPoolId.String()
	}

	newIpv4Pool := func(t testing.TB) string {
		t.Helper()

		ip4subnets := []net.IPNet{
			randomPrefix(t, "10.0.0.0/16", 24),
			randomPrefix(t, "10.1.0.0/16", 25),
			randomPrefix(t, "10.2.0.0/16", 26),
			randomPrefix(t, "10.3.0.0/16", 27),
		}

		netStrings := make([]apstra.NewIpSubnet, len(ip4subnets))
		for i := range netStrings {
			netStrings[i].Network = ip4subnets[i].String()
		}

		ipv4PoolId, err := bp.Client().CreateIp4Pool(ctx, &apstra.NewIpPoolRequest{
			DisplayName: acctest.RandString(6),
			Subnets:     netStrings,
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteIp4Pool(ctx, ipv4PoolId)) })
		return ipv4PoolId.String()
	}

	newIpv6Pool := func(t testing.TB) string {
		t.Helper()

		ip6subnets := []net.IPNet{
			randomPrefix(t, "2001:db8:0::/48", 112),
			randomPrefix(t, "2001:db8:1::/48", 112),
			randomPrefix(t, "2001:db8:2::/48", 112),
		}

		netStrings := make([]apstra.NewIpSubnet, len(ip6subnets))
		for i := range netStrings {
			netStrings[i].Network = ip6subnets[i].String()
		}

		ipv6PoolId, err := bp.Client().CreateIp6Pool(ctx, &apstra.NewIpPoolRequest{
			DisplayName: acctest.RandString(6),
			Subnets:     netStrings,
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteIp6Pool(ctx, ipv6PoolId)) })
		return ipv6PoolId.String()
	}

	namesByKey := make(map[string]string)
	nameByKey := func(key string) string {
		if name, ok := namesByKey[key]; ok {
			return name
		}
		namesByKey[key] = acctest.RandString(6)
		return namesByKey[key]
	}

	type testStep struct {
		config resourceAllocGroup
	}
	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"start_asn": {
			steps: []testStep{
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_asn"),
						groupType:   enum.ResourcePoolTypeAsn,
						poolIds:     []string{newAsnPool(t)},
					},
				},
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_asn"),
						groupType:   enum.ResourcePoolTypeAsn,
						poolIds:     []string{newAsnPool(t), newAsnPool(t)},
					},
				},
			},
		},
		"test_int": {
			steps: []testStep{
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_int"),
						groupType:   enum.ResourcePoolTypeInt,
						poolIds:     []string{newIntPool(t)},
					},
				},
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_int"),
						groupType:   enum.ResourcePoolTypeInt,
						poolIds:     []string{newIntPool(t)},
					},
				},
			},
		},
		"test_vni": {
			steps: []testStep{
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_vni"),
						groupType:   enum.ResourcePoolTypeVni,
						poolIds:     []string{newVniPool(t)},
					},
				},
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_vni"),
						groupType:   enum.ResourcePoolTypeVni,
						poolIds:     []string{newVniPool(t)},
					},
				},
			},
		},
		"test_ipv4": {
			steps: []testStep{
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_ipv4"),
						groupType:   enum.ResourcePoolTypeIpv4,
						poolIds:     []string{newIpv4Pool(t)},
					},
				},
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_ipv4"),
						groupType:   enum.ResourcePoolTypeIpv4,
						poolIds:     []string{newIpv4Pool(t)},
					},
				},
			},
		},
		"test_ipv6": {
			steps: []testStep{
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_ipv6"),
						groupType:   enum.ResourcePoolTypeIpv6,
						poolIds:     []string{newIpv6Pool(t)},
					},
				},
				{
					config: resourceAllocGroup{
						blueprintId: bp.Id().String(),
						name:        nameByKey("test_ipv6"),
						groupType:   enum.ResourcePoolTypeIpv6,
						poolIds:     []string{newIpv6Pool(t)},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformAllocGroup)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.apiVersionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------%s// -------- end config for %s ------\n\n", stepName, config, stepName)
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
}
