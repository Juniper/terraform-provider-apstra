//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"

	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceTemplatePodBasedHCL = `
resource %q %q {
  name                   = %q
  fabric_link_addressing = %s
  super_spine = {
    logical_device_id = %q
    per_plane_count   = %d
    plane_count       = %s
  }
  pod_infos = {%s 
  }
}
`
	resourceTemplatePodBasedPodHCL = "\n    %s = { count = %d }"
)

type resourceTestPodTemplate struct {
	name                 string
	fabricLinkAddressing *string
	ssLd                 string
	perPlaneCount        int
	planeCount           *int
	podInfo              map[string]int
}

func (o resourceTestPodTemplate) render(rType, rName string) string {
	sb := new(strings.Builder)
	for k, v := range o.podInfo {
		sb.WriteString(fmt.Sprintf(resourceTemplatePodBasedPodHCL, k, v))
	}
	return fmt.Sprintf(resourceTemplatePodBasedHCL,
		rType, rName,
		o.name,
		stringPtrOrNull(o.fabricLinkAddressing),
		o.ssLd,
		o.perPlaneCount,
		intPtrOrNull(o.planeCount),
		sb.String(),
	)
}

func (o resourceTestPodTemplate) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	if o.fabricLinkAddressing != nil {
		result.append(t, "TestCheckResourceAttr", "fabric_link_addressing", *o.fabricLinkAddressing)
	}

	result.append(t, "TestCheckResourceAttr", "super_spine.logical_device_id", o.ssLd)
	result.append(t, "TestCheckResourceAttr", "super_spine.per_plane_count", strconv.Itoa(o.perPlaneCount))

	if o.planeCount == nil {
		result.append(t, "TestCheckResourceAttr", "super_spine.plane_count", "1")
	} else {
		result.append(t, "TestCheckResourceAttr", "super_spine.plane_count", strconv.Itoa(*o.planeCount))
	}

	result.append(t, "TestCheckResourceAttr", "pod_infos.%", strconv.Itoa(len(o.podInfo)))
	for k, v := range o.podInfo {
		result.append(t, "TestCheckResourceAttr", fmt.Sprintf("pod_infos.%s.count", k), strconv.Itoa(v))
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
		"apstra_410_only": {
			apiVersionConstraints: apiversions.Eq410,
			steps: []testStep{
				{
					config: resourceTestPodTemplate{
						name:                 acctest.RandString(6),
						fabricLinkAddressing: utils.ToPtr(apstra.AddressingSchemeIp4.String()),
						ssLd:                 "AOS-4x40_8x10-1",
						perPlaneCount:        2,
						podInfo: map[string]int{
							"pod_single": 1,
							"pod_mlag":   1,
						},
					},
				},
				{
					config: resourceTestPodTemplate{
						name:                 acctest.RandString(6),
						fabricLinkAddressing: utils.ToPtr(apstra.AddressingSchemeIp4.String()),
						ssLd:                 "AOS-4x40_8x10-1",
						perPlaneCount:        4,
						planeCount:           utils.ToPtr(2),
						podInfo: map[string]int{
							"pod_single": 2,
							"pod_mlag":   2,
						},
					},
				},
			},
		},
		"apstra_411_and_later": {
			apiVersionConstraints: apiversions.Ge411,
			steps: []testStep{
				{
					config: resourceTestPodTemplate{
						name:          acctest.RandString(6),
						ssLd:          "AOS-4x40_8x10-1",
						perPlaneCount: 2,
						podInfo: map[string]int{
							"pod_single": 1,
							"pod_mlag":   1,
						},
					},
				},
				{
					config: resourceTestPodTemplate{
						name:          acctest.RandString(6),
						ssLd:          "AOS-24x10-2",
						perPlaneCount: 4,
						planeCount:    utils.ToPtr(2),
						podInfo: map[string]int{
							"pod_single": 2,
							"pod_mlag":   2,
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
