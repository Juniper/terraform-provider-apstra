package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"

	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceDatacenterBlueprintHCL = `
resource "apstra_datacenter_blueprint" "test" {
  name                                        = %q // mandatory field
  template_id                                 = %q // mandatory field
  fabric_addressing                           = %s

  anti_affinity_mode                          = %s
  anti_affinity_policy                        = %s
  default_ip_links_to_generic_mtu             = %s
  default_svi_l3_mtu                          = %s
  esi_mac_msb                                 = %s
  evpn_type_5_routes                          = %s
  fabric_mtu                                  = %s
  ipv6_applications                           = %s
  junos_evpn_max_nexthop_and_interface_number = %s
  junos_evpn_routing_instance_mode_mac_vrf    = %s
  junos_ex_overlay_ecmp                       = %s
  junos_graceful_restart                      = %s
  max_evpn_routes_count                       = %s
  max_external_routes_count                   = %s
  max_fabric_routes_count                     = %s
  max_mlag_routes_count                       = %s
  optimize_routing_zone_footprint             = %s
}
`

	resourceDatacenterBlueprintAntiAffinityPolicyHCL = `{
    max_links_count_per_slot            = %s
	max_links_count_per_system_per_slot = %s
	max_links_count_per_port            = %s
	max_links_count_per_system_per_port = %s
  }
`
)

