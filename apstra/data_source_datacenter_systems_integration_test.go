//go:build integration

package tfapstra_test

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	dataSourceDataCenterSystemsHCL = `
data %q %q {
  blueprint_id = %q
  filters      = %s
}
`

	dataSourceDataCenterSystemsFilterHCL = `
    {
      hostname    = %s
      id          = %s
      label       = %s
      role        = %s
      system_id   = %s
      system_type = %s
      tag_ids     = %s
    },`
)

type dataSourceDataCenterSystems struct {
	blueprintId apstra.ObjectId
	filters     []dataSourceDataCenterSystemsFilter
}

func (o dataSourceDataCenterSystems) render(rType, rName, bpId string) string {
	filters := new(bytes.Buffer)
	if len(o.filters) == 0 {
		filters.WriteString("null")
	} else {
		filters.WriteString("[")
		for _, filter := range o.filters {
			filters.WriteString(filter.render())
		}
		filters.WriteString("\n  ]")
	}

	return fmt.Sprintf(dataSourceDataCenterSystemsHCL,
		rType, rName,
		bpId,
		filters.String(),
	)
}

type dataSourceDataCenterSystemsFilter struct {
	hostname   string
	id         apstra.ObjectId
	label      string
	role       *apstra.SystemRole
	systemId   string
	systemType *apstra.SystemType
	tagIds     []string
}

func (o dataSourceDataCenterSystemsFilter) render() string {
	var systemType string
	if o.systemType != nil {
		systemType = o.systemType.String()
	}

	var role string
	if o.role != nil {
		role = o.role.String()
	}

	return fmt.Sprintf(dataSourceDataCenterSystemsFilterHCL,
		stringOrNull(o.hostname),
		stringOrNull(o.id),
		stringOrNull(o.label),
		stringOrNull(role),
		stringOrNull(o.systemId),
		stringOrNull(systemType),
		stringSliceOrNull(o.tagIds),
	)
}

func TestAccDataSourceDatacenterSystems(t *testing.T) {
	ctx := context.Background()
	bp := testutils.BlueprintA(t, ctx)

	var response struct {
		Nodes map[apstra.ObjectId]struct{} `json:"nodes"`
	}

	err := bp.GetNodes(ctx, apstra.NodeTypeSystem, &response)
	require.NoError(t, err)

	tags := make([]string, len(response.Nodes))

	var i int
	for nodeId := range response.Nodes {
		tags[i] = acctest.RandString(6)
		i++
		err = bp.SetNodeTags(ctx, nodeId, tags[:i])
		require.NoError(t, err)
	}

	type testCase struct {
		config dataSourceDataCenterSystems
		checks []resource.TestCheckFunc
	}

	datasourceType := tfapstra.DatasourceName(ctx, &tfapstra.DataSourceDatacenterSystemNodes)

	testCases := map[string]testCase{
		"unfiltered": {
			config: dataSourceDataCenterSystems{
				blueprintId: bp.Id(),
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+".unfiltered", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("data."+datasourceType+".unfiltered", "ids.#", strconv.Itoa(len(response.Nodes))),
			},
		},
		"all_systems_with_filter_per_type": {
			config: dataSourceDataCenterSystems{
				blueprintId: bp.Id(),
				filters: []dataSourceDataCenterSystemsFilter{
					{role: pointer.To(apstra.SystemRoleSuperSpine)},
					{role: pointer.To(apstra.SystemRoleSpine)},
					{role: pointer.To(apstra.SystemRoleLeaf)},
					{role: pointer.To(apstra.SystemRoleAccess)},
					{role: pointer.To(apstra.SystemRoleGeneric)},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+".all_systems_with_filter_per_type", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("data."+datasourceType+".all_systems_with_filter_per_type", "ids.#", strconv.Itoa(len(response.Nodes))),
			},
		},
		"spines_one_filter": {
			config: dataSourceDataCenterSystems{
				blueprintId: bp.Id(),
				filters: []dataSourceDataCenterSystemsFilter{
					{
						systemType: pointer.To(apstra.SystemTypeSwitch),
						role:       pointer.To(apstra.SystemRoleSpine),
					},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+".spines_one_filter", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("data."+datasourceType+".spines_one_filter", "ids.#", "2"),
			},
		},
		"spines_two_filters": {
			config: dataSourceDataCenterSystems{
				blueprintId: bp.Id(),
				filters: []dataSourceDataCenterSystemsFilter{
					{
						systemType: pointer.To(apstra.SystemTypeSwitch),
						role:       pointer.To(apstra.SystemRoleSpine),
					},
					{
						systemType: pointer.To(apstra.SystemTypeSwitch),
						role:       pointer.To(apstra.SystemRoleSpine),
					},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+".spines_two_filters", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("data."+datasourceType+".spines_two_filters", "ids.#", "2"),
			},
		},
	}

	filters := make([]dataSourceDataCenterSystemsFilter, len(tags))

	// generate additional test cases
	for i := range len(tags) {
		filters[i] = dataSourceDataCenterSystemsFilter{tagIds: tags[i : i+1]}

		tName := fmt.Sprintf("single_tag_filter_should_find_%d_hosts", len(tags)-i)
		testCases[tName] = testCase{
			config: dataSourceDataCenterSystems{
				blueprintId: bp.Id(),
				filters:     []dataSourceDataCenterSystemsFilter{{tagIds: tags[:i+1]}},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+"."+tName, "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("data."+datasourceType+"."+tName, "ids.#", strconv.Itoa(len(tags)-i)),
			},
		}
	}

	// generate additional test cases
	for i := range tags {
		tName := fmt.Sprintf("multiple_tag_filter_should_find_%d_hosts", i+1)
		testCases[tName] = testCase{
			config: dataSourceDataCenterSystems{
				blueprintId: bp.Id(),
				filters:     filters[len(filters)-1-i:],
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+"."+tName, "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("data."+datasourceType+"."+tName, "ids.#", strconv.Itoa(i+1)),
			},
		}
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			config := tCase.config.render(datasourceType, tName, bp.Id().String())
			t.Logf("\n// ------ begin config for %s ------%s// -------- end config for %s ------\n\n", tName, config, tName)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: insecureProviderConfigHCL + config,
						Check:  resource.ComposeAggregateTestCheckFunc(tCase.checks...),
					},
				},
			})
		})
	}
}
