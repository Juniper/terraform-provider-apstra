//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
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
  assigned_to         = %s
 }
`
)

type resourceFreeformResource struct {
	blueprintId   string
	name          string
	groupId       apstra.ObjectId
	resourceType  enum.FFResourceType
	integerValue  *int
	ipv4Value     net.IPNet
	ipv6Value     net.IPNet
	allocatedFrom apstra.ObjectId
	assignedTo    []string
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
		stringOrNull(o.allocatedFrom.String()),
		stringSliceOrNull(o.assignedTo),
	)
}

func (o resourceFreeformResource) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "group_id", o.groupId.String())
	result.append(t, "TestCheckResourceAttr", "type", utils.StringersToFriendlyString(o.resourceType))
	if o.integerValue != nil {
		result.append(t, "TestCheckResourceAttr", "integer_value", strconv.Itoa(*o.integerValue))
	} else {
		if o.resourceType == enum.FFResourceTypeHostIpv4 || o.resourceType == enum.FFResourceTypeHostIpv6 {
			result.append(t, "TestCheckNoResourceAttr", "integer_value")
		}
		if o.resourceType == enum.FFResourceTypeIpv4 && o.allocatedFrom == "" {
			result.append(t, "TestCheckNoResourceAttr", "integer_value")
		}
		if o.resourceType == enum.FFResourceTypeIpv6 && o.allocatedFrom == "" {
			result.append(t, "TestCheckNoResourceAttr", "integer_value")
		}
	}
	if o.ipv4Value.String() != "<nil>" {
		result.append(t, "TestCheckResourceAttr", "ipv4_value", o.ipv4Value.String())
	}
	if o.ipv6Value.String() != "<nil>" {
		result.append(t, "TestCheckResourceAttr", "ipv6_value", o.ipv6Value.String())
	}
	if o.resourceType == enum.FFResourceTypeIpv4 || o.resourceType == enum.FFResourceTypeHostIpv4 {
		result.append(t, "TestCheckResourceAttrSet", "ipv4_value")
	}
	if o.resourceType == enum.FFResourceTypeIpv6 || o.resourceType == enum.FFResourceTypeHostIpv6 {
		result.append(t, "TestCheckResourceAttrSet", "ipv6_value")
	}
	if o.allocatedFrom == "" {
		result.append(t, "TestCheckNoResourceAttr", "allocated_from")
	} else {
		result.append(t, "TestCheckResourceAttr", "allocated_from", o.allocatedFrom.String())
	}
	if len(o.assignedTo) > 0 {
		result.append(t, "TestCheckResourceAttr", "assigned_to.#", strconv.Itoa(len(o.assignedTo)))
		for _, assignedTo := range o.assignedTo {
			result.append(t, "TestCheckTypeSetElemAttr", "assigned_to.*", assignedTo)
		}
	} else {
		result.append(t, "TestCheckNoResourceAttr", "assigned_to")
	}

	return result
}

func TestResourceFreeformResource(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	var err error

	// create a blueprint and a group...
	intSysCount := 5
	extSysCount := 5
	sysCount := intSysCount + extSysCount
	bp, intSysIds, extSysIds := testutils.FfBlueprintB(t, ctx, intSysCount, extSysCount)
	require.Equal(t, intSysCount, len(intSysIds))
	require.Equal(t, extSysCount, len(extSysIds))

	// create resource groups
	groupCount := 2
	groupIds := make([]apstra.ObjectId, groupCount)
	for i := range groupCount {
		groupIds[i], err = bp.CreateRaGroup(ctx, &apstra.FreeformRaGroupData{Label: acctest.RandString(6)})
		require.NoError(t, err)
	}

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
		allocGroup, err := bp.CreateAllocGroup(ctx, pointer.To(allocGroupCfg))
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
		allocGroup, err := bp.CreateAllocGroup(ctx, pointer.To(allocGroupCfg))
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
		allocGroup, err := bp.CreateAllocGroup(ctx, pointer.To(allocGroupCfg))
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
		allocGroup, err := bp.CreateAllocGroup(ctx, pointer.To(allocGroupCfg))
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
		allocGroup, err := bp.CreateAllocGroup(ctx, pointer.To(allocGroupCfg))
		require.NoError(t, err)
		return allocGroup
	}

	var allSysIDs []string
	randomSysIds := func(t testing.TB, count int) []string {
		if count > intSysCount+extSysCount {
			t.Fatalf("caller requested %d systemIDs blueprint only  has %d internal and %d external", count, intSysCount, extSysCount)
		}

		if allSysIDs == nil {
			allSysIDs = make([]string, intSysCount+extSysCount)
			for i, id := range intSysIds {
				allSysIDs[i] = id.String()
			}
			for i, id := range extSysIds {
				allSysIDs[i+intSysCount] = id.String()
			}
		}

		resultMap := make(map[string]struct{}, count)
		for len(resultMap) < count {
			resultMap[allSysIDs[rand.IntN(len(allSysIDs))]] = struct{}{}
		}

		result := make([]string, len(resultMap))
		var i int
		for k := range resultMap {
			result[i] = k
			i++
		}

		return result
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
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeAsn,
						integerValue: pointer.To(65535),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      groupIds[1],
						resourceType: enum.FFResourceTypeAsn,
						integerValue: pointer.To(65536),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       groupIds[0],
						resourceType:  enum.FFResourceTypeAsn,
						allocatedFrom: newAsnAllocationGroup(t),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       groupIds[1],
						resourceType:  enum.FFResourceTypeAsn,
						allocatedFrom: newAsnAllocationGroup(t),
						assignedTo:    randomSysIds(t, sysCount/3),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeAsn,
						integerValue: pointer.To(65536),
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
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeVni,
						integerValue: pointer.To(4498),
						assignedTo:   randomSysIds(t, sysCount/3),
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
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeInt,
						integerValue: pointer.To(4498),
						assignedTo:   randomSysIds(t, sysCount/3),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       groupIds[1],
						resourceType:  enum.FFResourceTypeInt,
						allocatedFrom: newIntAllocationGroup(t),
						assignedTo:    randomSysIds(t, sysCount/3),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeInt,
						integerValue: pointer.To(459),
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
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeIpv4,
						ipv4Value:    randomSlash31(t, "192.168.2.0/24"),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       groupIds[1],
						resourceType:  enum.FFResourceTypeIpv4,
						integerValue:  pointer.To(30),
						allocatedFrom: newIpv4AllocationGroup(t),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeIpv4,
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
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeIpv6,
						ipv6Value:    randomSlash127(t, "2001:db8::/32"),
						assignedTo:   randomSysIds(t, sysCount/3),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       groupIds[1],
						resourceType:  enum.FFResourceTypeIpv6,
						integerValue:  pointer.To(64),
						allocatedFrom: newIpv6AllocationGroup(t),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeIpv6,
						ipv6Value:    randomSlash127(t, "2001:db8::/32"),
					},
				},
			},
		},
		"start_vni_with_alloc_group": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       groupIds[0],
						resourceType:  enum.FFResourceTypeVni,
						allocatedFrom: newVniAllocationGroup(t),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      groupIds[1],
						resourceType: enum.FFResourceTypeVni,
						integerValue: pointer.To(rand.IntN(10000) + 4096),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      groupIds[0],
						resourceType: enum.FFResourceTypeVni,
						integerValue: pointer.To(rand.IntN(10000) + 4096),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId:   bp.Id().String(),
						name:          acctest.RandString(6),
						groupId:       groupIds[1],
						resourceType:  enum.FFResourceTypeVni,
						allocatedFrom: newVniAllocationGroup(t),
						assignedTo:    randomSysIds(t, sysCount/3),
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
