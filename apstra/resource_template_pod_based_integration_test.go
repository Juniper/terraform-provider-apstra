//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceTemplatePodBasedHCL = `
resource %q %q {
  name = %q
  super_spine = {
    logical_device_id = %q
    per_plane_count = %d
    plane_count = %d
  }
  pod_infos = {%s 
  }
}
`
	resourceTemplatePodBasedPodHCL = "\n    %s = { count = %d }"
)

type resourceTestPodTemplate struct {
	name          string
	ssLd          string
	perPlaneCount int
	planeCount    int
	podInfo       map[string]int
}

func (o resourceTestPodTemplate) render(rType, rName string) string {
	sb := new(strings.Builder)
	for k, v := range o.podInfo {
		sb.WriteString(fmt.Sprintf(resourceTemplatePodBasedPodHCL, k, v))
	}
	return fmt.Sprintf(resourceTemplatePodBasedHCL,
		rType, rName,
		o.name,
		o.ssLd,
		o.perPlaneCount,
		o.planeCount,
		sb.String(),
	)
}

func (o resourceTestPodTemplate) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "super_spine.logical_device_id", o.ssLd)
	result.append(t, "TestCheckResourceAttr", "super_spine.per_plane_count", strconv.Itoa(o.perPlaneCount))
	result.append(t, "TestCheckResourceAttr", "super_spine.plane_count", strconv.Itoa(o.planeCount))

	result.append(t, "TestCheckResourceAttr", "pod_infos.%", strconv.Itoa(len(o.podInfo)))
	for k, v := range o.podInfo {
		result.append(t, "TestCheckResourceAttr", fmt.Sprintf("pod_infos[%s].count", k), strconv.Itoa(v))
	}

	return result
}

func TestResourceTemplatePodBased(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	type testStep struct {
		config resourceTestPodTemplate
	}
	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"minimal": {
			apiVersionConstraints: apiversions.Ge411,
			steps: []testStep{
				{
					config: resourceTestPodTemplate{
						name:          acctest.RandString(6),
						ssLd:          "AOS-32x40-3",
						perPlaneCount: 2,
						planeCount:    4,
						podInfo: map[string]int{
							"pod1": rand.Intn(3) + 2,
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceTemplatePodBased)

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
