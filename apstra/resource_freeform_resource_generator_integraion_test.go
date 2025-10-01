//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
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
	resourceFreeformResourceGeneratorHcl = `
resource %q %q {
  blueprint_id      = %q
  name              = %q
  type              = %q
  scope             = %q
  allocated_from    = %s
  container_id      = %q
  subnet_prefix_len = %s
 }
`
)

type resourceFreeformResourceGenerator struct {
	blueprintId     string
	name            string
	resourceType    enum.FFResourceType
	scope           string
	allocatedFrom   string
	containerId     string
	subnetPrefixLen *int
}

func (o resourceFreeformResourceGenerator) render(rType, rName string) string {
	return fmt.Sprintf(resourceFreeformResourceGeneratorHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		utils.StringersToFriendlyString(o.resourceType),
		o.scope,
		stringOrNull(o.allocatedFrom),
		o.containerId,
		intPtrOrNull(o.subnetPrefixLen),
	)
}

func (o resourceFreeformResourceGenerator) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "scope", o.scope)
	result.append(t, "TestCheckResourceAttr", "container_id", o.containerId)
	result.append(t, "TestCheckResourceAttr", "type", utils.StringersToFriendlyString(o.resourceType))
	if o.allocatedFrom == "" {
		result.append(t, "TestCheckNoResourceAttr", "allocated_from")
	} else {
		result.append(t, "TestCheckResourceAttr", "allocated_from", o.allocatedFrom)
	}
	if o.subnetPrefixLen == nil {
		result.append(t, "TestCheckNoResourceAttr", "subnet_prefix_len")
	} else {
		result.append(t, "TestCheckResourceAttr", "subnet_prefix_len", strconv.Itoa(*o.subnetPrefixLen))
	}

	return result
}

