//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const resourceDatacenterBlueprintHCL = `resource %q %q {
  name                                        = %q
  template_id                                 = %q

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
  underlay_addressing                         = %s
  vtep_addressing                             = %s
  disable_ipv4                                = %s
}
`

const datasourceDatacenterBlueprintHCL = `
data %q %q {
  id   = %s
  name = %s
}
`

type resourceDatacenterBlueprint struct {
	name                                  string
	templateID                            string
	fabricAddressing                      *enum.AddressingScheme
	antiAffinityMode                      *enum.AntiAffinityMode
	antiAffinityPolicy                    *resourceDatacenterBlueprintAntiAffinityPolicy
	defaultIPLinksToGenericMTU            *int
	defaultSVIL3MTU                       *int
	esiMACMSB                             *int
	evpnType5Routes                       *bool
	fabricMTU                             *int
	ipv6Applications                      *bool
	junosEVPNMaxNexthopAndInterfaceNumber *bool
	junosEVPNRoutingInstanceModeMACVRF    *bool
	junosEXOverlayECMP                    *bool
	junosGracefulRestart                  *bool
	maxEVPNRoutesCount                    *int
	maxExternalRoutesCount                *int
	maxFabricRoutesCount                  *int
	maxMLAGRoutesCount                    *int
	optimizeRoutingZoneFootprint          *bool
	underlayAddressing                    *enum.AddressingScheme
	vtepAddressing                        *enum.AddressingScheme
	disableIPv4                           *bool
}

func (o resourceDatacenterBlueprint) render(rType, rName string) string {
	resource := fmt.Sprintf(resourceDatacenterBlueprintHCL,
		rType, rName,
		o.name,
		o.templateID,
		stringerOrNull(o.fabricAddressing),
		stringerOrNull(o.antiAffinityMode),
		o.antiAffinityPolicy.render(),
		intPtrOrNull(o.defaultIPLinksToGenericMTU),
		intPtrOrNull(o.defaultSVIL3MTU),
		intPtrOrNull(o.esiMACMSB),
		boolPtrOrNull(o.evpnType5Routes),
		intPtrOrNull(o.fabricMTU),
		boolPtrOrNull(o.ipv6Applications),
		boolPtrOrNull(o.junosEVPNMaxNexthopAndInterfaceNumber),
		boolPtrOrNull(o.junosEVPNRoutingInstanceModeMACVRF),
		boolPtrOrNull(o.junosEXOverlayECMP),
		boolPtrOrNull(o.junosGracefulRestart),
		intPtrOrNull(o.maxEVPNRoutesCount),
		intPtrOrNull(o.maxExternalRoutesCount),
		intPtrOrNull(o.maxFabricRoutesCount),
		intPtrOrNull(o.maxMLAGRoutesCount),
		boolPtrOrNull(o.optimizeRoutingZoneFootprint),
		stringerOrNull(o.underlayAddressing),
		stringerOrNull(o.vtepAddressing),
		boolPtrOrNull(o.disableIPv4),
	)

	datasourceByID := fmt.Sprintf(datasourceDatacenterBlueprintHCL, rType, rName+"_by_id", fmt.Sprintf("%s.%s.id", rType, rName), "null")
	datasourceByName := fmt.Sprintf(datasourceDatacenterBlueprintHCL, rType, rName+"_by_name", "null", fmt.Sprintf("%s.%s.name", rType, rName))

	return resource + datasourceByID + datasourceByName
}

func (o resourceDatacenterBlueprint) testChecks(t testing.TB, rType, rName string) []testChecks {
	resourceChecks := newTestChecks(rType + "." + rName)
	dataByIDChecks := newTestChecks("data." + rType + "." + rName + "_by_id")
	dataByNameChecks := newTestChecks("data." + rType + "." + rName + "_by_name")

	// required and computed attributes can always be checked
	resourceChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	resourceChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	resourceChecks.append(t, "TestCheckResourceAttr", "template_id", o.templateID)

	// optional+computed attributes
	if o.fabricAddressing != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "fabric_addressing", o.fabricAddressing.String())
		dataByIDChecks.append(t, "TestCheckResourceAttr", "fabric_addressing", o.fabricAddressing.String())
		dataByNameChecks.append(t, "TestCheckResourceAttr", "fabric_addressing", o.fabricAddressing.String())
	}

	if o.antiAffinityMode != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "anti_affinity_mode", o.antiAffinityMode.String())
		dataByIDChecks.append(t, "TestCheckResourceAttr", "anti_affinity_mode", o.antiAffinityMode.String())
		dataByIDChecks.append(t, "TestCheckResourceAttr", "anti_affinity_mode", o.antiAffinityMode.String())
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "anti_affinity_mode")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "anti_affinity_mode")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "anti_affinity_mode")
	}

	o.antiAffinityPolicy.testChecks(t, &resourceChecks)
	o.antiAffinityPolicy.testChecks(t, &dataByIDChecks)
	o.antiAffinityPolicy.testChecks(t, &dataByNameChecks)

	if o.defaultIPLinksToGenericMTU != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "default_ip_links_to_generic_mtu", strconv.Itoa(*o.defaultIPLinksToGenericMTU))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "default_ip_links_to_generic_mtu", strconv.Itoa(*o.defaultIPLinksToGenericMTU))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "default_ip_links_to_generic_mtu", strconv.Itoa(*o.defaultIPLinksToGenericMTU))
	}

	if o.defaultSVIL3MTU != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "default_svi_l3_mtu", strconv.Itoa(*o.defaultSVIL3MTU))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "default_svi_l3_mtu", strconv.Itoa(*o.defaultSVIL3MTU))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "default_svi_l3_mtu", strconv.Itoa(*o.defaultSVIL3MTU))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "default_svi_l3_mtu")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "default_svi_l3_mtu")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "default_svi_l3_mtu")
	}

	if o.esiMACMSB != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "esi_mac_msb", strconv.Itoa(*o.esiMACMSB))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "esi_mac_msb", strconv.Itoa(*o.esiMACMSB))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "esi_mac_msb", strconv.Itoa(*o.esiMACMSB))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "esi_mac_msb")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "esi_mac_msb")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "esi_mac_msb")
	}

	if o.evpnType5Routes != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "evpn_type_5_routes", strconv.FormatBool(*o.evpnType5Routes))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "evpn_type_5_routes", strconv.FormatBool(*o.evpnType5Routes))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "evpn_type_5_routes", strconv.FormatBool(*o.evpnType5Routes))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "evpn_type_5_routes")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "evpn_type_5_routes")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "evpn_type_5_routes")
	}

	if o.fabricMTU != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "fabric_mtu", strconv.Itoa(*o.fabricMTU))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "fabric_mtu", strconv.Itoa(*o.fabricMTU))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "fabric_mtu", strconv.Itoa(*o.fabricMTU))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "fabric_mtu")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "fabric_mtu")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "fabric_mtu")
	}

	if o.ipv6Applications != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "ipv6_applications", strconv.FormatBool(*o.ipv6Applications))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "ipv6_applications", strconv.FormatBool(*o.ipv6Applications))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "ipv6_applications", strconv.FormatBool(*o.ipv6Applications))
	}

	if o.junosEVPNMaxNexthopAndInterfaceNumber != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "junos_evpn_max_nexthop_and_interface_number", strconv.FormatBool(*o.junosEVPNMaxNexthopAndInterfaceNumber))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "junos_evpn_max_nexthop_and_interface_number", strconv.FormatBool(*o.junosEVPNMaxNexthopAndInterfaceNumber))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "junos_evpn_max_nexthop_and_interface_number", strconv.FormatBool(*o.junosEVPNMaxNexthopAndInterfaceNumber))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "junos_evpn_max_nexthop_and_interface_number")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "junos_evpn_max_nexthop_and_interface_number")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "junos_evpn_max_nexthop_and_interface_number")
	}

	if o.junosEVPNRoutingInstanceModeMACVRF != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "junos_evpn_routing_instance_mode_mac_vrf", strconv.FormatBool(*o.junosEVPNRoutingInstanceModeMACVRF))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "junos_evpn_routing_instance_mode_mac_vrf", strconv.FormatBool(*o.junosEVPNRoutingInstanceModeMACVRF))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "junos_evpn_routing_instance_mode_mac_vrf", strconv.FormatBool(*o.junosEVPNRoutingInstanceModeMACVRF))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "junos_evpn_routing_instance_mode_mac_vrf")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "junos_evpn_routing_instance_mode_mac_vrf")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "junos_evpn_routing_instance_mode_mac_vrf")
	}

	if o.junosEXOverlayECMP != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "junos_ex_overlay_ecmp", strconv.FormatBool(*o.junosEXOverlayECMP))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "junos_ex_overlay_ecmp", strconv.FormatBool(*o.junosEXOverlayECMP))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "junos_ex_overlay_ecmp", strconv.FormatBool(*o.junosEXOverlayECMP))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "junos_ex_overlay_ecmp")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "junos_ex_overlay_ecmp")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "junos_ex_overlay_ecmp")
	}

	if o.junosGracefulRestart != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "junos_graceful_restart", strconv.FormatBool(*o.junosGracefulRestart))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "junos_graceful_restart", strconv.FormatBool(*o.junosGracefulRestart))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "junos_graceful_restart", strconv.FormatBool(*o.junosGracefulRestart))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "junos_graceful_restart")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "junos_graceful_restart")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "junos_graceful_restart")
	}

	if o.maxEVPNRoutesCount != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "max_evpn_routes_count", strconv.Itoa(*o.maxEVPNRoutesCount))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "max_evpn_routes_count", strconv.Itoa(*o.maxEVPNRoutesCount))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "max_evpn_routes_count", strconv.Itoa(*o.maxEVPNRoutesCount))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "max_evpn_routes_count")
		// dataByIDChecks.append(t, "TestCheckResourceAttrSet", "max_evpn_routes_count")
		// dataByNameChecks.append(t, "TestCheckResourceAttrSet", "max_evpn_routes_count")
	}

	if o.maxExternalRoutesCount != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "max_external_routes_count", strconv.Itoa(*o.maxExternalRoutesCount))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "max_external_routes_count", strconv.Itoa(*o.maxExternalRoutesCount))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "max_external_routes_count", strconv.Itoa(*o.maxExternalRoutesCount))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "max_external_routes_count")
		// dataByIDChecks.append(t, "TestCheckResourceAttrSet", "max_external_routes_count")
		// dataByNameChecks.append(t, "TestCheckResourceAttrSet", "max_external_routes_count")
	}

	if o.maxFabricRoutesCount != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "max_fabric_routes_count", strconv.Itoa(*o.maxFabricRoutesCount))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "max_fabric_routes_count", strconv.Itoa(*o.maxFabricRoutesCount))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "max_fabric_routes_count", strconv.Itoa(*o.maxFabricRoutesCount))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "max_fabric_routes_count")
		// dataByIDChecks.append(t, "TestCheckResourceAttrSet", "max_fabric_routes_count")
		// dataByNameChecks.append(t, "TestCheckResourceAttrSet", "max_fabric_routes_count")
	}

	if o.maxMLAGRoutesCount != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "max_mlag_routes_count", strconv.Itoa(*o.maxMLAGRoutesCount))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "max_mlag_routes_count", strconv.Itoa(*o.maxMLAGRoutesCount))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "max_mlag_routes_count", strconv.Itoa(*o.maxMLAGRoutesCount))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "max_mlag_routes_count")
		// dataByIDChecks.append(t, "TestCheckResourceAttrSet", "max_mlag_routes_count")
		// dataByNameChecks.append(t, "TestCheckResourceAttrSet", "max_mlag_routes_count")
	}

	if o.optimizeRoutingZoneFootprint != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "optimize_routing_zone_footprint", strconv.FormatBool(*o.optimizeRoutingZoneFootprint))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "optimize_routing_zone_footprint", strconv.FormatBool(*o.optimizeRoutingZoneFootprint))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "optimize_routing_zone_footprint", strconv.FormatBool(*o.optimizeRoutingZoneFootprint))
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "optimize_routing_zone_footprint")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "optimize_routing_zone_footprint")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "optimize_routing_zone_footprint")
	}

	if o.underlayAddressing != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "underlay_addressing", o.underlayAddressing.String())
		dataByIDChecks.append(t, "TestCheckResourceAttr", "underlay_addressing", o.underlayAddressing.String())
		dataByNameChecks.append(t, "TestCheckResourceAttr", "underlay_addressing", o.underlayAddressing.String())
	}

	if o.vtepAddressing != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "vtep_addressing", o.vtepAddressing.String())
		dataByIDChecks.append(t, "TestCheckResourceAttr", "vtep_addressing", o.vtepAddressing.String())
		dataByNameChecks.append(t, "TestCheckResourceAttr", "vtep_addressing", o.vtepAddressing.String())
	}

	if o.disableIPv4 != nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "disable_ipv4", strconv.FormatBool(*o.disableIPv4))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "disable_ipv4", strconv.FormatBool(*o.disableIPv4))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "disable_ipv4", strconv.FormatBool(*o.disableIPv4))
	}

	// computed-only attributes
	resourceChecks.append(t, "TestCheckResourceAttrSet", "status")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "status")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "status")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "superspine_switch_count")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "superspine_switch_count")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "superspine_switch_count")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "spine_switch_count")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "spine_switch_count")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "spine_switch_count")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "leaf_switch_count")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "leaf_switch_count")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "leaf_switch_count")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "access_switch_count")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "access_switch_count")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "access_switch_count")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "generic_system_count")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "generic_system_count")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "generic_system_count")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "external_router_count")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "external_router_count")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "external_router_count")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "has_uncommitted_changes")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "has_uncommitted_changes")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "has_uncommitted_changes")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "version")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "version")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "version")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "build_errors_count")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "build_errors_count")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "build_errors_count")
	resourceChecks.append(t, "TestCheckResourceAttrSet", "build_warnings_count")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "build_warnings_count")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "build_warnings_count")

	return []testChecks{resourceChecks, dataByIDChecks, dataByNameChecks}
}

const resourceDatacenterBlueprintAntiAffinityPolicyHCL = `{
    max_links_count_per_slot            = %s
	max_links_count_per_system_per_slot = %s
	max_links_count_per_port            = %s
	max_links_count_per_system_per_port = %s
  }`

type resourceDatacenterBlueprintAntiAffinityPolicy struct {
	maxLinksPerSlot          *int
	maxLinksPerSystemPerSlot *int
	maxLinksPerPort          *int
	maxLinksPerSystemPerPort *int
}

func (o *resourceDatacenterBlueprintAntiAffinityPolicy) render() string {
	if o == nil {
		return "null"
	}

	return fmt.Sprintf(resourceDatacenterBlueprintAntiAffinityPolicyHCL,
		intPtrOrNull(o.maxLinksPerSlot),
		intPtrOrNull(o.maxLinksPerSystemPerSlot),
		intPtrOrNull(o.maxLinksPerPort),
		intPtrOrNull(o.maxLinksPerSystemPerPort),
	)
}

func (o *resourceDatacenterBlueprintAntiAffinityPolicy) testChecks(t testing.TB, testChecks *testChecks) {
	if o == nil {
		testChecks.append(t, "TestCheckResourceAttrSet", "anti_affinity_policy.max_links_count_per_slot")
		testChecks.append(t, "TestCheckResourceAttrSet", "anti_affinity_policy.max_links_count_per_system_per_slot")
		testChecks.append(t, "TestCheckResourceAttrSet", "anti_affinity_policy.max_links_count_per_port")
		testChecks.append(t, "TestCheckResourceAttrSet", "anti_affinity_policy.max_links_count_per_system_per_port")
		return
	}

	if o.maxLinksPerSlot != nil {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_slot", strconv.Itoa(*o.maxLinksPerSlot))
	} else {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_slot", "0")
	}

	if o.maxLinksPerSystemPerSlot != nil {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_system_per_slot", strconv.Itoa(*o.maxLinksPerSystemPerSlot))
	} else {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_system_per_slot", "0")
	}

	if o.maxLinksPerPort != nil {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_port", strconv.Itoa(*o.maxLinksPerPort))
	} else {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_port", "0")
	}

	if o.maxLinksPerSystemPerPort != nil {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_system_per_port", strconv.Itoa(*o.maxLinksPerSystemPerPort))
	} else {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_system_per_port", "0")
	}
}

func TestResourceDatacenterBlueprint(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)
	clientVersion := version.Must(version.NewVersion(client.ApiVersion()))

	type testStep struct {
		config                     resourceDatacenterBlueprint
		preApplyResourceActionType plancheck.ResourceActionType
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"disable_ipv6_requires_replace_version_before_610": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:             acctest.RandString(6),
						templateID:       "L2_Virtual_EVPN",
						ipv6Applications: pointer.To(true),
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionDestroyBeforeCreate,
					config: resourceDatacenterBlueprint{
						name:             acctest.RandString(6),
						templateID:       "L2_Virtual_EVPN",
						ipv6Applications: pointer.To(false),
					},
				},
			},
		},
		"simple_test_all_versions": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
			},
		},
		"start_minimal_all_versions": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:             acctest.RandString(6),
						templateID:       "L2_Virtual_EVPN",
						antiAffinityMode: oneOf(&enum.AntiAffinityModeDisabled, &enum.AntiAffinityModeStrict, &enum.AntiAffinityModeLoose),
						antiAffinityPolicy: &resourceDatacenterBlueprintAntiAffinityPolicy{
							maxLinksPerSlot:          pointer.To(128 + rand.Intn(128)), // 128 - 255
							maxLinksPerPort:          pointer.To(64 + rand.Intn(64)),   //  64 - 127
							maxLinksPerSystemPerSlot: pointer.To(32 + rand.Intn(32)),   //  32  - 63
							maxLinksPerSystemPerPort: pointer.To(16 + rand.Intn(16)),   //  16  - 31
						},
						defaultIPLinksToGenericMTU:            pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100
						defaultSVIL3MTU:                       pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100,
						esiMACMSB:                             pointer.To(2 + (rand.Intn(126))*2),   // even number 2-254
						evpnType5Routes:                       oneOf(pointer.To(true), pointer.To(false)),
						fabricMTU:                             pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100,
						junosEVPNMaxNexthopAndInterfaceNumber: oneOf(pointer.To(true), pointer.To(false)),
						junosEVPNRoutingInstanceModeMACVRF:    oneOf(pointer.To(true), pointer.To(false)),
						junosEXOverlayECMP:                    oneOf(pointer.To(true), pointer.To(false)),
						junosGracefulRestart:                  oneOf(pointer.To(true), pointer.To(false)),
						maxEVPNRoutesCount:                    oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxExternalRoutesCount:                oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxFabricRoutesCount:                  oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxMLAGRoutesCount:                    oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						optimizeRoutingZoneFootprint:          oneOf(pointer.To(true), pointer.To(false)),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
			},
		},
		"start_maximal_all_versions": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:             acctest.RandString(6),
						templateID:       "L2_Virtual_EVPN",
						antiAffinityMode: oneOf(&enum.AntiAffinityModeDisabled, &enum.AntiAffinityModeStrict, &enum.AntiAffinityModeLoose),
						antiAffinityPolicy: &resourceDatacenterBlueprintAntiAffinityPolicy{
							maxLinksPerSlot:          pointer.To(128 + rand.Intn(128)), // 128 - 255
							maxLinksPerPort:          pointer.To(64 + rand.Intn(64)),   //  64 - 127
							maxLinksPerSystemPerSlot: pointer.To(32 + rand.Intn(32)),   //  32  - 63
							maxLinksPerSystemPerPort: pointer.To(16 + rand.Intn(16)),   //  16  - 31
						},
						defaultIPLinksToGenericMTU:            pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100
						defaultSVIL3MTU:                       pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100,
						esiMACMSB:                             pointer.To(2 + (rand.Intn(126))*2),   // even number 2-254
						evpnType5Routes:                       oneOf(pointer.To(true), pointer.To(false)),
						fabricMTU:                             pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100,
						junosEVPNMaxNexthopAndInterfaceNumber: oneOf(pointer.To(true), pointer.To(false)),
						junosEVPNRoutingInstanceModeMACVRF:    oneOf(pointer.To(true), pointer.To(false)),
						junosEXOverlayECMP:                    oneOf(pointer.To(true), pointer.To(false)),
						junosGracefulRestart:                  oneOf(pointer.To(true), pointer.To(false)),
						maxEVPNRoutesCount:                    oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxExternalRoutesCount:                oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxFabricRoutesCount:                  oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxMLAGRoutesCount:                    oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						optimizeRoutingZoneFootprint:          oneOf(pointer.To(true), pointer.To(false)),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:             acctest.RandString(6),
						templateID:       "L2_Virtual_EVPN",
						antiAffinityMode: oneOf(&enum.AntiAffinityModeDisabled, &enum.AntiAffinityModeStrict, &enum.AntiAffinityModeLoose),
						antiAffinityPolicy: &resourceDatacenterBlueprintAntiAffinityPolicy{
							maxLinksPerSlot:          pointer.To(128 + rand.Intn(128)), // 128 - 255
							maxLinksPerPort:          pointer.To(64 + rand.Intn(64)),   //  64 - 127
							maxLinksPerSystemPerSlot: pointer.To(32 + rand.Intn(32)),   //  32  - 63
							maxLinksPerSystemPerPort: pointer.To(16 + rand.Intn(16)),   //  16  - 31
						},
						defaultIPLinksToGenericMTU:            pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100
						defaultSVIL3MTU:                       pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100,
						esiMACMSB:                             pointer.To(2 + (rand.Intn(126))*2),   // even number 2-254
						evpnType5Routes:                       oneOf(pointer.To(true), pointer.To(false)),
						fabricMTU:                             pointer.To(9000 + (rand.Intn(51))*2), // even number 9000-9100,
						junosEVPNMaxNexthopAndInterfaceNumber: oneOf(pointer.To(true), pointer.To(false)),
						junosEVPNRoutingInstanceModeMACVRF:    oneOf(pointer.To(true), pointer.To(false)),
						junosEXOverlayECMP:                    oneOf(pointer.To(true), pointer.To(false)),
						junosGracefulRestart:                  oneOf(pointer.To(true), pointer.To(false)),
						maxEVPNRoutesCount:                    oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxExternalRoutesCount:                oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxFabricRoutesCount:                  oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						maxMLAGRoutesCount:                    oneOf(pointer.To(0), pointer.To(1+rand.Intn(math.MaxUint8))),
						optimizeRoutingZoneFootprint:          oneOf(pointer.To(true), pointer.To(false)),
					},
				},
			},
		},
		"default_vrf_610_and_later": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_EVPN",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv6),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
						disableIPv4:        pointer.To(true),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_EVPN",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv4),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv4),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_EVPN",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv46),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_EVPN",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv6),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
						disableIPv4:        pointer.To(true),
					},
				},
			},
		},
		"introduce_default_vrf_params_610_and_later": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_EVPN",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv6),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
						disableIPv4:        pointer.To(true),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
			},
		},
		"withdraw_default_vrf_params_610_and_later": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_EVPN",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv6),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
						disableIPv4:        pointer.To(true),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_EVPN",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv46),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
					},
				},
			},
		},
		"template_forces_blueprint_replacement": {
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionDestroyBeforeCreate,
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_Dual_2x_Links",
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionDestroyBeforeCreate,
					config: resourceDatacenterBlueprint{
						name:       acctest.RandString(6),
						templateID: "L2_Virtual_EVPN",
					},
				},
			},
		},
		"ip_fabric_610_and_later": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
			steps: []testStep{
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_Dual_2x_Links",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv6),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
						disableIPv4:        pointer.To(true),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_Dual_2x_Links",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv46),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv4),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_Dual_2x_Links",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv6),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
						disableIPv4:        pointer.To(true),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_Dual_2x_Links",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv4),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv4),
					},
				},
				{
					config: resourceDatacenterBlueprint{
						name:               acctest.RandString(6),
						templateID:         "L2_Virtual_Dual_2x_Links",
						underlayAddressing: pointer.To(enum.AddressingSchemeIPv46),
						vtepAddressing:     pointer.To(enum.AddressingSchemeIPv6),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterBlueprint)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() don't run in parallel -- too many blueprints!'

			if !tCase.versionConstraints.Check(clientVersion) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				var checkLog string
				var checkFuncs []resource.TestCheckFunc
				for _, checkList := range checks {
					checkLog = checkLog + checkList.string(len(checkFuncs))
					checkFuncs = append(checkFuncs, checkList.checks...)
				}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, checkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checkFuncs...),
				}

				// add expected per-step resource action, if any
				if step.preApplyResourceActionType != "" {
					steps[i].ConfigPlanChecks.PreApply = append(steps[i].ConfigPlanChecks.PreApply, plancheck.ExpectResourceAction(resourceType+"."+tName, step.preApplyResourceActionType))
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
