//go:build integration

package tfapstra_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const dataSourceDatacenterBlueprintHCL = `
data "apstra_datacenter_blueprint" "test" {
  id   = %s
  name = %s
}
`

func TestDatasourceDatacenterBlueprint(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

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
		label                string
		apiVersionConstrants version.Constraints
		templateId           apstra.ObjectId
		fabricSettings       apstra.FabricSettings
		checks               []resource.TestCheckFunc
		ipv6                 bool
	}

	testCases := map[string]testCase{
		"evpn_all_versions_ipv4": {
			label:      acctest.RandString(5),
			templateId: "L2_Virtual_EVPN",
			fabricSettings: apstra.FabricSettings{
				AntiAffinityPolicy: &apstra.AntiAffinityPolicy{
					Algorithm:                apstra.AlgorithmHeuristic,
					MaxLinksPerPort:          4,
					MaxLinksPerSlot:          8,
					MaxPerSystemLinksPerPort: 2,
					MaxPerSystemLinksPerSlot: 4,
					Mode:                     apstra.AntiAffinityModeEnabledStrict,
				},
				// DefaultSviL3Mtu:                       nil,
				EsiMacMsb:                   utils.ToPtr(uint8(4)),
				EvpnGenerateType5HostRoutes: &enum.FeatureSwitchEnabled,
				ExternalRouterMtu:           utils.ToPtr(uint16(9002)),
				// FabricL3Mtu:                 nil,
				Ipv6Enabled: utils.ToPtr(false),
				// JunosEvpnDuplicateMacRecoveryTime:     nil,
				// JunosEvpnMaxNexthopAndInterfaceNumber: nil,
				// JunosEvpnRoutingInstanceVlanAware:     nil,
				// JunosExOverlayEcmp:                    nil,
				// JunosGracefulRestart:                  nil,
				MaxEvpnRoutes:     utils.ToPtr(uint32(10001)),
				MaxExternalRoutes: utils.ToPtr(uint32(10002)),
				MaxFabricRoutes:   utils.ToPtr(uint32(10003)),
				MaxMlagRoutes:     utils.ToPtr(uint32(10004)),
				// OptimiseSzFootprint:                   nil,
				// OverlayControlProtocol:                nil,
				// SpineLeafLinks:                        nil,
				// SpineSuperspineLinks:                  nil,
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "status", "created"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "access_switch_count", "0"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "generic_system_count", "8"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "external_router_count", "0"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
				resource.TestCheckResourceAttrWith("data.apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
				resource.TestCheckResourceAttrWith("data.apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "4"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "ipv6_applications", "false"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
			},
		},
		"evpn_all_versions_ipv6": {
			ipv6:       true,
			label:      acctest.RandString(5),
			templateId: "L2_Virtual_EVPN",
			fabricSettings: apstra.FabricSettings{
				AntiAffinityPolicy: &apstra.AntiAffinityPolicy{
					Algorithm:                apstra.AlgorithmHeuristic,
					MaxLinksPerPort:          4,
					MaxLinksPerSlot:          8,
					MaxPerSystemLinksPerPort: 2,
					MaxPerSystemLinksPerSlot: 4,
					Mode:                     apstra.AntiAffinityModeEnabledStrict,
				},
				// DefaultSviL3Mtu:                       nil,
				EsiMacMsb:                   utils.ToPtr(uint8(4)),
				EvpnGenerateType5HostRoutes: &enum.FeatureSwitchEnabled,
				ExternalRouterMtu:           utils.ToPtr(uint16(9002)),
				// FabricL3Mtu:                 nil,
				Ipv6Enabled: utils.ToPtr(false),
				// JunosEvpnDuplicateMacRecoveryTime:     nil,
				// JunosEvpnMaxNexthopAndInterfaceNumber: nil,
				// JunosEvpnRoutingInstanceVlanAware:     nil,
				// JunosExOverlayEcmp:                    nil,
				// JunosGracefulRestart:                  nil,
				MaxEvpnRoutes:     utils.ToPtr(uint32(10001)),
				MaxExternalRoutes: utils.ToPtr(uint32(10002)),
				MaxFabricRoutes:   utils.ToPtr(uint32(10003)),
				MaxMlagRoutes:     utils.ToPtr(uint32(10004)),
				// OptimiseSzFootprint:                   nil,
				// OverlayControlProtocol:                nil,
				// SpineLeafLinks:                        nil,
				// SpineSuperspineLinks:                  nil,
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "status", "created"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "superspine_switch_count", "0"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "spine_switch_count", "2"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "leaf_switch_count", "4"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "access_switch_count", "0"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "generic_system_count", "8"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "external_router_count", "0"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "has_uncommitted_changes", "true"),
				resource.TestCheckResourceAttrWith("data.apstra_datacenter_blueprint.test", "version", testCheckIntGE1),
				resource.TestCheckResourceAttrWith("data.apstra_datacenter_blueprint.test", "build_errors_count", atleast50),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "build_warnings_count", "0"),

				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_mode", apstra.AntiAffinityModeEnabledStrict.String()),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_slot", "8"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_slot", "4"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_port", "4"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "anti_affinity_policy.max_links_count_per_system_per_port", "2"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "default_ip_links_to_generic_mtu", "9002"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "default_svi_l3_mtu", "9000"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "esi_mac_msb", "4"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "evpn_type_5_routes", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "fabric_mtu", "9170"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "ipv6_applications", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "junos_evpn_max_nexthop_and_interface_number", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "junos_evpn_routing_instance_mode_mac_vrf", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "junos_ex_overlay_ecmp", "true"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "junos_graceful_restart", "true"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "max_evpn_routes_count", "10001"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "max_external_routes_count", "10002"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "max_fabric_routes_count", "10003"),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "max_mlag_routes_count", "10004"),
				// resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "optimize_routing_zone_footprint", "true"),
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.apiVersionConstrants.Check(apiVersion) {
				t.Skipf("API version %s does not satisfy version constraints(%s) of test %q",
					apiVersion, tCase.apiVersionConstrants, tName)
			}

			// create the blueprint
			id, err := client.CreateBlueprintFromTemplate(ctx, &apstra.CreateBlueprintFromTemplateRequest{
				RefDesign:      enum.RefDesignDatacenter,
				Label:          tCase.label,
				TemplateId:     tCase.templateId,
				FabricSettings: &tCase.fabricSettings,
			})
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, client.DeleteBlueprint(ctx, id)) })

			var bpClient *apstra.TwoStageL3ClosClient

			if tCase.ipv6 {
				if bpClient == nil {
					bpClient, err = client.NewTwoStageL3ClosClient(ctx, id)
					require.NoError(t, err)
				}

				fs, err := bpClient.GetFabricSettings(ctx)
				require.NoError(t, err)

				fs.Ipv6Enabled = utils.ToPtr(true)
				err = bpClient.SetFabricSettings(ctx, fs)
				require.NoError(t, err)
			}

			// set anti-affinity policy as needed with Apstra 4.2.0
			if compatibility.TemplateRequiresAntiAffinityPolicy.Check(apiVersion) && tCase.fabricSettings.AntiAffinityPolicy != nil {
				if bpClient == nil {
					bpClient, err = client.NewTwoStageL3ClosClient(ctx, id)
					require.NoError(t, err)
				}

				// force IPv6 lever as specified by the test case (it may need to be on)
				if tCase.ipv6 {
					tCase.fabricSettings.Ipv6Enabled = utils.ToPtr(true)
				}

				err = bpClient.SetFabricSettings(ctx, &tCase.fabricSettings)
				require.NoError(t, err)
			}

			// add unpredictable name and ID to existing checks
			tCase.checks = append(tCase.checks,
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "id", id.String()),
				resource.TestCheckResourceAttr("data.apstra_datacenter_blueprint.test", "name", tCase.label),
			)

			// test lookup by ID
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{ // look up by ID
						Config: insecureProviderConfigHCL +
							fmt.Sprintf(dataSourceDatacenterBlueprintHCL, `"`+id+`"`, "null"),
						Check: resource.ComposeAggregateTestCheckFunc(tCase.checks...),
					},
					{ // look up by Name
						Config: insecureProviderConfigHCL +
							fmt.Sprintf(dataSourceDatacenterBlueprintHCL, "null", `"`+tCase.label+`"`),
						Check: resource.ComposeAggregateTestCheckFunc(tCase.checks...),
					},
				},
			})
		})
	}
}