func TestResourceDatacenterBlueprint(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)

	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))
	rs := acctest.RandString(6)

	type config struct {
		name                                  string
		templateId                            string
		fabricAddressing                      *apstra.AddressingScheme
		antiAffinityPolicy                    *apstra.AntiAffinityPolicy
		defaultIpLinksToGenericMtu            *int
		defaultSviL3Mtu                       *int
		esiMacMsb                             *int
		evpnType5Routes                       *bool
		fabricMtu                             *int
		ipv6Applications                      *bool
		junosEvpnMaxNexthopAndInterfaceNumber *bool
		junosEvpnRoutingInstanceModeMacVrf    *bool
		junosExOverlayEcmp                    *bool
		junosGracefulRestart                  *bool
		maxEvpnRoutesCount                    *int
		maxExternalRoutesCount                *int
		maxFabricRoutesCount                  *int
		maxMlagRoutesCount                    *int
		optimizeRoutingZoneFootprint          *bool
	}

	renderAntiAffinityPolicy := func(cfg *apstra.AntiAffinityPolicy) string {
		if cfg == nil {
			return "null"
		}

		return fmt.Sprintf(resourceDatacenterBlueprintAntiAffinityPolicyHCL,
			fmt.Sprintf("%d", cfg.MaxLinksPerSlot),
			fmt.Sprintf("%d", cfg.MaxPerSystemLinksPerSlot),
			fmt.Sprintf("%d", cfg.MaxLinksPerPort),
			fmt.Sprintf("%d", cfg.MaxPerSystemLinksPerPort),
		)
	}

	renderConfig := func(cfg config) string {
		fabricAddressing := "null"
		if cfg.fabricAddressing != nil {
			fabricAddressing = cfg.fabricAddressing.String()
		}

		antiAffinityMode, antiAffinitiyPolicy := "null", "null"
		if cfg.antiAffinityPolicy != nil {
			antiAffinityMode = cfg.antiAffinityPolicy.Mode.String()
			antiAffinitiyPolicy = renderAntiAffinityPolicy(cfg.antiAffinityPolicy)
		}

		return insecureProviderConfigHCL + fmt.Sprintf(resourceDatacenterBlueprintHCL,
			cfg.name,
			cfg.templateId,
			fabricAddressing,
			antiAffinityMode,
			antiAffinitiyPolicy,
			intPtrOrNull(cfg.defaultIpLinksToGenericMtu),
			intPtrOrNull(cfg.defaultSviL3Mtu),
			intPtrOrNull(cfg.esiMacMsb),
			boolPtrOrNull(cfg.evpnType5Routes),
			intPtrOrNull(cfg.fabricMtu),
			boolPtrOrNull(cfg.ipv6Applications),
			boolPtrOrNull(cfg.junosEvpnMaxNexthopAndInterfaceNumber),
			boolPtrOrNull(cfg.junosEvpnRoutingInstanceModeMacVrf),
			boolPtrOrNull(cfg.junosExOverlayEcmp),
			boolPtrOrNull(cfg.junosGracefulRestart),
			intPtrOrNull(cfg.maxEvpnRoutesCount),
			intPtrOrNull(cfg.maxExternalRoutesCount),
			intPtrOrNull(cfg.maxFabricRoutesCount),
			intPtrOrNull(cfg.maxMlagRoutesCount),
			boolPtrOrNull(cfg.optimizeRoutingZoneFootprint),
		)
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		testCase              resource.TestCase
	}

	testCases := map[string]testCase{
		// no version constraints
		// create with default values
		"start_minimal": {
			apiVersionConstraints: apiversions.Ge420,
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
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "version", "1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_errors_count", "52"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", "disabled"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "0"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "0"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "0"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
				},
			},
		},

		//// no version constraints
		//// create with non-default values
		//// make no changes
		//"b": {
		//	apiVersionConstraints: nil,
		//	testCase: resource.TestCase{
		//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		//		Steps: []resource.TestStep{
		//			{
		//				Config: renderConfig(config{
		//					name:             "b0_" + rs,
		//					templateId:       "L2_Virtual_EVPN",
		//					esiMacMsb:        "4",
		//					ipv6Applications: "true",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "b0_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
		//				}...),
		//			},
		//			{
		//				Config: renderConfig(config{
		//					name:             "b1_" + rs,
		//					templateId:       "L2_Virtual_EVPN",
		//					esiMacMsb:        "4",
		//					ipv6Applications: "true",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "b1_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
		//				}...),
		//			},
		//		},
		//	},
		//},

		//"c": {
		//	apiVersionConstraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra411)),
		//	testCase: resource.TestCase{
		//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		//		Steps: []resource.TestStep{
		//			{
		//				Config: renderConfig(config{
		//					name:       "c0_" + rs,
		//					templateId: "L2_Virtual_EVPN",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c0_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//				}...),
		//			},
		//			{
		//				Config: renderConfig(config{
		//					name:             "c1_" + rs,
		//					templateId:       "L2_Virtual_EVPN",
		//					fabricAddressing: "ipv4",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c1_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
		//				}...),
		//			},
		//			{
		//				Config: renderConfig(config{
		//					name:             "c2_" + rs,
		//					templateId:       "L2_Virtual_EVPN",
		//					fabricAddressing: "ipv4_ipv6",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c2_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4_ipv6"),
		//				}...),
		//			},
		//			{
		//				Config: renderConfig(config{
		//					name:             "c3_" + rs,
		//					templateId:       "L2_Virtual_EVPN",
		//					fabricAddressing: "ipv4",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "c3_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_addressing", "ipv4"),
		//				}...),
		//			},
		//		},
		//	},
		//},

		//"d": {
		//	apiVersionConstraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra420)),
		//	testCase: resource.TestCase{
		//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		//		Steps: []resource.TestStep{
		//			{
		//				Config: renderConfig(config{
		//					name:       "d0_" + rs,
		//					templateId: "L2_Virtual_EVPN",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d0_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
		//				}...),
		//			},
		//			{
		//				Config: renderConfig(config{
		//					name:       "d1_" + rs,
		//					templateId: "L2_Virtual_EVPN",
		//					fabricMtu:  "9170",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d1_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
		//				}...),
		//			},
		//			{
		//				Config: renderConfig(config{
		//					name:       "d2_" + rs,
		//					templateId: "L2_Virtual_EVPN",
		//					fabricMtu:  "9100",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d2_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
		//				}...),
		//			},
		//			{
		//				Config: renderConfig(config{
		//					name:       "d3_" + rs,
		//					templateId: "L2_Virtual_EVPN",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d3_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
		//				}...),
		//			},
		//			{
		//				Taint: []string{"apstra_datacenter_blueprint.test"},
		//				Config: renderConfig(config{
		//					name:       "d4_" + rs,
		//					templateId: "L2_Virtual_EVPN",
		//					fabricMtu:  "9100",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d4_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
		//				}...),
		//			},
		//			{
		//				Config: renderConfig(config{
		//					name:       "d5_" + rs,
		//					templateId: "L2_Virtual_EVPN",
		//					fabricMtu:  "9100",
		//				}),
		//				Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		//					resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "d5_"+rs),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
		//					resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9100"),
		//				}...),
		//			},
		//		},
		//	},
		//},
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
