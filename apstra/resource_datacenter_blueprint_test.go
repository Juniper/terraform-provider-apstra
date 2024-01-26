package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
	"time"
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

	apstraUrl, ok := os.LookupEnv(utils.EnvApstraUrl)
	if !ok || apstraUrl == "" {
		t.Fatalf("apstra url environment variable (%s) must be set and non-empty", utils.EnvApstraUrl)
	}

	clientCfg, err := utils.NewClientConfig(apstraUrl)
	if err != nil {
		t.Fatal(err)
	}

	client, err := clientCfg.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}

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
		return fmt.Sprintf(resourceDatacenterBlueprintHCL,
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
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:       "a1_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "a1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "a2_" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "4",
							ipv6Applications: "false",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "a2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "a3_" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "6",
							ipv6Applications: "true",
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
						Config: insecureProviderConfigHCL + renderConfig(config{
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
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "a3_" + rs,
							templateId:       "L2_Virtual_EVPN",
							ipv6Applications: "false",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "a3_"+rs),
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
						Config: insecureProviderConfigHCL + renderConfig(config{
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
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
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

		// version 4.1.1 and later
		"c": {
			apiVersionConstraints: version.MustConstraints(version.NewConstraint(">=4.1.1")),
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:       "c1_" + rs,
							templateId: "L2_Virtual_EVPN",
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
						Config: insecureProviderConfigHCL + renderConfig(config{
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
						Config: insecureProviderConfigHCL + renderConfig(config{
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
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "c2_" + rs,
							templateId:       "L2_Virtual_EVPN",
							fabricAddressing: "ipv6",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv6"),
						}...),
					},
				},
			},
		},

		// version 4.2.0 and later
		"d": {
			apiVersionConstraints: version.MustConstraints(version.NewConstraint(">=4.2.0")),
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:       "d1_" + rs,
							templateId: "L2_Virtual_EVPN",
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
						Config: insecureProviderConfigHCL + renderConfig(config{
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
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:       "d1_" + rs,
							templateId: "L2_Virtual_EVPN",
							fabricMtu:  "9100",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
						}...),
					},
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:       "d1_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
						}...),
					},
					{
						Taint: []string{"apstra_datacenter_blueprint.test"},
						Config: insecureProviderConfigHCL + renderConfig(config{
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
						Config: insecureProviderConfigHCL + renderConfig(config{
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
				},
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Logf("%s test case %q", time.Now().String(), tName)
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("API version %s does not satisfy version constraints(%s) of test %q",
					apiVersion.String(), tCase.apiVersionConstraints.String(), tName)
			}
			resource.Test(t, tCase.testCase)
		})
	}
}
