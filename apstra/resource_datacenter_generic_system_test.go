package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
				//name:     "",
				//hostname: "",
				//tags:     []string{},
				links: []link{
					{
						//lagMode: apstra.RackLinkLagModeNone,
						//groupLabel: "",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						//tags:     []string{},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "name"),
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "hostname"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "tags"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags"),
			}...),
		},
		{
			genericSystem: genericSystem{
				name:     "foo",
				hostname: "foo.com",
				tags:     []string{"a"},
				links: []link{
					{
						lagMode:        apstra.RackLinkLagModeActive,
						groupLabel:     "foo",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						tags:           []string{"b"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "name", "foo"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "hostname", "foo.com"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "tags.0", "a"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label", "foo"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode", apstra.RackLinkLagModeActive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.0", "b"),
			}...),
		},
		{
			genericSystem: genericSystem{
				//name:     "foo",
				//hostname: "foo.com",
				//tags:     []string{"a"},
				links: []link{
					{
						//lagMode:        apstra.RackLinkLagModeActive,
						//groupLabel:     "foo",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						//tags:           []string{"b"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_generic_system.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "blueprint_id", bpClient.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "name", "foo"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "hostname", "foo.com"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "tags"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_id", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags"),
			}...),
		},
		{
			genericSystem: genericSystem{
				//name:     "foo",
				//hostname: "foo.com",
				//tags:     []string{"a"},
				links: []link{
					{
						lagMode:        apstra.RackLinkLagModePassive,
						groupLabel:     "bar",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						tags:           []string{"c"},
					},
					{
						lagMode:        apstra.RackLinkLagModePassive,
						groupLabel:     "bar",
						targetSwitchId: leafIds[1],
						targetSwitchIf: "xe-0/0/6",
						targetSwitchTf: 1,
						tags:           []string{"c"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "2"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label", "bar"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode", apstra.RackLinkLagModePassive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags.0", "c"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.group_label", "bar"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.lag_mode", apstra.RackLinkLagModePassive.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_name", "xe-0/0/6"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_transform_id", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags.0", "c"),
			}...),
		},
		{
			genericSystem: genericSystem{
				//name:     "foo",
				//hostname: "foo.com",
				//tags:     []string{"a"},
				links: []link{
					{
						lagMode:        apstra.RackLinkLagModeStatic,
						groupLabel:     "baz",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/7",
						targetSwitchTf: 1,
						//tags:           []string{"c"},
					},
					{
						lagMode:        apstra.RackLinkLagModeStatic,
						groupLabel:     "baz",
						targetSwitchId: leafIds[1],
						targetSwitchIf: "xe-0/0/7",
						targetSwitchTf: 1,
						//tags:           []string{"c"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "2"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label", "baz"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode", apstra.RackLinkLagModeStatic.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/7"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.group_label", "baz"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.lag_mode", apstra.RackLinkLagModeStatic.String()),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_name", "xe-0/0/7"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.1.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.1.tags"),
			}...),
		},
		{
			genericSystem: genericSystem{
				//name:     "foo",
				//hostname: "foo.com",
				//tags:     []string{"a"},
				links: []link{
					{
						//lagMode:        apstra.RackLinkLagModeStatic,
						//groupLabel:     "baz",
						targetSwitchId: leafIds[0],
						targetSwitchIf: "xe-0/0/8",
						targetSwitchTf: 1,
						//tags:           []string{"c"},
					},
				},
			},
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.#", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.group_label"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.lag_mode"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_name", "xe-0/0/8"),
				resource.TestCheckResourceAttr("apstra_datacenter_generic_system.test", "links.0.target_switch_if_transform_id", "1"),
				resource.TestCheckNoResourceAttr("apstra_datacenter_generic_system.test", "links.0.tags"),
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
}
