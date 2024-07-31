//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"net"
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
	resourceFreeformResourceHcl = `
resource %q %q {
  blueprint_id        = %q
  name                = %q
  group_id            = %q
  type                = %q
  integer_value       = %s
  ipv4_value          = %s
  ipv6_value          = %s
  allocated_from      = %s
 }
`
)

type resourceFreeformResource struct {
	blueprintId   string
	name          string
	groupId       string
	resourceType  apstra.FFResourceType
	integerValue  *int
	ipv4Value     net.IPNet
	ipv6Value     net.IPNet
	allocatedFrom string
}

func (o resourceFreeformResource) render(rType, rName string) string {
	return fmt.Sprintf(resourceFreeformResourceHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		o.groupId,
		utils.StringersToFriendlyString(o.resourceType),
		intPtrOrNull(o.integerValue),
		ipNetOrNull(o.ipv4Value),
		ipNetOrNull(o.ipv6Value),
		stringOrNull(o.allocatedFrom),
	)
}

func (o resourceFreeformResource) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "group_id", o.groupId)
	result.append(t, "TestCheckResourceAttr", "type", utils.StringersToFriendlyString(o.resourceType))
	if o.integerValue != nil {
		result.append(t, "TestCheckResourceAttr", "integer_value", strconv.Itoa(*o.integerValue))
	} else {
		if o.resourceType == apstra.FFResourceTypeHostIpv4 || o.resourceType == apstra.FFResourceTypeHostIpv6 {
			result.append(t, "TestCheckNoResourceAttr", "integer_value")
		}
		if o.resourceType == apstra.FFResourceTypeIpv4 && o.allocatedFrom == "" {
			result.append(t, "TestCheckNoResourceAttr", "integer_value")
		}
		if o.resourceType == apstra.FFResourceTypeIpv6 && o.allocatedFrom == "" {
			result.append(t, "TestCheckNoResourceAttr", "integer_value")
		}
	}
	if o.ipv4Value.String() != "<nil>" {
		result.append(t, "TestCheckResourceAttr", "ipv4_value", o.ipv4Value.String())
	}
	if o.ipv6Value.String() != "<nil>" {
		result.append(t, "TestCheckResourceAttr", "ipv6_value", o.ipv6Value.String())
	}
	if o.resourceType == apstra.FFResourceTypeIpv4 || o.resourceType == apstra.FFResourceTypeHostIpv4 {
		result.append(t, "TestCheckResourceAttrSet", "ipv4_value")
	}
	if o.resourceType == apstra.FFResourceTypeIpv6 || o.resourceType == apstra.FFResourceTypeHostIpv6 {
		result.append(t, "TestCheckResourceAttrSet", "ipv6_value")
	}
	if o.allocatedFrom == "" {
		result.append(t, "TestCheckNoResourceAttr", "allocated_from")
	} else {
		result.append(t, "TestCheckResourceAttr", "allocated_from", o.allocatedFrom)
	}

	return result
}

func TestResourceFreeformResource(t *testing.T) {
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
			Type:    apstra.ResourcePoolTypeIpv4,
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
			Type:    apstra.ResourcePoolTypeIpv6,
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
			Type:    apstra.ResourcePoolTypeVni,
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
			Type:    apstra.ResourcePoolTypeAsn,
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
			Type:    apstra.ResourcePoolTypeInt,
			PoolIds: []apstra.ObjectId{intPoolId},
		}
		allocGroup, err := bp.CreateAllocGroup(ctx, utils.ToPtr(allocGroupCfg))
		require.NoError(t, err)
		return allocGroup
	}

	type testStep struct {
		config resourceFreeformResource
	}
	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"start_asn_with_static_value": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeAsn,
						integerValue: utils.ToPtr(65535),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeAsn,
						integerValue: utils.ToPtr(65536),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       string(groupId),
						resourceType:  apstra.FFResourceTypeAsn,
						allocatedFrom: string(newAsnAllocationGroup(t)),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       string(groupId),
						resourceType:  apstra.FFResourceTypeAsn,
						integerValue:  utils.ToPtr(65536),
						allocatedFrom: string(newAsnAllocationGroup(t)),
					},
				},
			},
		},
		"start_vni_with_static_value": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeVni,
						integerValue: utils.ToPtr(4498),
					},
				},
			},
		},
		"start_int_with_static_value": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeInt,
						integerValue: utils.ToPtr(4498),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       string(groupId),
						resourceType:  apstra.FFResourceTypeInt,
						allocatedFrom: string(newIntAllocationGroup(t)),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeInt,
						integerValue: utils.ToPtr(459),
					},
				},
			},
		},
		"start_ipv4_with_static_value": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeIpv4,
						ipv4Value:    randomSlash31(t, "192.168.2.0/24"),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       string(groupId),
						resourceType:  apstra.FFResourceTypeIpv4,
						integerValue:  utils.ToPtr(30),
						allocatedFrom: string(newIpv4AllocationGroup(t)),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeIpv4,
						ipv4Value:    randomSlash31(t, "10.168.2.0/24"),
					},
				},
			},
		},
		"start_ipv6_with_static_value": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeIpv6,
						ipv6Value:    randomSlash127(t, "2001:db8::/32"),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       string(groupId),
						resourceType:  apstra.FFResourceTypeIpv6,
						integerValue:  utils.ToPtr(64),
						allocatedFrom: string(newIpv6AllocationGroup(t)),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeIpv6,
						ipv6Value:    randomSlash127(t, "2001:db8::/32"),
					},
				},
			},
		},
		"start_Vni_with_alloc_group": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       string(groupId),
						resourceType:  apstra.FFResourceTypeVni,
						allocatedFrom: string(newVniAllocationGroup(t)),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeVni,
						integerValue: utils.ToPtr(4498),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       string(groupId),
						resourceType:  apstra.FFResourceTypeVni,
						allocatedFrom: string(newVniAllocationGroup(t)),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformResource)

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
