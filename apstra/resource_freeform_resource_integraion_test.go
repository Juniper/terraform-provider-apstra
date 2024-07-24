//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"net"
	"testing"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
  allocation_group_id = %s
 }
`
)

type resourceFreeformResource struct {
	blueprintId     string
	name            string
	groupId         string
	resourceType    string
	integerValue    *int
	ipv4Value       net.IP
	ipv6Value       net.IP
	allocationGroup string
}

func (o resourceFreeformResource) render(rType, rName string) string {
	return fmt.Sprintf(resourceFreeformResourceHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		o.groupId,
		o.resourceType,
		intPtrOrNull(o.integerValue),
		ipOrNull(o.ipv4Value),
		ipOrNull(o.ipv6Value),
		stringOrNull(o.allocationGroup),
	)
}

func (o resourceFreeformResource) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	return result
}

func TestResourceFreeformResource(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint and a group...
	bp, groupId := testutils.FfBlueprintC(t, ctx)

	//newIpv4AllocationGroup := func(t testing.TB) apstra.ObjectId {
	//	t.Helper()
	//
	//	// create an ipv4 pool
	//	randomNet := net.IPNet{
	//		IP:   randIpvAddressMust(t,"10.0.0.0/8"),
	//		Mask: net.CIDRMask(24,32),
	//	}
	//	poolId, err := bp.Client().CreateIp4Pool(ctx, &apstra.NewIpPoolRequest{
	//		DisplayName: acctest.RandString(6),
	//		Subnets:     []apstra.NewIpSubnet{{ Network: randomNet.String() }},
	//	})
	//	require.NoError(t, err)
	//	t.Cleanup(func() {
	//		require.NoError(t,bp.Client().DeleteIp4Pool(ctx,poolId))
	//	})
	//}

	type testStep struct {
		config resourceFreeformResource
	}
	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"start_with_asn": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId:  bp.Id().String(),
						name:         acctest.RandString(6),
						groupId:      string(groupId),
						resourceType: apstra.FFResourceTypeAsn.String(),
						integerValue: utils.ToPtr(65535),
					},
				},
			},
		},
		"start_with_tags": {
			steps: []testStep{
				{
					config: resourceFreeformResource{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceFreeformResource{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
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
}
