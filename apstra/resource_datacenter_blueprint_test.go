//go:build integration

package tfapstra_test

import (
	"context"
	"errors"
	"fmt"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	return fmt.Sprintf(resourceDatacenterBlueprintHCL,
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
}

func (o resourceDatacenterBlueprint) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "template_id", o.templateID)

	if o.fabricAddressing != nil {
		result.append(t, "TestCheckResourceAttr", "fabric_addressing", o.fabricAddressing.String())
	}

	if o.antiAffinityMode != nil {
		result.append(t, "TestCheckResourceAttr", "anti_affinity_mode", o.antiAffinityMode.String())
	} else {
		result.append(t, "TestCheckResourceAttr", "anti_affinity_mode", enum.AntiAffinityModeDisabled.String())
	}

	o.antiAffinityPolicy.testChecks(t, &result)

	if o.defaultIPLinksToGenericMTU != nil {
		result.append(t, "TestCheckResourceAttr", "default_ip_links_to_generic_mtu", strconv.Itoa(*o.defaultIPLinksToGenericMTU))
	}

	if o.defaultSVIL3MTU != nil {
		result.append(t, "TestCheckResourceAttr", "default_svi_l3_mtu", strconv.Itoa(*o.defaultSVIL3MTU))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "default_svi_l3_mtu")
	}

	if o.esiMACMSB != nil {
		result.append(t, "TestCheckResourceAttr", "esi_mac_msb", strconv.Itoa(*o.esiMACMSB))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "esi_mac_msb")
	}

	if o.evpnType5Routes != nil {
		result.append(t, "TestCheckResourceAttr", "evpn_type_5_routes", strconv.FormatBool(*o.evpnType5Routes))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "evpn_type_5_routes")
	}

	if o.fabricMTU != nil {
		result.append(t, "TestCheckResourceAttr", "fabric_mtu", strconv.Itoa(*o.fabricMTU))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "fabric_mtu")
	}

	if o.ipv6Applications != nil {
		result.append(t, "TestCheckResourceAttr", "ipv6_applications", strconv.FormatBool(*o.ipv6Applications))
	}

	if o.junosEVPNMaxNexthopAndInterfaceNumber != nil {
		result.append(t, "TestCheckResourceAttr", "junos_evpn_max_nexthop_and_interface_number", strconv.FormatBool(*o.junosEVPNMaxNexthopAndInterfaceNumber))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "junos_evpn_max_nexthop_and_interface_number")
	}

	if o.junosEVPNRoutingInstanceModeMACVRF != nil {
		result.append(t, "TestCheckResourceAttr", "junos_evpn_routing_instance_mode_mac_vrf", strconv.FormatBool(*o.junosEVPNRoutingInstanceModeMACVRF))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "junos_evpn_routing_instance_mode_mac_vrf")
	}

	if o.junosEXOverlayECMP != nil {
		result.append(t, "TestCheckResourceAttr", "junos_ex_overlay_ecmp", strconv.FormatBool(*o.junosEXOverlayECMP))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "junos_ex_overlay_ecmp")
	}

	if o.junosGracefulRestart != nil {
		result.append(t, "TestCheckResourceAttr", "junos_graceful_restart", strconv.FormatBool(*o.junosGracefulRestart))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "junos_graceful_restart")
	}

	if o.maxEVPNRoutesCount != nil {
		result.append(t, "TestCheckResourceAttr", "max_evpn_routes_count", strconv.Itoa(*o.maxEVPNRoutesCount))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "max_evpn_routes_count")
	}

	if o.maxExternalRoutesCount != nil {
		result.append(t, "TestCheckResourceAttr", "max_external_routes_count", strconv.Itoa(*o.maxExternalRoutesCount))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "max_external_routes_count")
	}

	if o.maxFabricRoutesCount != nil {
		result.append(t, "TestCheckResourceAttr", "max_fabric_routes_count", strconv.Itoa(*o.maxFabricRoutesCount))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "max_fabric_routes_count")
	}

	if o.maxMLAGRoutesCount != nil {
		result.append(t, "TestCheckResourceAttr", "max_mlag_routes_count", strconv.Itoa(*o.maxMLAGRoutesCount))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "max_mlag_routes_count")
	}

	if o.optimizeRoutingZoneFootprint != nil {
		result.append(t, "TestCheckResourceAttr", "optimize_routing_zone_footprint", strconv.FormatBool(*o.optimizeRoutingZoneFootprint))
	} else {
		result.append(t, "TestCheckResourceAttrSet", "optimize_routing_zone_footprint")
	}

	if o.underlayAddressing != nil {
		result.append(t, "TestCheckResourceAttr", "underlay_addressing", o.underlayAddressing.String())
	}

	if o.vtepAddressing != nil {
		result.append(t, "TestCheckResourceAttr", "vtep_addressing", o.vtepAddressing.String())
	}

	if o.disableIPv4 != nil {
		result.append(t, "TestCheckResourceAttr", "disable_ipv4", strconv.FormatBool(*o.disableIPv4))
	}

	return result
}

