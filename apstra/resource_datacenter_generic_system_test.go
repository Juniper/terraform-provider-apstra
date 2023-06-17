package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"log"
	"strings"
	testutils "terraform-provider-apstra/apstra/test_utils"
	"testing"
)

const (
	resourceDataCenterGenericSystemHCL = `
resource "apstra_datacenter_generic_system" "test" {
  blueprint_id = %s
  name         = %s
  hostname     = %s
  tags = %s
  links = [
%s  ]
}
`

	resourceDataCenterGenericSystemLinkHCL = `    {
      tags                          = %s
      lag_mode                      = %s
      target_switch_id              = %s
      target_switch_if_name         = %s
      target_switch_if_transform_id = %d
      group_label                   = %s
    },
`
)

func TestResourceDatacenterGenericSystem_A(t *testing.T) {
	ctx := context.Background()

	bpClient, bpDelete, err := testutils.BlueprintD(ctx)
	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}
	defer func() {
		err = bpDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	stringOrNull := func(in string) string {
		if in == "" {
			return "null"
		}
		return `"` + in + `"`
	}

	type tagSlice []string
	renderTags := func(in tagSlice) string {
		if len(in) == 0 {
			return "null"
		}
		return `["` + strings.Join(in, `","`) + `"]`
	}

	type link struct {
		tags           []string
		lagMode        apstra.RackLinkLagMode
		targetSwitchId string
		targetSwitchIf string
		targetSwitchTf int
		groupLabel     string
	}
	renderLink := func(in link) string {
		return fmt.Sprintf(resourceDataCenterGenericSystemLinkHCL,
			renderTags(in.tags),
			stringOrNull(in.lagMode.String()),
			stringOrNull(in.targetSwitchId),
			stringOrNull(in.targetSwitchIf),
			in.targetSwitchTf,
			stringOrNull(in.groupLabel),
		)
	}
	renderLinks := func(in []link) string {
		sb := strings.Builder{}
		for i := range in {
			sb.WriteString(renderLink(in[i]))
		}
		return sb.String()
	}

	type genericSystem struct {
		bpId     string
		name     string
		hostname string
		tags     tagSlice
		links    []link
	}
	renderGenericSystem := func(in genericSystem) string {
		return fmt.Sprintf(resourceDataCenterGenericSystemHCL,
			stringOrNull(bpClient.Id().String()),
			stringOrNull(in.name),
			stringOrNull(in.hostname),
			renderTags(in.tags),
			renderLinks(in.links),
		)
	}
	_ = renderGenericSystem

	leafQuery := new(apstra.PathQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(bpClient.Id()).
		SetClient(bpClient.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal("leaf")},
			{Key: "name", Value: apstra.QEStringVal("n_leaf")},
		})
	var leafQueryResult struct {
		Items []struct {
			Leaf struct {
				Id string `json:"id"`
			} `json:"n_leaf"`
		} `json:"items"`
	}
	err = leafQuery.Do(ctx, &leafQueryResult)
	if err != nil {
		t.Fatal(err)
	}
	leafIds := make([]string, len(leafQueryResult.Items))
	for i, item := range leafQueryResult.Items {
		leafIds[i] = item.Leaf.Id
	}

	type testCase struct {
		genericSystem genericSystem
		testCheckFunc resource.TestCheckFunc
	}

	testCases := []testCase{
		{
			genericSystem: genericSystem{
				links: []link{
					{
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "id"),
				//resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "blueprint_id", bpClient.Id().String()),
				//resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "name"),
				//resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "hostname"),
			}...),
		},
	}

	steps := make([]resource.TestStep, len(testCases))
	for i, tc := range testCases {
		tc.genericSystem.bpId = bpClient.Id().String()
		steps[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + renderGenericSystem(tc.genericSystem),
			Check:  tc.testCheckFunc,
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
	log.Println("hello")
}
