//go:build integration

package tfapstra_test

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const dataSourceBlueprintDeviceRenderedConfigHCL = `
data %q "test" {
  blueprint_id = %q
  node_id      = %s
  system_id    = %s
}
`

type dataSourceBlueprintDeviceRenderedConfig struct {
	nodeId   apstra.ObjectId
	systemId apstra.ObjectId
}

func (o dataSourceBlueprintDeviceRenderedConfig) render(rType, bpId string) string {
	return fmt.Sprintf(dataSourceBlueprintDeviceRenderedConfigHCL,
		rType,
		bpId,
		stringOrNull(o.nodeId),
		stringOrNull(o.systemId),
	)
}

func TestAccDatasourceBlueprintDeviceRenderedConfig(t *testing.T) {
	ctx := context.Background()
	bp := testutils.BlueprintI(t, ctx)
	leafMap := testutils.GetSystemIds(t, ctx, bp, "leaf")
	require.Equal(t, 2, len(leafMap))

	type testCase struct {
		preFunc func(testing.TB, context.Context, *apstra.TwoStageL3ClosClient)
		config  dataSourceBlueprintDeviceRenderedConfig
		checks  []resource.TestCheckFunc
	}

	nodeLabels := make([]string, len(leafMap))
	nodeIds := make([]apstra.ObjectId, len(leafMap))
	sysIds := make([]apstra.ObjectId, len(leafMap))
	var i int
	for k, v := range leafMap {
		var node struct {
			SystemId apstra.ObjectId `json:"system_id"`
		}
		require.NoError(t, bp.Client().GetNode(ctx, bp.Id(), v, &node))

		nodeLabels[i] = k
		nodeIds[i] = v
		sysIds[i] = node.SystemId

		i++
	}

	changeLeafAsn := func(t testing.TB, ctx context.Context, leafId apstra.ObjectId, client *apstra.TwoStageL3ClosClient) {
		t.Helper()

		query := new(apstra.PathQuery).
			SetBlueprintId(bp.Id()).
			SetClient(bp.Client()).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "id", Value: apstra.QEStringVal(leafId)},
			}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOfSystems.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeDomain.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("n_domain")},
			})

		type node struct {
			Id  apstra.ObjectId `json:"id"`
			Asn string          `json:"domain_id"`
		}

		var queryResult struct {
			Items []struct {
				Node node `json:"n_domain"`
			} `json:"items"`
		}

		err := query.Do(ctx, &queryResult)
		require.NoError(t, err)
		require.Equal(t, 1, len(queryResult.Items))

		err = client.Client().PatchNodeUnsafe(ctx, bp.Id(), queryResult.Items[0].Node.Id, node{Asn: strconv.Itoa(rand.IntN(math.MaxUint16))}, nil)
		require.NoError(t, err)
	}

	datasourceType := tfapstra.DatasourceName(ctx, &tfapstra.DataSourceBlueprintNodeConfig)

	atLeast100Lines := func(value string) error {
		s := bufio.NewScanner(strings.NewReader(value))
		var i int
		for s.Scan() {
			i++
			if i >= 100 {
				return nil
			}
		}
		return fmt.Errorf("expected 100 lines, got %d lines", i)
	}

	expectAsnChange := func(value string) error {
		var asnAdded, asnRemoved bool
		s := bufio.NewScanner(strings.NewReader(value))
		for s.Scan() {
			line := s.Text()
			switch {
			case strings.HasPrefix(line, "- autonomous-system "):
				asnRemoved = true
			case strings.HasPrefix(line, "+ autonomous-system "):
				asnAdded = true
			}
		}
		if !asnAdded || !asnRemoved {
			return fmt.Errorf("diff should show ASN removed and added; got %q", value)
		}
		return nil
	}

	testCases := map[string]testCase{
		"leaf_0_pre_change_by_node": {
			config: dataSourceBlueprintDeviceRenderedConfig{nodeId: nodeIds[0]},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "node_id", nodeIds[0].String()),
				resource.TestCheckNoResourceAttr("data."+datasourceType+".test", "system_id"),
				resource.TestCheckResourceAttrSet("data."+datasourceType+".test", "deployed_config"),
				resource.TestCheckResourceAttrSet("data."+datasourceType+".test", "staged_config"),
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "incremental_config", ""),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "deployed_config", atLeast100Lines),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "staged_config", atLeast100Lines),
			},
		},
		"leaf_0_pre_change_by_system": {
			config: dataSourceBlueprintDeviceRenderedConfig{systemId: sysIds[0]},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "blueprint_id", bp.Id().String()),
				resource.TestCheckNoResourceAttr("data."+datasourceType+".test", "node_id"),
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "system_id", sysIds[0].String()),
				resource.TestCheckResourceAttrSet("data."+datasourceType+".test", "deployed_config"),
				resource.TestCheckResourceAttrSet("data."+datasourceType+".test", "staged_config"),
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "incremental_config", ""),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "deployed_config", atLeast100Lines),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "staged_config", atLeast100Lines),
			},
		},
		"leaf_0_post_change_by_node": {
			preFunc: func(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient) {
				changeLeafAsn(t, ctx, nodeIds[0], bp)
			},
			config: dataSourceBlueprintDeviceRenderedConfig{nodeId: nodeIds[0]},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "node_id", nodeIds[0].String()),
				resource.TestCheckNoResourceAttr("data."+datasourceType+".test", "system_id"),
				resource.TestCheckResourceAttrSet("data."+datasourceType+".test", "deployed_config"),
				resource.TestCheckResourceAttrSet("data."+datasourceType+".test", "staged_config"),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "incremental_config", expectAsnChange),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "deployed_config", atLeast100Lines),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "staged_config", atLeast100Lines),
			},
		},
		"leaf_0_post_change_by_system": {
			preFunc: func(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient) {
				changeLeafAsn(t, ctx, nodeIds[0], bp)
			},
			config: dataSourceBlueprintDeviceRenderedConfig{systemId: sysIds[0]},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "blueprint_id", bp.Id().String()),
				resource.TestCheckNoResourceAttr("data."+datasourceType+".test", "node_id"),
				resource.TestCheckResourceAttr("data."+datasourceType+".test", "system_id", sysIds[0].String()),
				resource.TestCheckResourceAttrSet("data."+datasourceType+".test", "deployed_config"),
				resource.TestCheckResourceAttrSet("data."+datasourceType+".test", "staged_config"),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "incremental_config", expectAsnChange),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "deployed_config", atLeast100Lines),
				resource.TestCheckResourceAttrWith("data."+datasourceType+".test", "staged_config", atLeast100Lines),
			},
		},
	}

	// bpModificationWg delays modifications to the blueprint until pre-modification tests are complete
	bpModificationWg := new(sync.WaitGroup)

	// testCaseStartWg ensures that no test case starts begins before all have had a chance
	// to pile onto bpModificationWg
	testCaseStartWg := new(sync.WaitGroup)
	testCaseStartWg.Add(len(testCases))

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			if tCase.config.nodeId == "" && tCase.config.systemId == "" {
				testCaseStartWg.Done()
				t.Skipf("skipping because node has no system assigned")
				return
			}
			t.Parallel()

			if tCase.preFunc == nil {
				bpModificationWg.Add(1)
				testCaseStartWg.Done()
			} else {
				testCaseStartWg.Done()
				bpModificationWg.Wait()
				tCase.preFunc(t, ctx, bp)
			}

			config := tCase.config.render(datasourceType, bp.Id().String())
			t.Logf("\n// ------ begin config ------%s// -------- end config ------\n\n", config)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: insecureProviderConfigHCL + config,
						Check:  resource.ComposeAggregateTestCheckFunc(tCase.checks...),
					},
				},
			})

			if tCase.preFunc == nil {
				bpModificationWg.Done() // release test cases which will make changes
			}
		})
	}
}
