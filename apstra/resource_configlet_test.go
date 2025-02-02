//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceConfigletGeneratorHCL = `    {
      config_style           = %q
      section                = %q
      template_text          = %q
      negation_template_text = %s
      filename               = %s
    },
`

	resourceConfigletHCL = `resource %q %q  {
  name                 = %s
  generators           = %s
}
`
)

type resourceConfigletGenerator struct {
	style                string
	section              string
	templateText         string
	negationTemplateText string
	filename             string
}

func (o resourceConfigletGenerator) render() string {
	return fmt.Sprintf(resourceConfigletGeneratorHCL,
		o.style,
		o.section,
		o.templateText,
		stringOrNull(o.negationTemplateText),
		stringOrNull(o.filename),
	)
}

type resourceConfiglet struct {
	name       string
	generators []resourceConfigletGenerator
}

func (o resourceConfiglet) render(rType, rName string) string {
	generators := "null"

	if len(o.generators) != 0 {
		sb := new(strings.Builder)
		for _, v := range o.generators {
			sb.WriteString(tfapstra.Indent(2, v.render()))
		}
		generators = "[\n" + sb.String() + "  ]"

	}

	return fmt.Sprintf(resourceConfigletHCL,
		rType, rName,
		stringOrNull(o.name),
		generators,
	)
}

func (o resourceConfiglet) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	result.append(t, "TestCheckResourceAttr", "generators.#", strconv.Itoa(len(o.generators)))
	for i, generator := range o.generators {
		result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.config_style", i), generator.style)
		result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.section", i), generator.section)
		result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.template_text", i), generator.templateText)

		if generator.negationTemplateText == "" {
			result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("generators.%d.negation_template_text", i))
		} else {
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.negation_template_text", i), generator.negationTemplateText)
		}

		if generator.filename == "" {
			result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("generators.%d.filename", i))
		} else {
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.filename", i), generator.filename)
		}
	}

	return result
}

func TestAccResourceConfiglet(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)

	type testStep struct {
		config resourceConfiglet
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"swap_generators": {
			steps: []testStep{
				{
					config: resourceConfiglet{
						name: acctest.RandString(6),
						generators: []resourceConfigletGenerator{
							{
								style:        "junos",
								section:      "top_level_hierarchical",
								templateText: "set foo",
							},
							{
								style:                "nxos",
								section:              "system",
								templateText:         "bar",
								negationTemplateText: "no bar",
							},
						},
					},
				},
				{
					config: resourceConfiglet{
						name: acctest.RandString(6),
						generators: []resourceConfigletGenerator{
							{
								style:                "nxos",
								section:              "system",
								templateText:         "bar",
								negationTemplateText: "no bar",
							},
							{
								style:        "junos",
								section:      "top_level_hierarchical",
								templateText: "set foo",
							},
						},
					},
				},
			},
		},
		"file": {
			steps: []testStep{
				{
					config: resourceConfiglet{
						name: acctest.RandString(6),
						generators: []resourceConfigletGenerator{
							{
								style:        "sonic",
								section:      "file",
								templateText: acctest.RandString(6),
								filename:     "/etc/" + acctest.RandString(6),
							},
						},
					},
				},
				{
					config: resourceConfiglet{
						name: acctest.RandString(6),
						generators: []resourceConfigletGenerator{
							{
								style:        "sonic",
								section:      "file",
								templateText: acctest.RandString(6),
								filename:     "/etc/" + acctest.RandString(6),
							},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceConfiglet)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(client.ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
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
