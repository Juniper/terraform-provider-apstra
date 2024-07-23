//go:build integration

package tfapstra_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/stretchr/testify/require"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceFreeformResourceGroupHcl = `
resource %q %q {
  blueprint_id = %q
  name         = %q
  parent_id    = %s
  data         = %s
}
`
)

type resourceFreeformResourceGroup struct {
	blueprintId string
	name        string
	parentId    string
	data        json.RawMessage
	// tags        []string
}

func (o resourceFreeformResourceGroup) render(rType, rName string) string {
	data := "null"
	if o.data != nil {
		data = fmt.Sprintf("%q", string(o.data))
	}

	return fmt.Sprintf(resourceFreeformResourceGroupHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		stringOrNull(o.parentId),
		data,
	)
}

func (o resourceFreeformResourceGroup) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	//if len(o.tags) > 0 {
	//	result.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	//	for _, tag := range o.tags {
	//		result.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
	//	}
	//}

	if len(o.data) > 0 {
		result.append(t, "TestCheckResourceAttr", "data", string(o.data))
	} else {
		result.append(t, "TestCheckResourceAttr", "data", "{}")
	}

	if len(o.parentId) > 0 {
		result.append(t, "TestCheckResourceAttr", "parent_id", o.parentId)
	} else {
		result.append(t, "TestCheckNoResourceAttr", "parent_id")
	}

	return result
}

func TestResourceFreeformResourceGroup(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint
	bp := testutils.FfBlueprintA(t, ctx)

	makeParentResourceGroup := func(t testing.TB, ctx context.Context) apstra.ObjectId {
		id, err := bp.CreateRaGroup(ctx, &apstra.FreeformRaGroupData{Label: acctest.RandString(6)})
		require.NoError(t, err)
		return id
	}

	type testStep struct {
		config resourceFreeformResourceGroup
	}
	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceFreeformResourceGroup{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceFreeformResourceGroup{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						parentId:    makeParentResourceGroup(t, ctx).String(),
						data:        randomJson(t, 6, 12, 4),
					},
				},
				{
					config: resourceFreeformResourceGroup{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: resourceFreeformResourceGroup{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						parentId:    makeParentResourceGroup(t, ctx).String(),
						data:        randomJson(t, 6, 12, 4),
					},
				},
				{
					config: resourceFreeformResourceGroup{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
					},
				},
				{
					config: resourceFreeformResourceGroup{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						parentId:    makeParentResourceGroup(t, ctx).String(),
						data:        randomJson(t, 6, 12, 4),
					},
				},
			},
		},
		"swap_parents": {
			steps: []testStep{
				{
					config: resourceFreeformResourceGroup{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						parentId:    makeParentResourceGroup(t, ctx).String(),
					},
				},
				{
					config: resourceFreeformResourceGroup{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						parentId:    makeParentResourceGroup(t, ctx).String(),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformResourceGroup)

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
