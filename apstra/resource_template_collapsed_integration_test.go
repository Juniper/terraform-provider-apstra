//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strconv"
	"testing"
)

const (
	resourceTemplateCollapsedHCL = `
resource %q %q {
  name            = %q // mandatory field
  rack_type_id    = %q // mandatory field
  mesh_link_speed = %q // mandatory field
  mesh_link_count = %d // mandatory field
}
`
)

type collapsedTemplateConfig struct {
	name          string
	rackTypeId    string
	meshLinkSpeed string
	meshLinkCount int
}

func (o collapsedTemplateConfig) render(rType, rName string) string {
	return fmt.Sprintf(resourceTemplateCollapsedHCL,
		rType, rName,
		o.name,
		o.rackTypeId,
		o.meshLinkSpeed,
		o.meshLinkCount,
	)
}

func (o collapsedTemplateConfig) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "rack_type_id", o.rackTypeId)
	result.append(t, "TestCheckResourceAttr", "mesh_link_speed", o.meshLinkSpeed)
	result.append(t, "TestCheckResourceAttr", "mesh_link_count", strconv.Itoa(o.meshLinkCount))

	return result
}

func TestResourceTemplateCollapsed(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)

	type testCase struct {
		apiVersionConstraints version.Constraints
		stepConfigs           []collapsedTemplateConfig
	}

	testCases := map[string]testCase{
		"a": {
			//apiVersionConstraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra411)),
			stepConfigs: []collapsedTemplateConfig{
				{
					name:          acctest.RandString(6),
					rackTypeId:    "L3_collapsed_acs",
					meshLinkSpeed: "10G",
					meshLinkCount: 1,
				},
				{
					name:          acctest.RandString(6),
					rackTypeId:    "L3_collapsed_acs",
					meshLinkSpeed: "10G",
					meshLinkCount: 2,
				},
				{
					name:          acctest.RandString(6),
					rackTypeId:    "L3_collapsed_ESI",
					meshLinkSpeed: "10G",
					meshLinkCount: 1,
				},
			},
		},
	}

	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))
	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceTemplateCollapsed)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("API version %s does not satisfy version constraints(%s) of test %q",
					apiVersion, tCase.apiVersionConstraints, tName)
			}

			steps := make([]resource.TestStep, len(tCase.stepConfigs))
			for i, stepConfig := range tCase.stepConfigs {
				config := stepConfig.render(resourceType, tName)
				checks := stepConfig.testChecks(t, resourceType, tName)

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
