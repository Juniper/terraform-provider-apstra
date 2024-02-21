package tfapstra_test

import (
	"context"
	"fmt"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	resourceDatacenterBlueprintHCL = `
resource "apstra_datacenter_blueprint" "test" {
  name              = %q // mandatory field
  template_id       = %q // mandatory field
  esi_mac_msb       = %s
  ipv6_applications = %s
  fabric_mtu        = %s
  fabric_addressing = %s
}
`
)

func TestResourceDatacenterBlueprint(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)

	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))
	rs := acctest.RandString(6)

	type config struct {
		name             string
		templateId       string
		esiMacMsb        string
		ipv6Applications string
		fabricMtu        string
		fabricAddressing string
	}

	renderConfig := func(cfg config) string {
		return insecureProviderConfigHCL + fmt.Sprintf(resourceDatacenterBlueprintHCL,
			cfg.name,
			cfg.templateId,
			stringOrNull(cfg.esiMacMsb),
			stringOrNull(cfg.ipv6Applications),
			stringOrNull(cfg.fabricMtu),
			stringOrNull(cfg.fabricAddressing),
		)
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		testCase              resource.TestCase
	}

	testCases := map[string]testCase{
		// no version constraints
		// create with default values
		"a": {
			apiVersionConstraints: nil,
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "a0_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "a0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:             "a1_" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "4",
							ipv6Applications: "false",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "a1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:             "a2_" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "6",
							ipv6Applications: "true",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "a2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "a3_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "a3_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:             "a4_" + rs,
							templateId:       "L2_Virtual_EVPN",
							ipv6Applications: "false",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "a4_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
				},
			},
		},

		// no version constraints
		// create with non-default values
		// make no changes
		"b": {
			apiVersionConstraints: nil,
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:             "b0_" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "4",
							ipv6Applications: "true",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "b0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:             "b1_" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "4",
							ipv6Applications: "true",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "b1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
				},
			},
		},

		"c": {
			apiVersionConstraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra411)),
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "c0_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:             "c1_" + rs,
							templateId:       "L2_Virtual_EVPN",
							fabricAddressing: "ipv4",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:             "c2_" + rs,
							templateId:       "L2_Virtual_EVPN",
							fabricAddressing: "ipv4_ipv6",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4_ipv6"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:             "c3_" + rs,
							templateId:       "L2_Virtual_EVPN",
							fabricAddressing: "ipv6",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c3_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv6"),
						}...),
					},
				},
			},
		},

		"d": {
			apiVersionConstraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra420)),
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "d0_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "d1_" + rs,
							templateId: "L2_Virtual_EVPN",
							fabricMtu:  "9170",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "d2_" + rs,
							templateId: "L2_Virtual_EVPN",
							fabricMtu:  "9100",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "d3_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d3_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
						}...),
					},
					{
						Taint: []string{"apstra_datacenter_blueprint.test"},
						Config: renderConfig(config{
							name:       "d4_" + rs,
							templateId: "L2_Virtual_EVPN",
							fabricMtu:  "9100",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d4_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "d5_" + rs,
							templateId: "L2_Virtual_EVPN",
							fabricMtu:  "9100",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d5_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
						}...),
					},
				},
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("API version %s does not satisfy version constraints(%s) of test %q",
					apiVersion, tCase.apiVersionConstraints, tName)
			}
			resource.Test(t, tCase.testCase)
		})
	}
}
