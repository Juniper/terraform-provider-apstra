//go:build integration

package tfapstra_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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

func testCheckIntGE1(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	if !(i >= 1) {
		return errors.New("expected value >= 1, got " + s)
	}

	return nil
}

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
			antiAffinityMode = stringOrNull(cfg.antiAffinityPolicy.Mode.String())
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

	atleast50 := func(s string) error {
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}

		if i < 50 {
			return errors.New("expected value >= 50, got " + s)
		}

		return nil
	}

	type testCase struct {
		apiVersionConstraints version.Constraints
		testCase              resource.TestCase
	}

	testCases := map[string]testCase{
		"evpn_start_minimal_all_versions": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "esMinAV_0_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "esMinAV_1_" + rs,
							templateId: "L2_Virtual_EVPN",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       nil,
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             nil,
							ipv6Applications:                      utils.ToPtr(true),
							junosEvpnMaxNexthopAndInterfaceNumber: nil,
							junosEvpnRoutingInstanceModeMacVrf:    nil,
							junosExOverlayEcmp:                    nil,
							junosGracefulRestart:                  nil,
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          nil,
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "esMinAV_2_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
				},
			},
		},
		"evpn_start_maximal_all_versions": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "esMaxAV_0_" + rs,
							templateId: "L2_Virtual_EVPN",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       nil,
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             nil,
							ipv6Applications:                      utils.ToPtr(false),
							junosEvpnMaxNexthopAndInterfaceNumber: nil,
							junosEvpnRoutingInstanceModeMacVrf:    nil,
							junosExOverlayEcmp:                    nil,
							junosGracefulRestart:                  nil,
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          nil,
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "esMaxAV_1_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "esMaxAV_2_" + rs,
							templateId: "L2_Virtual_EVPN",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          6,
								MaxLinksPerSlot:          10,
								MaxPerSystemLinksPerPort: 4,
								MaxPerSystemLinksPerSlot: 8,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9004),
							defaultSviL3Mtu:                       nil,
							esiMacMsb:                             utils.ToPtr(6),
							evpnType5Routes:                       utils.ToPtr(false),
							fabricMtu:                             nil,
							ipv6Applications:                      utils.ToPtr(true),
							junosEvpnMaxNexthopAndInterfaceNumber: nil,
							junosEvpnRoutingInstanceModeMacVrf:    nil,
							junosExOverlayEcmp:                    nil,
							junosGracefulRestart:                  nil,
							maxEvpnRoutesCount:                    utils.ToPtr(20001),
							maxExternalRoutesCount:                utils.ToPtr(20002),
							maxFabricRoutesCount:                  utils.ToPtr(20003),
							maxMlagRoutesCount:                    utils.ToPtr(20004),
							optimizeRoutingZoneFootprint:          nil,
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "10"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9004"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "20001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "20002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "20003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "20004"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
				},
			},
		},
		"evpn_start_minimal_42x": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "esMin42x_0_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
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
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "esMin42x_1_" + rs,
							templateId: "L2_Virtual_EVPN",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       utils.ToPtr(9004),
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             utils.ToPtr(9006),
							ipv6Applications:                      utils.ToPtr(true),
							junosEvpnMaxNexthopAndInterfaceNumber: utils.ToPtr(false),
							junosEvpnRoutingInstanceModeMacVrf:    utils.ToPtr(false),
							junosExOverlayEcmp:                    utils.ToPtr(false),
							junosGracefulRestart:                  utils.ToPtr(false),
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          utils.ToPtr(false),
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "esMin42x_2_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
				},
			},
		},
		"evpn_start_maximal_42x": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "esMax42x_0_" + rs,
							templateId: "L2_Virtual_EVPN",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       utils.ToPtr(9004),
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             utils.ToPtr(9006),
							ipv6Applications:                      utils.ToPtr(true),
							junosEvpnMaxNexthopAndInterfaceNumber: utils.ToPtr(false),
							junosEvpnRoutingInstanceModeMacVrf:    utils.ToPtr(false),
							junosExOverlayEcmp:                    utils.ToPtr(false),
							junosGracefulRestart:                  utils.ToPtr(false),
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          utils.ToPtr(false),
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "esMax42x_1_" + rs,
							templateId: "L2_Virtual_EVPN",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "esMax42x_2_" + rs,
							templateId: "L2_Virtual_EVPN",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       utils.ToPtr(9004),
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             utils.ToPtr(9006),
							ipv6Applications:                      utils.ToPtr(true),
							junosEvpnMaxNexthopAndInterfaceNumber: utils.ToPtr(false),
							junosEvpnRoutingInstanceModeMacVrf:    utils.ToPtr(false),
							junosExOverlayEcmp:                    utils.ToPtr(false),
							junosGracefulRestart:                  utils.ToPtr(false),
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          utils.ToPtr(false),
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
				},
			},
		},
		"ip_start_minimal_all_versions": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "isMinAV_0_" + rs,
							templateId: "L2_Virtual",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMinAV_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "isMinAV_1_" + rs,
							templateId: "L2_Virtual",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       nil,
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             nil,
							ipv6Applications:                      utils.ToPtr(false),
							junosEvpnMaxNexthopAndInterfaceNumber: nil,
							junosEvpnRoutingInstanceModeMacVrf:    nil,
							junosExOverlayEcmp:                    nil,
							junosGracefulRestart:                  nil,
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          nil,
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMinAV_1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "isMinAV_2_" + rs,
							templateId: "L2_Virtual",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMinAV_2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
				},
			},
		},
		"ip_start_maximal_all_versions": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "isMaxAV_0_" + rs,
							templateId: "L2_Virtual",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       nil,
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             nil,
							ipv6Applications:                      utils.ToPtr(false),
							junosEvpnMaxNexthopAndInterfaceNumber: nil,
							junosEvpnRoutingInstanceModeMacVrf:    nil,
							junosExOverlayEcmp:                    nil,
							junosGracefulRestart:                  nil,
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          nil,
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMaxAV_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "isMaxAV_1_" + rs,
							templateId: "L2_Virtual",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMaxAV_1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "isMaxAV_2_" + rs,
							templateId: "L2_Virtual",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          6,
								MaxLinksPerSlot:          10,
								MaxPerSystemLinksPerPort: 4,
								MaxPerSystemLinksPerSlot: 8,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9004),
							defaultSviL3Mtu:                       nil,
							esiMacMsb:                             utils.ToPtr(6),
							evpnType5Routes:                       utils.ToPtr(false),
							fabricMtu:                             nil,
							ipv6Applications:                      utils.ToPtr(false),
							junosEvpnMaxNexthopAndInterfaceNumber: nil,
							junosEvpnRoutingInstanceModeMacVrf:    nil,
							junosExOverlayEcmp:                    nil,
							junosGracefulRestart:                  nil,
							maxEvpnRoutesCount:                    utils.ToPtr(20001),
							maxExternalRoutesCount:                utils.ToPtr(20002),
							maxFabricRoutesCount:                  utils.ToPtr(20003),
							maxMlagRoutesCount:                    utils.ToPtr(20004),
							optimizeRoutingZoneFootprint:          nil,
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMaxAV_2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "10"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9004"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "20001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "20002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "20003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "20004"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
				},
			},
		},
		"ip_start_minimal_42x": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "isMin42x_0_" + rs,
							templateId: "L2_Virtual",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
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
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "isMin42x_1_" + rs,
							templateId: "L2_Virtual",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       utils.ToPtr(9004),
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             utils.ToPtr(9006),
							ipv6Applications:                      utils.ToPtr(false),
							junosEvpnMaxNexthopAndInterfaceNumber: utils.ToPtr(false),
							junosEvpnRoutingInstanceModeMacVrf:    utils.ToPtr(false),
							junosExOverlayEcmp:                    utils.ToPtr(false),
							junosGracefulRestart:                  utils.ToPtr(false),
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          utils.ToPtr(false),
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "isMin42x_2_" + rs,
							templateId: "L2_Virtual",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
				},
			},
		},
		"ip_start_maximal_42x": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:       "isMax42x_0_" + rs,
							templateId: "L2_Virtual",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       utils.ToPtr(9004),
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             utils.ToPtr(9006),
							ipv6Applications:                      utils.ToPtr(false),
							junosEvpnMaxNexthopAndInterfaceNumber: utils.ToPtr(false),
							junosEvpnRoutingInstanceModeMacVrf:    utils.ToPtr(false),
							junosExOverlayEcmp:                    utils.ToPtr(false),
							junosGracefulRestart:                  utils.ToPtr(false),
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          utils.ToPtr(false),
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "isMax42x_1_" + rs,
							templateId: "L2_Virtual",
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_1_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
					{
						Config: renderConfig(config{
							name:       "isMax42x_2_" + rs,
							templateId: "L2_Virtual",
							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
								MaxLinksPerPort:          4,
								MaxLinksPerSlot:          8,
								MaxPerSystemLinksPerPort: 2,
								MaxPerSystemLinksPerSlot: 6,
								Mode:                     apstra.AntiAffinityModeEnabledStrict,
							},
							defaultIpLinksToGenericMtu:            utils.ToPtr(9002),
							defaultSviL3Mtu:                       utils.ToPtr(9004),
							esiMacMsb:                             utils.ToPtr(4),
							evpnType5Routes:                       utils.ToPtr(true),
							fabricMtu:                             utils.ToPtr(9006),
							ipv6Applications:                      utils.ToPtr(false),
							junosEvpnMaxNexthopAndInterfaceNumber: utils.ToPtr(false),
							junosEvpnRoutingInstanceModeMacVrf:    utils.ToPtr(false),
							junosExOverlayEcmp:                    utils.ToPtr(false),
							junosGracefulRestart:                  utils.ToPtr(false),
							maxEvpnRoutesCount:                    utils.ToPtr(10001),
							maxExternalRoutesCount:                utils.ToPtr(10002),
							maxFabricRoutesCount:                  utils.ToPtr(10003),
							maxMlagRoutesCount:                    utils.ToPtr(10004),
							optimizeRoutingZoneFootprint:          utils.ToPtr(false),
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_2_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
						}...),
					},
				},
			},
		},
		"evpn_start_with_ipv6": {
			testCase: resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: renderConfig(config{
							name:             "esv6AV_0_" + rs,
							templateId:       "L2_Virtual_EVPN",
							ipv6Applications: utils.ToPtr(true),
						}),
						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esv6AV_0_"+rs),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
						}...),
					},
				},
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			// t.Parallel()
			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("API version %s does not satisfy version constraints(%s) of test %q",
					apiVersion, tCase.apiVersionConstraints, tName)
			}
			resource.Test(t, tCase.testCase)
		})
	}
}