func TestResourceFreeformResourceGenerator(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint and a group...
	bp, groupId := testutils.FfBlueprintC(t, ctx)

	newIpv4AllocationGroup := func(t testing.TB) apstra.ObjectId {
		t.Helper()

		// create an ipv4 pool
		randomNet := net.IPNet{
			IP:   randIpvAddressMust(t, "10.0.0.0/8"),
			Mask: net.CIDRMask(24, 32),
		}
		ipv4poolId, err := bp.Client().CreateIp4Pool(ctx, &apstra.NewIpPoolRequest{
			DisplayName: acctest.RandString(6),
			Subnets:     []apstra.NewIpSubnet{{Network: randomNet.String()}},
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteIp4Pool(ctx, ipv4poolId)) })

		// now create the allocation group
		allocGroupCfg := apstra.FreeformAllocGroupData{
			Name:    "ipv4AlGr-" + acctest.RandString(6),
			Type:    enum.ResourcePoolTypeIpv4,
			PoolIds: []apstra.ObjectId{ipv4poolId},
		}
		allocGroup, err := bp.CreateAllocGroup(ctx, utils.ToPtr(allocGroupCfg))
		require.NoError(t, err)
		return allocGroup
	}

	newIpv6AllocationGroup := func(t testing.TB) apstra.ObjectId {
		t.Helper()

		// create an ipv6 pool
		randomNet := net.IPNet{
			IP:   randIpvAddressMust(t, "2002::1234:abcd:ffff:c0a8:101/64"),
			Mask: net.CIDRMask(64, 128),
		}
		ipv6poolId, err := bp.Client().CreateIp6Pool(ctx, &apstra.NewIpPoolRequest{
			DisplayName: acctest.RandString(6),
			Subnets:     []apstra.NewIpSubnet{{Network: randomNet.String()}},
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteIp6Pool(ctx, ipv6poolId)) })

		// now create the allocation group
		allocGroupCfg := apstra.FreeformAllocGroupData{
			Name:    "ipv6AlGr-" + acctest.RandString(6),
			Type:    enum.ResourcePoolTypeIpv6,
			PoolIds: []apstra.ObjectId{ipv6poolId},
		}
		allocGroup, err := bp.CreateAllocGroup(ctx, utils.ToPtr(allocGroupCfg))
		require.NoError(t, err)
		return allocGroup
	}

	newVniAllocationGroup := func(t testing.TB) apstra.ObjectId {
		t.Helper()

		vniRange := []apstra.IntfIntRange{
			apstra.IntRangeRequest{First: 5000, Last: 6000},
		}

		vniPoolId, err := bp.Client().CreateVniPool(ctx, &apstra.VniPoolRequest{
			DisplayName: acctest.RandString(6),
			Ranges:      vniRange,
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteVniPool(ctx, vniPoolId)) })

		// now create the allocation group
		allocGroupCfg := apstra.FreeformAllocGroupData{
			Name:    "vniAlGr-" + acctest.RandString(6),
			Type:    enum.ResourcePoolTypeVni,
			PoolIds: []apstra.ObjectId{vniPoolId},
		}
		allocGroup, err := bp.CreateAllocGroup(ctx, utils.ToPtr(allocGroupCfg))
		require.NoError(t, err)
		return allocGroup
	}

	newAsnAllocationGroup := func(t testing.TB) apstra.ObjectId {
		t.Helper()

		asnRange := []apstra.IntfIntRange{
			apstra.IntRangeRequest{First: 65535, Last: 65700},
		}

		asnPoolId, err := bp.Client().CreateAsnPool(ctx, &apstra.AsnPoolRequest{
			DisplayName: acctest.RandString(6),
			Ranges:      asnRange,
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteAsnPool(ctx, asnPoolId)) })

		// now create the allocation group
		allocGroupCfg := apstra.FreeformAllocGroupData{
			Name:    "asnAlGr-" + acctest.RandString(6),
			Type:    enum.ResourcePoolTypeAsn,
			PoolIds: []apstra.ObjectId{asnPoolId},
		}
		allocGroup, err := bp.CreateAllocGroup(ctx, utils.ToPtr(allocGroupCfg))
		require.NoError(t, err)
		return allocGroup
	}

	newIntAllocationGroup := func(t testing.TB) apstra.ObjectId {
		t.Helper()

		intRange := []apstra.IntfIntRange{
			apstra.IntRangeRequest{First: 10, Last: 65700},
		}

		intPoolId, err := bp.Client().CreateIntegerPool(ctx, &apstra.IntPoolRequest{
			DisplayName: acctest.RandString(6),
			Ranges:      intRange,
		})
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, bp.Client().DeleteIntegerPool(ctx, intPoolId)) })

		// now create the allocation group
		allocGroupCfg := apstra.FreeformAllocGroupData{
			Name:    "intAlGr-" + acctest.RandString(6),
			Type:    enum.ResourcePoolTypeInt,
			PoolIds: []apstra.ObjectId{intPoolId},
		}
		allocGroup, err := bp.CreateAllocGroup(ctx, utils.ToPtr(allocGroupCfg))
		require.NoError(t, err)
		return allocGroup
	}

	type testStep struct {
		config resourceFreeformResourceGenerator
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"start_asn_resource_generator": {
			steps: []testStep{
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						resourceType:  enum.FFResourceTypeAsn,
						scope:         "node('system', name='target')",
						allocatedFrom: string(newAsnAllocationGroup(t)),
						containerId:   string(groupId),
					},
				},
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						resourceType:  enum.FFResourceTypeAsn,
						scope:         "node('system', deploy_mode='deploy', name='target')",
						allocatedFrom: string(newAsnAllocationGroup(t)),
						containerId:   string(groupId),
					},
				},
			},
		},
		//"start_vni_with_static_value": {
		//	steps: []testStep{
		//		{
		//			config: resourceFreeformResourceGenerator{
		//				blueprintId:  bp.Id().String(),
		//				name:         acctest.RandString(6),
		//				containerId:  string(groupId),
		//				resourceType: enum.FFResourceTypeVni,
		//			},
		//		},
		//	},
		//},
		"start_int_resource_generator": {
			steps: []testStep{
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						scope:         "node('system', name='target')",
						containerId:   string(groupId),
						resourceType:  enum.FFResourceTypeInt,
						allocatedFrom: string(newIntAllocationGroup(t)),
					},
				},
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						scope:         "node('system', deploy_mode='deploy', name='target')",
						containerId:   string(groupId),
						resourceType:  enum.FFResourceTypeInt,
						allocatedFrom: string(newIntAllocationGroup(t)),
					},
				},
			},
		},
		"start_ipv4_resource_generator": {
			steps: []testStep{
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						scope:           "node('system', name='target')",
						containerId:     string(groupId),
						resourceType:    enum.FFResourceTypeIpv4,
						allocatedFrom:   string(newIpv4AllocationGroup(t)),
						subnetPrefixLen: utils.ToPtr(27),
					},
				},
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						scope:           "node('system', deploy_mode='deploy', name='target')",
						containerId:     string(groupId),
						resourceType:    enum.FFResourceTypeIpv4,
						allocatedFrom:   string(newIpv4AllocationGroup(t)),
						subnetPrefixLen: utils.ToPtr(28),
					},
				},
			},
		},
		"start_ipv6_with_static_value": {
			steps: []testStep{
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						scope:           "node('system', name='target')",
						containerId:     string(groupId),
						resourceType:    enum.FFResourceTypeIpv6,
						allocatedFrom:   string(newIpv6AllocationGroup(t)),
						subnetPrefixLen: utils.ToPtr(127),
					},
				},
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						scope:           "node('system', deploy_mode='deploy', name='target')",
						containerId:     string(groupId),
						resourceType:    enum.FFResourceTypeIpv6,
						allocatedFrom:   string(newIpv6AllocationGroup(t)),
						subnetPrefixLen: utils.ToPtr(126),
					},
				},
			},
		},
		"start_Vni_with_alloc_group": {
			steps: []testStep{
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						scope:         "node('system', name='target')",
						containerId:   string(groupId),
						resourceType:  enum.FFResourceTypeVni,
						allocatedFrom: string(newVniAllocationGroup(t)),
					},
				},
				{
					config: resourceFreeformResourceGenerator{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						scope:         "node('system', deploy_mode='deploy', name='target')",
						containerId:   string(groupId),
						resourceType:  enum.FFResourceTypeVni,
						allocatedFrom: string(newVniAllocationGroup(t)),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformResourceGenerator)

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
