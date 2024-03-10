//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"strings"
	"testing"
)

const (
	resourceTemplateRackBasedHCL = `
resource "apstra_template_rack_based" "test" {
  name                     = %q // mandatory field
  asn_allocation_scheme    = %q // mandatory field
  overlay_control_protocol = %q // mandatory field
  rack_infos               = %s // mandatory field
  spine                    = %s // mandatory field
  fabric_link_addressing   = %s
}
`
	resourceTemplateRackBasedRackInfoHcl = `
    %q = { count = %d }
`
	resourceTemplateRackBasedSpineHcl = `{
      count             = %d
      logical_device_id = %q
    }`
)

func TestResourceTemplateRackBased(t *testing.T) {
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

	type spine struct {
		count           int
		logicalDeviceId string
	}

	renderSpine := func(spine spine) string {
		return fmt.Sprintf(resourceTemplateRackBasedSpineHcl, spine.count, spine.logicalDeviceId)
	}

	renderRackInfos := func(rackInfos map[string]int) string {
		var sb strings.Builder
		sb.WriteString("{")
		for k, v := range rackInfos {
			sb.WriteString(fmt.Sprintf(resourceTemplateRackBasedRackInfoHcl, k, v))
		}
		sb.WriteString("  }")
		return sb.String()
	}

	type config struct {
		name                   string
		asnAllocationScheme    string
		overlayControlProtocol string
		rackInfos              map[string]int
		spine                  spine
		fabricLinkAddressing   string
	}

	renderConfig := func(config config) string {
		return insecureProviderConfigHCL + fmt.Sprintf(resourceTemplateRackBasedHCL,
			config.name,
			config.asnAllocationScheme,
			config.overlayControlProtocol,
			renderRackInfos(config.rackInfos),
			renderSpine(config.spine),
			stringOrNull(config.fabricLinkAddressing),
		)
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		testCase              resource.TestCase
	}

	testCases := map[string]testCase{
		"a": {
			apiVersionConstraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra411)),
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:                   "a1_" + rs,
							asnAllocationScheme:    "unique",
							overlayControlProtocol: "evpn",
							rackInfos: map[string]int{
								"L2_Virtual": 1,
							},
							spine: spine{
								count:           1,
								logicalDeviceId: "AOS-7x10-Spine",
							},
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_template_rack_based.test", "id"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "name", "a1_"+rs),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "asn_allocation_scheme", "unique"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "overlay_control_protocol", "evpn"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "spine.count", "1"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "spine.logical_device_id", "AOS-7x10-Spine"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "rack_infos.%", "1"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "rack_infos.L2_Virtual.count", "1"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:                   "a2_" + rs,
							asnAllocationScheme:    "single",
							overlayControlProtocol: "static",
							rackInfos: map[string]int{
								"L2_Virtual_Dual_2x_Links": 2,
							},
							spine: spine{
								count:           2,
								logicalDeviceId: "AOS-4x10-1",
							},
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_template_rack_based.test", "id"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "name", "a2_"+rs),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "asn_allocation_scheme", "single"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "overlay_control_protocol", "static"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "spine.count", "2"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "spine.logical_device_id", "AOS-4x10-1"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "rack_infos.%", "1"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "rack_infos.L2_Virtual_Dual_2x_Links.count", "2"),
						}...),
					},
				},
			},
		},

		"b": {
			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.Apstra410)),
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:                   "b1_" + rs,
							asnAllocationScheme:    "unique",
							overlayControlProtocol: "evpn",
							rackInfos: map[string]int{
								"L2_Virtual": 1,
							},
							spine: spine{
								count:           1,
								logicalDeviceId: "AOS-7x10-Spine",
							},
							fabricLinkAddressing: "ipv4",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_template_rack_based.test", "id"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "name", "b1_"+rs),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "asn_allocation_scheme", "unique"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "overlay_control_protocol", "evpn"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "fabric_link_addressing", "ipv4"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "spine.count", "1"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "spine.logical_device_id", "AOS-7x10-Spine"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "rack_infos.%", "1"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "rack_infos.L2_Virtual.count", "1"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:                   "b2_" + rs,
							asnAllocationScheme:    "single",
							overlayControlProtocol: "static",
							rackInfos: map[string]int{
								"L2_Virtual_Dual_2x_Links": 2,
							},
							spine: spine{
								count:           2,
								logicalDeviceId: "AOS-4x10-1",
							},
							fabricLinkAddressing: "ipv6",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_template_rack_based.test", "id"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "name", "b2_"+rs),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "asn_allocation_scheme", "single"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "overlay_control_protocol", "static"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "fabric_link_addressing", "ipv6"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "spine.count", "2"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "spine.logical_device_id", "AOS-4x10-1"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "rack_infos.%", "1"),
							resource.TestCheckResourceAttr("apstra_template_rack_based.test", "rack_infos.L2_Virtual_Dual_2x_Links.count", "2"),
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
			t.Logf("testing against Apstra %s", apiVersion)
			resource.Test(t, tCase.testCase)
		})
	}
}
