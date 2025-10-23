//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const resourceDatacenterRoutingPolicyHCL = `resource %q %q {
  blueprint_id        = %q
  name                = %q
  description         = %s
  import_policy       = %s
  extra_imports       = %s
  extra_exports       = %s
  export_policy       = %s
  aggregate_prefixes  = %s
  expect_default_ipv4 = %s
  expect_default_ipv6 = %s
}
`

type resourceDatacenterRoutingPolicy struct {
	name              string
	description       string
	importPolicy      *apstra.DcRoutingPolicyImportPolicy
	extraImports      []resourceDatacenterRoutingPolicyExtraImportExport
	extraExports      []resourceDatacenterRoutingPolicyExtraImportExport
	exportPolicy      *resourceDatacenterRoutingPolicyExportPolicy
	aggregatePrefixes []net.IPNet
	expectDefaultIPv4 *bool
	expectDefaultIPv6 *bool
}

func (o resourceDatacenterRoutingPolicy) render(rType, rName, bpID string) string {
	extraImports := new(strings.Builder)
	if len(o.extraImports) == 0 {
		extraImports.WriteString("null\n")
	} else {
		extraImports.WriteString("[\n")
		for _, v := range o.extraImports {
			extraImports.WriteString(fmt.Sprintf("\n%s", v.render()))
		}
		extraImports.WriteString("]\n")
	}

	extraExports := new(strings.Builder)
	if len(o.extraExports) == 0 {
		extraExports.WriteString("null\n")
	} else {
		extraExports.WriteString("[\n")
		for _, v := range o.extraExports {
			extraExports.WriteString(fmt.Sprintf("\n%s", v.render()))
		}
		extraExports.WriteString("]\n")
	}

	var aggregatePrefixes []string
	if o.aggregatePrefixes != nil {
		aggregatePrefixes = make([]string, len(o.aggregatePrefixes))
		for i, v := range o.aggregatePrefixes {
			aggregatePrefixes[i] = v.String()
		}
	}

	return fmt.Sprintf(resourceDatacenterRoutingPolicyHCL, rType, rName,
		bpID,
		o.name,
		stringOrNull(o.description),
		stringerOrNull(o.importPolicy),
		extraImports.String(),
		extraExports.String(),
		o.exportPolicy.render(),
		stringSliceOrNull(aggregatePrefixes),
		boolPtrOrNull(o.expectDefaultIPv4),
		boolPtrOrNull(o.expectDefaultIPv6),
	)
}

func (o resourceDatacenterRoutingPolicy) testChecks(t testing.TB, bpID, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpID)
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	if o.description == "" {
		result.append(t, "TestCheckNoResourceAttr", "description")
	} else {
		result.append(t, "TestCheckResourceAttr", "description", o.description)
	}

	if o.importPolicy == nil {
		result.append(t, "TestCheckResourceAttr", "import_policy", apstra.DcRoutingPolicyImportPolicyDefaultOnly.String())
	} else {
		result.append(t, "TestCheckResourceAttr", "import_policy", o.importPolicy.String())
	}

	if o.extraImports == nil {
		result.append(t, "TestCheckNoResourceAttr", "extra_imports")
	} else {
		result.append(t, "TestCheckResourceAttr", "extra_imports.#", strconv.Itoa(len(o.extraImports)))
		for i, extra := range o.extraImports {
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_imports.%d.prefix", i), extra.prefix)
			if extra.geMask == nil {
				result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("extra_imports.%d.ge_mask", i))
			} else {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_imports.%d.ge_mask", i), strconv.Itoa(*extra.geMask))
			}
			if extra.leMask == nil {
				result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("extra_imports.%d.le_mask", i))
			} else {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_imports.%d.le_mask", i), strconv.Itoa(*extra.leMask))
			}
			if extra.action == nil {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_imports.%d.action", i), apstra.PrefixFilterActionPermit.String())
			} else {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_imports.%d.action", i), extra.action.String())
			}
		}
	}

	if o.extraExports == nil {
		result.append(t, "TestCheckNoResourceAttr", "extra_exports")
	} else {
		result.append(t, "TestCheckResourceAttr", "extra_exports.#", strconv.Itoa(len(o.extraExports)))
		for i, extra := range o.extraExports {
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_exports.%d.prefix", i), extra.prefix)
			if extra.geMask == nil {
				result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("extra_exports.%d.ge_mask", i))
			} else {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_exports.%d.ge_mask", i), strconv.Itoa(*extra.geMask))
			}
			if extra.leMask == nil {
				result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("extra_exports.%d.le_mask", i))
			} else {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_exports.%d.le_mask", i), strconv.Itoa(*extra.leMask))
			}
			if extra.action == nil {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_exports.%d.action", i), apstra.PrefixFilterActionPermit.String())
			} else {
				result.append(t, "TestCheckResourceAttr", fmt.Sprintf("extra_exports.%d.action", i), extra.action.String())
			}
		}
	}

	if o.exportPolicy == nil {
		result.append(t, "TestCheckResourceAttr", "export_policy.export_spine_leaf_links", "false")
		result.append(t, "TestCheckResourceAttr", "export_policy.export_spine_superspine_links", "false")
		result.append(t, "TestCheckResourceAttr", "export_policy.export_l3_edge_server_links", "false")
		result.append(t, "TestCheckResourceAttr", "export_policy.export_l2_edge_subnets", "false")
		result.append(t, "TestCheckResourceAttr", "export_policy.export_loopbacks", "false")
		result.append(t, "TestCheckResourceAttr", "export_policy.export_static_routes", "false")
	} else {
		if o.exportPolicy.SpineLeafLinks == nil {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_spine_leaf_links", "false")
		} else {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_spine_leaf_links", strconv.FormatBool(*o.exportPolicy.SpineLeafLinks))
		}
		if o.exportPolicy.SpineSuperspineLinks == nil {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_spine_superspine_links", "false")
		} else {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_spine_superspine_links", strconv.FormatBool(*o.exportPolicy.SpineSuperspineLinks))
		}
		if o.exportPolicy.L3EdgeServerLinks == nil {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_l3_edge_server_links", "false")
		} else {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_l3_edge_server_links", strconv.FormatBool(*o.exportPolicy.L3EdgeServerLinks))
		}
		if o.exportPolicy.L2EdgeSubnets == nil {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_l2_edge_subnets", "false")
		} else {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_l2_edge_subnets", strconv.FormatBool(*o.exportPolicy.L2EdgeSubnets))
		}
		if o.exportPolicy.Loopbacks == nil {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_loopbacks", "false")
		} else {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_loopbacks", strconv.FormatBool(*o.exportPolicy.Loopbacks))
		}
		if o.exportPolicy.StaticRoutes == nil {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_static_routes", "false")
		} else {
			result.append(t, "TestCheckResourceAttr", "export_policy.export_static_routes", strconv.FormatBool(*o.exportPolicy.StaticRoutes))
		}
	}

	if o.aggregatePrefixes == nil {
		result.append(t, "TestCheckNoResourceAttr", "aggregate_prefixes")
	} else {
		result.append(t, "TestCheckResourceAttr", "aggregate_prefixes.#", strconv.Itoa(len(o.aggregatePrefixes)))
		for i, aggregatePrefix := range o.aggregatePrefixes {
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("aggregate_prefixes.%d", i), aggregatePrefix.String())
		}
	}

	if o.expectDefaultIPv4 == nil {
		result.append(t, "TestCheckResourceAttr", "expect_default_ipv4", "true")
	} else {
		result.append(t, "TestCheckResourceAttr", "expect_default_ipv4", strconv.FormatBool(*o.expectDefaultIPv4))
	}

	if o.expectDefaultIPv6 == nil {
		result.append(t, "TestCheckResourceAttr", "expect_default_ipv6", "true")
	} else {
		result.append(t, "TestCheckResourceAttr", "expect_default_ipv6", strconv.FormatBool(*o.expectDefaultIPv6))
	}

	return result
}

const resourceDatacenterRoutingPolicyExportPolicyHCL = `{
    export_spine_leaf_links       = %s
    export_spine_superspine_links = %s
    export_l3_edge_server_links   = %s
    export_l2_edge_subnets        = %s
    export_loopbacks              = %s
    export_static_routes          = %s
  }`

type resourceDatacenterRoutingPolicyExportPolicy struct {
	SpineLeafLinks       *bool
	SpineSuperspineLinks *bool
	L3EdgeServerLinks    *bool
	L2EdgeSubnets        *bool
	Loopbacks            *bool
	StaticRoutes         *bool
}

func (o *resourceDatacenterRoutingPolicyExportPolicy) render() string {
	if o == nil {
		return "null"
	}

	return fmt.Sprintf(resourceDatacenterRoutingPolicyExportPolicyHCL,
		boolPtrOrNull(o.SpineLeafLinks),
		boolPtrOrNull(o.SpineSuperspineLinks),
		boolPtrOrNull(o.L3EdgeServerLinks),
		boolPtrOrNull(o.L2EdgeSubnets),
		boolPtrOrNull(o.Loopbacks),
		boolPtrOrNull(o.StaticRoutes),
	)
}

const resourceDatacenterRoutingPolicyExtraImportExportHCL = `    {
      prefix  = %q
      ge_mask = %s
      le_mask = %s
      action  = %s
    },