func TestResourceDatacenterBlueprint(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)
	clientVersion := version.Must(version.NewVersion(client.ApiVersion()))

	type testCase struct {
		steps              []resourceDatacenterBlueprint
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"start_minimal_before_apstra_610": {
			versionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
			steps: []resourceDatacenterBlueprint{
				{
					name:       acctest.RandString(6),
					templateID: "L2_Virtual_EVPN",
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
				config := step.render(resourceType, tName)
				checks := step.testChecks(t, resourceType, tName)

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

const resourceDatacenterBlueprintAntiAffinityPolicyHCL = `{
    max_links_count_per_slot            = %s
	max_links_count_per_system_per_slot = %s
	max_links_count_per_port            = %s
	max_links_count_per_system_per_port = %s
  }`

type resourceDatacenterBlueprintAntiAffinityPolicy struct {
	maxLinksCountPerSlot          *int
	maxLinksCountPerSystemPerSlot *int
	maxLinksCountPerPort          *int
	maxLinksCountPerSystemPerPort *int
}

func (o *resourceDatacenterBlueprintAntiAffinityPolicy) render() string {
	if o == nil {
		return "null"
	}

	return fmt.Sprintf(resourceDatacenterBlueprintAntiAffinityPolicyHCL,
		intPtrOrNull(o.maxLinksCountPerSlot),
		intPtrOrNull(o.maxLinksCountPerSystemPerSlot),
		intPtrOrNull(o.maxLinksCountPerPort),
		intPtrOrNull(o.maxLinksCountPerSystemPerPort),
	)
}

func (o *resourceDatacenterBlueprintAntiAffinityPolicy) testChecks(t testing.TB, testChecks *testChecks) {
	if o == nil {
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_slot", "0")
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_system_per_slot", "0")
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_port", "0")
		testChecks.append(t, "TestCheckResourceAttr", "anti_affinity_policy.max_links_count_per_system_per_port", "0")
		return
	}

	if o.maxLinksCountPerSlot != nil {
		testChecks.append(t, "TestCheckResourceAttr", "max_links_count_per_slot", strconv.Itoa(*o.maxLinksCountPerSlot))
	} else {
		testChecks.append(t, "TestCheckResourceAttr", "max_links_count_per_slot", "0")
	}

	if o.maxLinksCountPerSystemPerSlot != nil {
		testChecks.append(t, "TestCheckResourceAttr", "max_links_count_per_system_per_slot", strconv.Itoa(*o.maxLinksCountPerSystemPerSlot))
	} else {
		testChecks.append(t, "TestCheckResourceAttr", "max_links_count_per_system_per_slot", "0")
	}

	if o.maxLinksCountPerPort != nil {
		testChecks.append(t, "TestCheckResourceAttr", "max_links_count_per_port", strconv.Itoa(*o.maxLinksCountPerPort))
	} else {
		testChecks.append(t, "TestCheckResourceAttr", "max_links_count_per_port", "0")
	}

	if o.maxLinksCountPerSystemPerPort != nil {
		testChecks.append(t, "TestCheckResourceAttr", "max_links_count_per_system_per_port", strconv.Itoa(*o.maxLinksCountPerSystemPerPort))
	} else {
		testChecks.append(t, "TestCheckResourceAttr", "max_links_count_per_system_per_port", "0")
	}
}

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

//func TestResourceDatacenterBlueprint(t *testing.T) {
//	ctx := context.Background()
//	client := testutils.GetTestClient(t, ctx)
//	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))
//	rs := acctest.RandString(6)
//
//	type config struct {
//		name                                  string
//		templateId                            string
//		fabricAddressing                      *enum.AddressingScheme
//		antiAffinityPolicy                    *apstra.AntiAffinityPolicy
//		defaultIpLinksToGenericMtu            *int
//		defaultSviL3Mtu                       *int
//		esiMacMsb                             *int
//		evpnType5Routes                       *bool
//		fabricMtu                             *int
//		ipv6Applications                      *bool
//		junosEvpnMaxNexthopAndInterfaceNumber *bool
//		junosEvpnRoutingInstanceModeMacVrf    *bool
//		junosExOverlayEcmp                    *bool
//		junosGracefulRestart                  *bool
//		maxEvpnRoutesCount                    *int
//		maxExternalRoutesCount                *int
//		maxFabricRoutesCount                  *int
//		maxMlagRoutesCount                    *int
//		optimizeRoutingZoneFootprint          *bool
//	}
//
//	renderAntiAffinityPolicy := func(cfg *apstra.AntiAffinityPolicy) string {
//		if cfg == nil {
//			return "null"
//		}
//
//		return fmt.Sprintf(resourceDatacenterBlueprintAntiAffinityPolicyHCL,
//			fmt.Sprintf("%d", cfg.MaxLinksPerSlot),
//			fmt.Sprintf("%d", cfg.MaxPerSystemLinksPerSlot),
//			fmt.Sprintf("%d", cfg.MaxLinksPerPort),
//			fmt.Sprintf("%d", cfg.MaxPerSystemLinksPerPort),
//		)
//	}
//
//	renderConfig := func(cfg config) string {
//		fabricAddressing := "null"
//		if cfg.fabricAddressing != nil {
//			fabricAddressing = cfg.fabricAddressing.String()
//		}
//
//		antiAffinityMode, antiAffinitiyPolicy := "null", "null"
//		if cfg.antiAffinityPolicy != nil {
//			antiAffinityMode = stringOrNull(cfg.antiAffinityPolicy.Mode.String())
//			antiAffinitiyPolicy = renderAntiAffinityPolicy(cfg.antiAffinityPolicy)
//		}
//
//		return insecureProviderConfigHCL + fmt.Sprintf(resourceDatacenterBlueprintHCL,
//			cfg.name,
//			cfg.templateId,
//			fabricAddressing,
//			antiAffinityMode,
//			antiAffinitiyPolicy,
//			intPtrOrNull(cfg.defaultIpLinksToGenericMtu),
//			intPtrOrNull(cfg.defaultSviL3Mtu),
//			intPtrOrNull(cfg.esiMacMsb),
//			boolPtrOrNull(cfg.evpnType5Routes),
//			intPtrOrNull(cfg.fabricMtu),
//			boolPtrOrNull(cfg.ipv6Applications),
//			boolPtrOrNull(cfg.junosEvpnMaxNexthopAndInterfaceNumber),
//			boolPtrOrNull(cfg.junosEvpnRoutingInstanceModeMacVrf),
//			boolPtrOrNull(cfg.junosExOverlayEcmp),
//			boolPtrOrNull(cfg.junosGracefulRestart),
//			intPtrOrNull(cfg.maxEvpnRoutesCount),
//			intPtrOrNull(cfg.maxExternalRoutesCount),
//			intPtrOrNull(cfg.maxFabricRoutesCount),
//			intPtrOrNull(cfg.maxMlagRoutesCount),
//			boolPtrOrNull(cfg.optimizeRoutingZoneFootprint),
//		)
//	}
//
//	atleast50 := func(s string) error {
//		i, err := strconv.Atoi(s)
//		if err != nil {
//			return err
//		}
//
//		if i < 50 {
//			return errors.New("expected value >= 50, got " + s)
//		}
//
//		return nil
//	}
//
//	type testCase struct {
//		apiVersionConstraints version.Constraints
//		testCase              resource.TestCase
//	}
//
//	testCases := map[string]testCase{
//		"evpn_start_minimal_before_6.1.0": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "esMinAV_0_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMinAV_1_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             nil,
//							ipv6Applications:                      pointer.To(true),
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMinAV_2_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//				},
//			},
//		},
//		"evpn_start_minimal_6.1.0_and_later": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "esMinAV_0_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMinAV_1_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             nil,
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMinAV_2_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMinAV_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//				},
//			},
//		},
//		"evpn_start_maximal_before_6.1.0": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "esMaxAV_0_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             nil,
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMaxAV_1_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMaxAV_2_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          6,
//								MaxLinksPerSlot:          10,
//								MaxPerSystemLinksPerPort: 4,
//								MaxPerSystemLinksPerSlot: 8,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9004),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(6),
//							evpnType5Routes:                       pointer.To(false),
//							fabricMtu:                             nil,
//							ipv6Applications:                      pointer.To(true),
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(20001),
//							maxExternalRoutesCount:                pointer.To(20002),
//							maxFabricRoutesCount:                  pointer.To(20003),
//							maxMlagRoutesCount:                    pointer.To(20004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "10"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "20001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "20002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "20003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "20004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//				},
//			},
//		},
//		"evpn_start_maximal_6.1.0_and_later": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "esMaxAV_0_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             nil,
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMaxAV_1_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMaxAV_2_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          6,
//								MaxLinksPerSlot:          10,
//								MaxPerSystemLinksPerPort: 4,
//								MaxPerSystemLinksPerSlot: 8,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9004),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(6),
//							evpnType5Routes:                       pointer.To(false),
//							fabricMtu:                             nil,
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(20001),
//							maxExternalRoutesCount:                pointer.To(20002),
//							maxFabricRoutesCount:                  pointer.To(20003),
//							maxMlagRoutesCount:                    pointer.To(20004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMaxAV_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "10"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "20001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "20002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "20003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "20004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//				},
//			},
//		},
//		"evpn_start_minimal_4.2.x_through_6.1.0": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "esMin42x_0_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMin42x_1_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							ipv6Applications:                      pointer.To(true),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMin42x_2_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//				},
//			},
//		},
//		"evpn_start_minimal_42x_6.1.0_and_later": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "esMin42x_0_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMin42x_1_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMin42x_2_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMin42x_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//				},
//			},
//		},
//		"evpn_start_maximal_42x_through_6.0.0": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "esMax42x_0_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							ipv6Applications:                      pointer.To(true),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMax42x_1_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMax42x_2_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							ipv6Applications:                      pointer.To(true),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//				},
//			},
//		},
//		"evpn_start_maximal_42x_6.1.0_and_later": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "esMax42x_0_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMax42x_1_" + rs,
//							templateId: "L2_Virtual_EVPN",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "esMax42x_2_" + rs,
//							templateId: "L2_Virtual_EVPN",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esMax42x_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//				},
//			},
//		},
//		"ip_start_minimal_all_versions_through_6.0.0": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "isMinAV_0_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMinAV_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMinAV_1_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             nil,
//							ipv6Applications:                      pointer.To(false),
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMinAV_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMinAV_2_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMinAV_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//				},
//			},
//		},
//		"ip_start_maximal_all_versions_6.1.0_and_later": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "isMaxAV_0_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             nil,
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMaxAV_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMaxAV_1_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMaxAV_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMaxAV_2_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          6,
//								MaxLinksPerSlot:          10,
//								MaxPerSystemLinksPerPort: 4,
//								MaxPerSystemLinksPerSlot: 8,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9004),
//							defaultSviL3Mtu:                       nil,
//							esiMacMsb:                             pointer.To(6),
//							evpnType5Routes:                       pointer.To(false),
//							fabricMtu:                             nil,
//							junosEvpnMaxNexthopAndInterfaceNumber: nil,
//							junosEvpnRoutingInstanceModeMacVrf:    nil,
//							junosExOverlayEcmp:                    nil,
//							junosGracefulRestart:                  nil,
//							maxEvpnRoutesCount:                    pointer.To(20001),
//							maxExternalRoutesCount:                pointer.To(20002),
//							maxFabricRoutesCount:                  pointer.To(20003),
//							maxMlagRoutesCount:                    pointer.To(20004),
//							optimizeRoutingZoneFootprint:          nil,
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMaxAV_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "10"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "20001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "20002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "20003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "20004"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//				},
//			},
//		},
//		"ip_start_minimal_42x_through_6.0.0": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "isMin42x_0_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMin42x_1_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							ipv6Applications:                      pointer.To(false),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMin42x_2_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//				},
//			},
//		},
//		"ip_start_minimal_6.1.0_and_later": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "isMin42x_0_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeDisabled.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "0"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "-1"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMin42x_1_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMin42x_2_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMin42x_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//				},
//			},
//		},
//		"ip_start_maximal_42x_through_6.0.0": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "isMax42x_0_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							ipv6Applications:                      pointer.To(false),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMax42x_1_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMax42x_2_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							ipv6Applications:                      pointer.To(false),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//				},
//			},
//		},
//		"ip_start_maximal_42x_6.1.0_and_later": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.GeApstra610)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:       "isMax42x_0_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMax42x_1_" + rs,
//							templateId: "L2_Virtual",
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_1_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),     // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"), // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),   // todo: this value is cleared when null, depends on SDK #230
//							// resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),     // todo: this value is cleared when null, depends on SDK #230
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//					{
//						Config: renderConfig(config{
//							name:       "isMax42x_2_" + rs,
//							templateId: "L2_Virtual",
//							antiAffinityPolicy: &apstra.AntiAffinityPolicy{
//								MaxLinksPerPort:          4,
//								MaxLinksPerSlot:          8,
//								MaxPerSystemLinksPerPort: 2,
//								MaxPerSystemLinksPerSlot: 6,
//								Mode:                     apstra.AntiAffinityModeEnabledStrict,
//							},
//							defaultIpLinksToGenericMtu:            pointer.To(9002),
//							defaultSviL3Mtu:                       pointer.To(9004),
//							esiMacMsb:                             pointer.To(4),
//							evpnType5Routes:                       pointer.To(true),
//							fabricMtu:                             pointer.To(9006),
//							junosEvpnMaxNexthopAndInterfaceNumber: pointer.To(false),
//							junosEvpnRoutingInstanceModeMacVrf:    pointer.To(false),
//							junosExOverlayEcmp:                    pointer.To(false),
//							junosGracefulRestart:                  pointer.To(false),
//							maxEvpnRoutesCount:                    pointer.To(10001),
//							maxExternalRoutesCount:                pointer.To(10002),
//							maxFabricRoutesCount:                  pointer.To(10003),
//							maxMlagRoutesCount:                    pointer.To(10004),
//							optimizeRoutingZoneFootprint:          pointer.To(false),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "isMax42x_2_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "status", "created"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "access_switch_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "generic_system_count", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "external_router_count", "0"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
//							resource.TestCheckResourceAttrWith("apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "build_warnings_count", "0"),
//
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "6"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "fabric_mtu", "9006"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "junos_graceful_restart", "false"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "false"),
//						}...),
//					},
//				},
//			},
//		},
//		"evpn_start_with_ipv6": {
//			apiVersionConstraints: version.MustConstraints(version.NewConstraint(apiversions.LeApstra600)),
//			testCase: resource.TestCase{
//				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//				Steps: []resource.TestStep{
//					{
//						Config: renderConfig(config{
//							name:             "esv6AV_0_" + rs,
//							templateId:       "L2_Virtual_EVPN",
//							ipv6Applications: pointer.To(true),
//						}),
//						Check: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
//							resource.TestCheckResourceAttrSet("apstra_datacenter_blueprint.test", "id"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "name", "esv6AV_0_"+rs),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "template_id", "L2_Virtual_EVPN"),
//							resource.TestCheckResourceAttr("apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
//						}...),
//					},
//				},
//			},
//		},
//	}
//
//	for tName, tCase := range testCases {
//		tName, tCase := tName, tCase
//		t.Run(tName, func(t *testing.T) {
//			// t.Parallel()
//			if !tCase.apiVersionConstraints.Check(apiVersion) {
//				t.Skipf("API version %s does not satisfy version constraints(%s) of test %q",
//					apiVersion, tCase.apiVersionConstraints, tName)
//			}
//			resource.Test(t, tCase.testCase)
//		})
//	}
//}
