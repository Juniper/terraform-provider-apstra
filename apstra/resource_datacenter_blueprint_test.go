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
  name              = "%s"
  template_id       = "%s"
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
		apiVersionConstrants version.Constraints
		testCase             resource.TestCase
	}

	testCases := map[string]testCase{
		// no version constraints
		// create with default values
		"a": {
			apiVersionConstrants: nil,
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:       "1a-" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "1a-"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "1b-" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "4",
							ipv6Applications: "false",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "1b-"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "1c-" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "6",
							ipv6Applications: "true",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "1c-"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:       "1c-" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "1c-"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "1c-" + rs,
							templateId:       "L2_Virtual_EVPN",
							ipv6Applications: "false",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "1c-"+rs),
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
			apiVersionConstrants: nil,
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "2a-" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "4",
							ipv6Applications: "true",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "2a-"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
						}...),
					},
					{
						Config: insecureProviderConfigHCL + renderConfig(config{
							name:             "2a-" + rs,
							templateId:       "L2_Virtual_EVPN",
							esiMacMsb:        "4",
							ipv6Applications: "true",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "2a-"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
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
			if !tCase.apiVersionConstrants.Check(apiVersion) {
				t.Skip(fmt.Sprintf("API version %s does not satisfy version constraints(%s) of test %q",
					apiVersion.String(), tCase.apiVersionConstrants.String(), tName))
			}
			resource.Test(t, tCase.testCase)
		})
	}
}
