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
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceDatacenterConfigletGeneratorHCL = `    {
      config_style           = %q
      section                = %q
      section_condition      = %s
      template_text          = %q
      negation_template_text = %s
      filename               = %s
    },
`

	resourceDatacenterConfigletHCL = `resource %q %q  {
  blueprint_id         = %q
  name                 = %s
  condition            = %q
  catalog_configlet_id = %s
  generators           = %s
}
`
)

type resourceDatacenterConfigletGenerator struct {
	style                string
	section              string
	sectionCondition     string
	templateText         string
	negationTemplateText string
	filename             string
}

func (o resourceDatacenterConfigletGenerator) render() string {
	return fmt.Sprintf(resourceDatacenterConfigletGeneratorHCL,
		o.style,
		o.section,
		stringOrNull(o.sectionCondition),
		o.templateText,
		stringOrNull(o.negationTemplateText),
		stringOrNull(o.filename),
	)
}

type resourceDatacenterConfiglet struct {
	blueprintId        string
	name               string
	condition          string
	catalogConfigletId string
	generators         []resourceDatacenterConfigletGenerator
}

func (o resourceDatacenterConfiglet) render(rType, rName string) string {
	generators := "null"

	if len(o.generators) != 0 {
		sb := new(strings.Builder)
		for _, v := range o.generators {
			sb.WriteString(tfapstra.Indent(2, v.render()))
		}
		generators = "[\n" + sb.String() + "  ]"

	}

	return fmt.Sprintf(resourceDatacenterConfigletHCL,
		rType, rName,
		o.blueprintId,
		stringOrNull(o.name),
		o.condition,
		stringOrNull(o.catalogConfigletId),
		generators,
	)
}

func (o resourceDatacenterConfiglet) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "condition", o.condition)

	if o.name == "" {
		result.append(t, "TestCheckResourceAttrSet", "name")
	} else {
		result.append(t, "TestCheckResourceAttr", "name", o.name)
	}

	if o.catalogConfigletId == "" {
		result.append(t, "TestCheckNoResourceAttr", "catalog_configlet_id")
	} else {
		result.append(t, "TestCheckResourceAttr", "catalog_configlet_id", o.catalogConfigletId)
	}

	// generators are computed, but we cannot anticipate their contents when using catalog
	// cofiglet id. Only check generators when they are specified direclty in the configuration.
	if len(o.generators) != 0 {
		result.append(t, "TestCheckResourceAttr", "generators.#", strconv.Itoa(len(o.generators)))
		for i, generator := range o.generators {
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.config_style", i), generator.style)
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.section", i), generator.section)
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.template_text", i), generator.templateText)

			if generator.sectionCondition == "" {
				result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("generators.%d.section_condition", i))
			} else {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("generators.%d.section_condition", i), generator.sectionCondition)
			}

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
	}

	return result
}

func TestAccResourceDatacenterConfiglet(t *testing.T) {
	ctx := context.Background()

	// create a blueprint
	bp := testutils.BlueprintA(t, ctx)

	type testStep struct {
		config resourceDatacenterConfiglet
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"multiple_catalog_configlet_ids": {
			steps: []testStep{
				{
					config: resourceDatacenterConfiglet{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						condition:          fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						catalogConfigletId: testutils.CatalogConfigletA(t, ctx, bp.Client()).String(),
					},
				},
				{
					config: resourceDatacenterConfiglet{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						condition:          fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						catalogConfigletId: testutils.CatalogConfigletA(t, ctx, bp.Client()).String(),
					},
				},
				{
					config: resourceDatacenterConfiglet{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						condition:          fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						catalogConfigletId: testutils.CatalogConfigletA(t, ctx, bp.Client()).String(),
					},
				},
			},
		},
		"catalog_generator_catalog_generator": {
			steps: []testStep{
				{
					config: resourceDatacenterConfiglet{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						condition:          fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						catalogConfigletId: testutils.CatalogConfigletA(t, ctx, bp.Client()).String(),
					},
				},
				{
					config: resourceDatacenterConfiglet{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						condition:   fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						generators: []resourceDatacenterConfigletGenerator{
							{
								style:                "junos",
								section:              "top_level_hierarchical",
								templateText:         "set foo",
								negationTemplateText: "del foo",
							},
							{
								style:                "nxos",
								section:              "system",
								templateText:         "foo",
								negationTemplateText: "no foo",
							},
						},
					},
				},
				{
					config: resourceDatacenterConfiglet{
						blueprintId:        bp.Id().String(),
						name:               acctest.RandString(6),
						condition:          fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						catalogConfigletId: testutils.CatalogConfigletA(t, ctx, bp.Client()).String(),
					},
				},
				{
					config: resourceDatacenterConfiglet{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						condition:   fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						generators: []resourceDatacenterConfigletGenerator{
							{
								style:                "sonic",
								section:              "system",
								templateText:         "set foo",
								negationTemplateText: "del foo",
							},
							{
								style:                "eos",
								section:              "system",
								templateText:         "foo",
								negationTemplateText: "no foo",
							},
						},
					},
				},
			},
		},
		"section_condition": {
			steps: []testStep{
				{
					config: resourceDatacenterConfiglet{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						condition:   fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						generators: []resourceDatacenterConfigletGenerator{
							{
								style:                "nxos",
								section:              "interface",
								templateText:         "foo",
								negationTemplateText: "no foo",
								sectionCondition:     `role in ["spine_leaf"]`,
							},
						},
					},
				},
				{
					config: resourceDatacenterConfiglet{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						condition:   fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						generators: []resourceDatacenterConfigletGenerator{
							{
								style:                "eos",
								section:              "interface",
								templateText:         "foo",
								negationTemplateText: "no foo",
								sectionCondition:     `role in ["spine_superspine"]`,
							},
						},
					},
				},
			},
		},
		"add_remove_add_negation": {
			steps: []testStep{
				{
					config: resourceDatacenterConfiglet{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						condition:   fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						generators: []resourceDatacenterConfigletGenerator{
							{
								style:                "nxos",
								section:              "interface",
								templateText:         "foo",
								negationTemplateText: "no foo",
								sectionCondition:     `role in ["spine_leaf"]`,
							},
						},
					},
				},
				{
					config: resourceDatacenterConfiglet{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						condition:   fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						generators: []resourceDatacenterConfigletGenerator{
							{
								style:            "nxos",
								section:          "interface",
								templateText:     "foo",
								sectionCondition: `role in ["spine_leaf"]`,
							},
						},
					},
				},
				{
					config: resourceDatacenterConfiglet{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						condition:   fmt.Sprintf("label in [%q]", acctest.RandString(6)),
						generators: []resourceDatacenterConfigletGenerator{
							{
								style:                "nxos",
								section:              "interface",
								templateText:         "foo",
								negationTemplateText: "no foo",
								sectionCondition:     `role in ["spine_leaf"]`,
							},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterConfiglet)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, bp.Id(), resourceType, tName)

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