`

type resourceDatacenterRoutingPolicyExtraImportExport struct {
	prefix string
	geMask *int
	leMask *int
	action *apstra.PrefixFilterAction
}

func (o *resourceDatacenterRoutingPolicyExtraImportExport) render() string {
	if o == nil {
		return "null"
	}

	return fmt.Sprintf(resourceDatacenterRoutingPolicyExtraImportExportHCL,
		o.prefix,
		intPtrOrNull(o.geMask),
		intPtrOrNull(o.leMask),
		stringerOrNull(o.action),
	)
}

func TestResourceDatacenteRoutingPolicy(t *testing.T) {
	ctx := context.Background()

	// create a blueprint
	bp := testutils.BlueprintA(t, ctx)

	type testStep struct {
		config      resourceDatacenterRoutingPolicy
		expectError *regexp.Regexp
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"l3_edge_okay": {
			versionConstraints: compatibility.RoutingPolicyExportL3EdgeServerOK.Constraints,
			steps: []testStep{
				{
					config: resourceDatacenterRoutingPolicy{
						name:         acctest.RandString(6),
						exportPolicy: &resourceDatacenterRoutingPolicyExportPolicy{L3EdgeServerLinks: utils.ToPtr(true)},
					},
				},
			},
		},
		"l3_edge_not_okay": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GtApstra422)),
			steps: []testStep{
				{
					config: resourceDatacenterRoutingPolicy{
						name:         acctest.RandString(6),
						exportPolicy: &resourceDatacenterRoutingPolicyExportPolicy{L3EdgeServerLinks: utils.ToPtr(true)},
					},
					expectError: regexp.MustCompile("This configuration requires Apstra <=4.2.2"),
				},
			},
		},
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceDatacenterRoutingPolicy{
						name: acctest.RandString(6),
					},
				},
				{
					config: resourceDatacenterRoutingPolicy{
						name:         acctest.RandString(6),
						description:  acctest.RandString(10),
						importPolicy: utils.ToPtr(apstra.DcRoutingPolicyImportPolicyExtraOnly),
						extraImports: []resourceDatacenterRoutingPolicyExtraImportExport{
							{prefix: "11.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
							{prefix: "12.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
						},
						extraExports: []resourceDatacenterRoutingPolicyExtraImportExport{
							{prefix: "21.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
							{prefix: "22.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
						},
						exportPolicy: &resourceDatacenterRoutingPolicyExportPolicy{
							SpineLeafLinks:       utils.ToPtr(true),
							SpineSuperspineLinks: utils.ToPtr(false),
							L3EdgeServerLinks:    nil, // not valid in 5.0.0 and later
							L2EdgeSubnets:        utils.ToPtr(true),
							Loopbacks:            utils.ToPtr(false),
							StaticRoutes:         nil,
						},
						expectDefaultIPv4: utils.ToPtr(true),
						expectDefaultIPv6: utils.ToPtr(false),
					},
				},
				{
					config: resourceDatacenterRoutingPolicy{
						name: acctest.RandString(6),
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: resourceDatacenterRoutingPolicy{
						name:         acctest.RandString(6),
						description:  acctest.RandString(10),
						importPolicy: utils.ToPtr(apstra.DcRoutingPolicyImportPolicyExtraOnly),
						extraImports: []resourceDatacenterRoutingPolicyExtraImportExport{
							{prefix: "11.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
							{prefix: "12.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
						},
						extraExports: []resourceDatacenterRoutingPolicyExtraImportExport{
							{prefix: "21.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
							{prefix: "22.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
						},
						exportPolicy: &resourceDatacenterRoutingPolicyExportPolicy{
							SpineLeafLinks:       utils.ToPtr(true),
							SpineSuperspineLinks: utils.ToPtr(false),
							L3EdgeServerLinks:    nil, // not valid in 5.0.0 and later
							L2EdgeSubnets:        utils.ToPtr(true),
							Loopbacks:            utils.ToPtr(false),
							StaticRoutes:         nil,
						},
						expectDefaultIPv4: utils.ToPtr(true),
						expectDefaultIPv6: utils.ToPtr(false),
					},
				},
				{
					config: resourceDatacenterRoutingPolicy{
						name: acctest.RandString(6),
					},
				},
				{
					config: resourceDatacenterRoutingPolicy{
						name:         acctest.RandString(6),
						description:  acctest.RandString(10),
						importPolicy: utils.ToPtr(apstra.DcRoutingPolicyImportPolicyExtraOnly),
						extraImports: []resourceDatacenterRoutingPolicyExtraImportExport{
							{prefix: "31.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
							{prefix: "32.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
						},
						extraExports: []resourceDatacenterRoutingPolicyExtraImportExport{
							{prefix: "41.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
							{prefix: "42.2.3.0/24", geMask: utils.ToPtr(25), leMask: utils.ToPtr(30), action: utils.ToPtr(apstra.PrefixFilterActionDeny)},
						},
						exportPolicy: &resourceDatacenterRoutingPolicyExportPolicy{
							SpineLeafLinks:       utils.ToPtr(false),
							SpineSuperspineLinks: utils.ToPtr(true),
							L3EdgeServerLinks:    nil, // not valid in 5.0.0 and later
							L2EdgeSubnets:        utils.ToPtr(false),
							Loopbacks:            utils.ToPtr(true),
							StaticRoutes:         nil,
						},
						expectDefaultIPv4: nil,
						expectDefaultIPv6: utils.ToPtr(true),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterRoutingPolicy)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName, bp.Id().String())
				checks := step.config.testChecks(t, bp.Id().String(), resourceType, tName)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config:      insecureProviderConfigHCL + config,
					Check:       resource.ComposeAggregateTestCheckFunc(checks.checks...),
					ExpectError: step.expectError,
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
