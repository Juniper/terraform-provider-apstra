package blueprint

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	apstraplanmodifier "github.com/Juniper/terraform-provider-apstra/apstra/apstra_plan_modifier"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"golang.org/x/exp/constraints"
)

type Blueprint struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	TemplateId       types.String `tfsdk:"template_id"`
	FabricAddressing types.String `tfsdk:"fabric_addressing"`

	// status
	Status                types.String `tfsdk:"status"`
	SuperspineCount       types.Int64  `tfsdk:"superspine_switch_count"`
	SpineCount            types.Int64  `tfsdk:"spine_switch_count"`
	LeafCount             types.Int64  `tfsdk:"leaf_switch_count"`
	AccessCount           types.Int64  `tfsdk:"access_switch_count"`
	GenericCount          types.Int64  `tfsdk:"generic_system_count"`
	ExternalCount         types.Int64  `tfsdk:"external_router_count"`
	HasUncommittedChanges types.Bool   `tfsdk:"has_uncommitted_changes"`
	Version               types.Int64  `tfsdk:"version"`
	BuildErrorsCount      types.Int64  `tfsdk:"build_errors_count"`
	BuildWarningsCount    types.Int64  `tfsdk:"build_warnings_count"`

	// fabric settings
	AntiAffinityMode                      types.String `tfsdk:"anti_affinity_mode"`
	AntiAffinityPolicy                    types.Object `tfsdk:"anti_affinity_policy"`
	DefaultIpLinksToGenericMtu            types.Int64  `tfsdk:"default_ip_links_to_generic_mtu"`
	DefaultSviL3Mtu                       types.Int64  `tfsdk:"default_svi_l3_mtu"`
	EsiMacMsb                             types.Int64  `tfsdk:"esi_mac_msb"`
	EvpnType5Routes                       types.Bool   `tfsdk:"evpn_type_5_routes"`
	FabricMtu                             types.Int64  `tfsdk:"fabric_mtu"`
	Ipv6Applications                      types.Bool   `tfsdk:"ipv6_applications"`
	JunosEvpnMaxNexthopAndInterfaceNumber types.Bool   `tfsdk:"junos_evpn_max_nexthop_and_interface_number"`
	JunosEvpnRoutingInstanceModeMacVrf    types.Bool   `tfsdk:"junos_evpn_routing_instance_mode_mac_vrf"`
	JunosExOverlayEcmp                    types.Bool   `tfsdk:"junos_ex_overlay_ecmp"`
	JunosGracefulRestart                  types.Bool   `tfsdk:"junos_graceful_restart"`
	MaxEvpnRoutesCount                    types.Int64  `tfsdk:"max_evpn_routes_count"`
	MaxExternalRoutesCount                types.Int64  `tfsdk:"max_external_routes_count"`
	MaxFabricRoutesCount                  types.Int64  `tfsdk:"max_fabric_routes_count"`
	MaxMlagRoutesCount                    types.Int64  `tfsdk:"max_mlag_routes_count"`
	OptimizeRoutingZoneFootprint          types.Bool   `tfsdk:"optimize_routing_zone_footprint"`
}

//func (o Blueprint) attrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"id":                types.StringType,
//		"name":              types.StringType,
//		"template_id":       types.StringType,
//		"fabric_addressing": types.StringType,
//
//		"status":                  types.StringType,
//		"superspine_switch_count": types.Int64Type,
//		"spine_switch_count":      types.Int64Type,
//		"leaf_switch_count":       types.Int64Type,
//		"access_switch_count":     types.Int64Type,
//		"generic_system_count":    types.Int64Type,
//		"external_router_count":   types.Int64Type,
//		"has_uncommitted_changes": types.BoolType,
//		"version":                 types.Int64Type,
//		"build_errors_count":      types.Int64Type,
//		"build_warnings_count":    types.Int64Type,
//
//		"anti_affinity_mode":                          types.StringType,
//		"anti_affinity_policy":                        types.ObjectType{AttrTypes: AntiAffinityPolicy{}.attrTypes()},
//		"default_ip_links_to_generic_mtu":             types.Int64Type,
//		"default_svi_l3_mtu":                          types.Int64Type,
//		"esi_mac_msb":                                 types.Int64Type,
//		"evpn_type_5_routes":                          types.BoolType,
//		"fabric_mtu":                                  types.Int64Type,
//		"ipv6_applications":                           types.BoolType,
//		"junos_evpn_max_nexthop_and_interface_number": types.BoolType,
//		"junos_evpn_routing_instance_mode_mac_vrf":    types.BoolType,
//		"junos_ex_overlay_ecmp":                       types.BoolType,
//		"junos_graceful_restart":                      types.BoolType,
//		"max_evpn_routes_count":                       types.Int64Type,
//		"max_external_routes_count":                   types.Int64Type,
//		"max_fabric_routes_count":                     types.Int64Type,
//		"max_mlag_routes_count":                       types.Int64Type,
//		"optimize_routing_zone_footprint":             types.BoolType,
//	}
//}

func (o Blueprint) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Blueprint. Required when `name` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name of the Blueprint. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"template_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "This attribute is always `null` in data source context. Ignore.",
			Computed:            true,
		},
		"fabric_addressing": dataSourceSchema.StringAttribute{
			MarkdownDescription: "This attribute is always `null` in data source context. Ignore.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status of the Blueprint",
			Computed:            true,
		},
		"superspine_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of superspine switches in the topology.",
			Computed:            true,
		},
		"spine_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of spine switches in the topology.",
			Computed:            true,
		},
		"leaf_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of leaf switches in the topology.",
			Computed:            true,
		},
		"access_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of access switches in the topology.",
			Computed:            true,
		},
		"generic_system_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of generic systems in the topology.",
			Computed:            true,
		},
		"external_router_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of external routers in the topology.",
			Computed:            true,
		},
		"has_uncommitted_changes": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the staging blueprint has uncommitted changes.",
			Computed:            true,
		},
		"version": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Currently active blueprint version",
			Computed:            true,
		},
		"build_errors_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build errors.",
			Computed:            true,
		},
		"build_warnings_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build warnings.",
			Computed:            true,
		},
		"anti_affinity_mode": dataSourceSchema.StringAttribute{
			Computed: true,
			MarkdownDescription: "The anti-affinity policy has three modes:\n" +
				"\t* `Disabled` (default) - ports selection is based on assigned interface maps and interface names " +
				"(provided or auto-assigned). Port breakouts could terminate on the same physical ports.\n" +
				"\t* `loose` - controls interface names that were not defined by the user. Does not control or override " +
				"user-defined cabling. (If you haven't explicitly assigned any interface names, loose and strict are " +
				"effectively the same policy.)\n" +
				"\t* `strict` - completely controls port distribution and could override user-defined assignments. " +
				"When you enable the strict policy, a statement appears at the top of the cabling map " +
				"(Staged/Active > Physical > Links and Staged/Active > Physical > Topology Selection) stating that the " +
				"anti-affinity policy is enabled.",
		},
		"anti_affinity_policy": dataSourceSchema.SingleNestedAttribute{
			Computed: true,
			MarkdownDescription: "When designing high availability (HA) systems, you want parallel links between two " +
				"devices to terminate on different physical ports, thus avoiding transceiver failures from impacting " +
				"both links on a device. Depending on the number of interfaces on a system, manually modifying these " +
				"links could be time-consuming. With the anti-affinity policy you can apply certain constraints to " +
				"the cabling map to control automatic port assignments.",
			Attributes: AntiAffinityPolicy{}.datasourceAttributes(),
		},
		"default_ip_links_to_generic_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Default L3 MTU for IP links to generic systems.",
			Computed:            true,
		},
		"default_svi_l3_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Default L3 MTU for SVI interfaces.",
			Computed:            true,
		},
		"esi_mac_msb": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "ESI MAC address most significant byte.",
			Computed:            true,
		},
		"evpn_type_5_routes": dataSourceSchema.BoolAttribute{
			Computed: true,
			MarkdownDescription: "When enabled, all EVPN VTEPs in the fabric will redistribute " +
				"ARP/IPV6 ND (when possible on NOS type) as EVPN type 5 /32 routes in the routing table.",
		},
		"fabric_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "MTU of fabric links.",
			Computed:            true,
		},
		"ipv6_applications": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables support for IPv6 virtual networks and IPv6 external " +
				"connectivity points. This adds resource requirements and device configurations, " +
				"including IPv6 loopback addresses on leafs, spines and superspines, IPv6 addresses " +
				"for MLAG SVI subnets and IPv6 addresses for leaf L3 peer links.",
			Computed: true,
		},
		"junos_evpn_max_nexthop_and_interface_number": dataSourceSchema.BoolAttribute{
			Computed: true,
			MarkdownDescription: "Enables configuring the maximum number of nexthops and interface numbers reserved " +
				"for use in EVPN-VXLAN overlay network on Junos leaf devices. Default is enabled.",
		},
		"junos_evpn_routing_instance_mode_mac_vrf": dataSourceSchema.BoolAttribute{
			Computed: true,
			MarkdownDescription: "In releases before 4.2, Apstra used a single default switch instance as the " +
				"configuration model for Junos. In Apstra 4.2, Apstra transitioned to using MAC-VRF for all new " +
				"blueprints and normalized the configuration of Junos to Junos Evolved. This option indicates whether " +
				"the blueprint is configured to transition Junos devices to the MAC-VRF configuration model for any " +
				"blueprints deployed before the 4.2 release. All models use the VLAN-Aware service type.",
		},
		"junos_ex_overlay_ecmp": dataSourceSchema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Enables VXLAN Overlay ECMP on Junos EX-series devices",
		},
		"junos_graceful_restart": dataSourceSchema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Enables the Graceful Restart feature on Junos devices",
		},
		"max_evpn_routes_count": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Maximum number of EVPN routes to accept on Leaf Switches. " +
				"A positive integer value indicates the route limit being rendered into to the device BGP " +
				"configuration as a maximum limit. A zero indicates that a `0` is being rendered into the same line of " +
				"configuration, resulting in platform-specific behavior: Eitehr *unlimited routes* are permitted, or " +
				"*no routes* are permitted, depending on the NOS in use. When `null`, Apstra is rendering no maximum " +
				"value into the configuration, so NOS default is being used.",
		},
		"max_external_routes_count": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Maximum number of routes to accept from external routers. " +
				"A positive integer value indicates the route limit being rendered into to the device BGP " +
				"configuration as a maximum limit. A zero indicates that a `0` is being rendered into the same line of " +
				"configuration, resulting in platform-specific behavior: Eitehr *unlimited routes* are permitted, or " +
				"*no routes* are permitted, depending on the NOS in use. When `null`, Apstra is rendering no maximum " +
				"value into the configuration, so NOS default is being used.",
		},
		"max_fabric_routes_count": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Maximum number of underlay routes permitted between fabric nodes. " +
				"A positive integer value indicates the route limit being rendered into to the device BGP " +
				"configuration as a maximum limit. A zero indicates that a `0` is being rendered into the same line of " +
				"configuration, resulting in platform-specific behavior: Eitehr *unlimited routes* are permitted, or " +
				"*no routes* are permitted, depending on the NOS in use. When `null`, Apstra is rendering no maximum " +
				"value into the configuration, so NOS default is being used." +
				"Setting this option may be required in the event of leaking EVPN routes from a Security Zone " +
				"into the default Security Zone (VRF) which may generate a large number of /32 and /128 routes. " +
				"It is suggested that this value be effectively unlimited on all Blueprints to ensure BGP stability in " +
				"the underlay. Unlimited is also suggested for non-EVPN Blueprints considering the impact to traffic if " +
				"spine-leaf sessions go offline.",
		},
		"max_mlag_routes_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of routes to accept between MLAG peers. " +
				"A positive integer value indicates the route limit being rendered into to the device BGP " +
				"configuration as a maximum limit. A zero indicates that a `0` is being rendered into the same line of " +
				"configuration, resulting in platform-specific behavior: Eitehr *unlimited routes* are permitted, or " +
				"*no routes* are permitted, depending on the NOS in use. When `null`, Apstra is rendering no maximum " +
				"value into the configuration, so NOS default is being used.",
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(1, math.MaxUint32),
			},
		},
		"optimize_routing_zone_footprint": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "When `true`: routing zones will not be rendered on leafs where it is not required," +
				"which results in less resource consumption. Routing zone will only be rendered for systems which have " +
				"other structures configured on top of routing zone, such as virtual networks, protocol sessions, " +
				"static routes, sub-interfaces, etc.",
			Computed: true,
		},
	}
}

func (o Blueprint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID assigned by Apstra.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint name.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"template_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of Rack Based Template used to instantiate the Blueprint.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"fabric_addressing": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Addressing scheme for both superspine/spine and spine/leaf links. Only "+
				"applicable to Apstra versions %s and later. Must be one of: %s",
				apiversions.Apstra411, strings.Join([]string{
					apstra.AddressingSchemeIp4.String(),
					apstra.AddressingSchemeIp6.String(),
					apstra.AddressingSchemeIp46.String(),
				}, ", ")),
			Optional: true,
			// todo once depreciated 4.1.0 add this
			// Default: stringdefault.StaticString(apstra.AddressingSchemeIp4.String()),
			Validators: []validator.String{stringvalidator.OneOf(
				apstra.AddressingSchemeIp4.String(),
				apstra.AddressingSchemeIp6.String(),
				apstra.AddressingSchemeIp46.String())},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status of the Blueprint",
			Computed:            true,
		},
		"superspine_switch_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "For 5-stage topologies, the count of superspine switches in the topology.",
			Computed:            true,
		},
		"spine_switch_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "For 3-stage and 5-stage topologies, the count of spine switches in the topology.",
			Computed:            true,
		},
		"leaf_switch_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of leaf switches in the topology.",
			Computed:            true,
		},
		"access_switch_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of access switches in the topology.",
			Computed:            true,
		},
		"generic_system_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of generic systems in the topology.",
			Computed:            true,
		},
		"external_router_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of external routers attached to the topology.",
			Computed:            true,
		},
		"has_uncommitted_changes": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the staging blueprint has uncommitted changes.",
			Computed:            true,
		},
		"version": resourceSchema.Int64Attribute{
			MarkdownDescription: "Currently active blueprint version",
			Computed:            true,
		},
		"build_errors_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build errors.",
			Computed:            true,
		},
		"build_warnings_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build warnings.",
			Computed:            true,
		},
		"anti_affinity_mode": resourceSchema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "The anti-affinity policy has three modes:\n" +
				"\t* `Disabled` (default) - ports selection is based on assigned interface maps and interface names " +
				"(provided or auto-assigned). Port breakouts could terminate on the same physical ports.\n" +
				"\t* `loose` - controls interface names that were not defined by the user. Does not control or override " +
				"user-defined cabling. (If you haven't explicitly assigned any interface names, loose and strict are " +
				"effectively the same policy.)\n" +
				"\t* `strict` - completely controls port distribution and could override user-defined assignments. " +
				"When you enable the strict policy, a statement appears at the top of the cabling map " +
				"(Staged/Active > Physical > Links and Staged/Active > Physical > Topology Selection) stating that the " +
				"anti-affinity policy is enabled.",
			Validators: []validator.String{
				stringvalidator.OneOf(
					apstra.AntiAffinityModeDisabled.String(),
					apstra.AntiAffinityModeEnabledStrict.String(),
					apstra.AntiAffinityModeEnabledLoose.String(),
				),
				stringvalidator.AlsoRequires(path.MatchRoot("anti_affinity_policy")),
			},
		},
		"anti_affinity_policy": resourceSchema.SingleNestedAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "When designing high availability (HA) systems, you want parallel links between two " +
				"devices to terminate on different physical ports, thus avoiding transceiver failures from impacting " +
				"both links on a device. Depending on the number of interfaces on a system, manually modifying these " +
				"links could be time-consuming. With the anti-affinity policy you can apply certain constraints to " +
				"the cabling map to control automatic port assignments.",
			Attributes: AntiAffinityPolicy{}.resourceAttributes(),
			Validators: []validator.Object{objectvalidator.AlsoRequires(path.MatchRoot("anti_affinity_mode"))},
		},
		"default_ip_links_to_generic_mtu": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("Default L3 MTU for IP links to generic systems. A null or empty "+
				"value implies AOS will not render explicit MTU value and system defaults will be used. Should be an "+
				"even number between %d and %d. Requires Apstra %s", constants.L3MtuMin, constants.L3MtuMax, apiversions.Ge420),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"default_svi_l3_mtu": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("Default L3 MTU for SVI interfaces. Should be an even number "+
				"between %d and %d. Requires Apstra %s.", constants.L3MtuMin, constants.L3MtuMax, apiversions.Ge420),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"esi_mac_msb": resourceSchema.Int64Attribute{
			MarkdownDescription: "ESI MAC address most significant byte. Must be an even number " +
				"between 0 and 254 inclusive.",
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
				int64validator.AtMost(254),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"evpn_type_5_routes": resourceSchema.BoolAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: fmt.Sprintf("When `true`, all EVPN VTEPs in the fabric will redistribute "+
				"ARP/IPV6 ND (when possible on NOS type) as EVPN type 5 /32 routes in the routing table. Currently, "+
				"this option is only certified for Juniper Junos. FRR (SONiC) does this implicitly and cannot be "+
				"disabled. This setting will be ignored. On Arista and Cisco, no configuration is rendered and will "+
				"result in a Blueprint warning that it is not supported by AOS. This value is disabled by default, as "+
				"it generates a very large number of routes in the BGP routing table and takes large amounts of TCAM. "+
				"When these /32 & /128 routes are generated, they enable direct unicast routing to host destinations "+
				"on VNIs that are not stretched to the ingress VTEP, and avoid a route lookup to a subnet (eg, /24) "+
				"that may be hosted on many leafs. Requires Apstra %s.", apiversions.Ge420),
		},
		"fabric_mtu": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("MTU of fabric links. Must be an even number between %d and %d. "+
				"Requires Apstra %s.", constants.L3MtuMin, constants.L3MtuMax, apiversions.Ge420),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"ipv6_applications": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables support for IPv6 virtual networks and IPv6 external " +
				"connectivity points. This adds resource requirements and device configurations, " +
				"including IPv6 loopback addresses on leafs, spines and superspines, IPv6 addresses " +
				"for MLAG SVI subnets and IPv6 addresses for leaf L3 peer links. This option cannot " +
				"be disabled without re-creating the Blueprint. Applies only to EVPN blueprints.",
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
				boolplanmodifier.RequiresReplaceIf(
					apstraplanmodifier.BoolRequiresReplaceWhenSwitchingTo(false),
					"Switching from \"false\" to \"true\" requires the Blueprint to be replaced",
					"Switching from `false` to `true` requires the Blueprint to be replaced",
				),
			},
		},
		"junos_evpn_max_nexthop_and_interface_number": resourceSchema.BoolAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: fmt.Sprintf("**Changing this value will result in a disruptive restart of the "+
				"PFE.** Enables configuring the maximum number of nexthops and interface numbers reserved for use in "+
				"EVPN-VXLAN overlay network on Junos leaf devices. AOS default is `true`. Requires Apstra %s",
				apiversions.Ge420),
		},
		"junos_evpn_routing_instance_mode_mac_vrf": resourceSchema.BoolAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: fmt.Sprintf("In releases before 4.2, Apstra used a single default switch "+
				"instance as the configuration model for Junos. In Apstra 4.2, Apstra transitioned to using MAC-VRF for "+
				"all new blueprints and normalized the configuration of Junos to Junos Evolved. This option allows you "+
				"to transition Junos devices to the MAC-VRF configuration model for any blueprints deployed before the "+
				"4.2 release. All models use the VLAN-Aware service type. Requires Apstra %s", apiversions.Ge420),
		},
		"junos_ex_overlay_ecmp": resourceSchema.BoolAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: fmt.Sprintf("**Changing this value will result in a disruptive restart of the "+
				"PFE on EX-series devices.** When `true,`VXLAN Overlay ECMP will be enabled on Junos EX-series devices. "+
				"Requires Apstra %s.", apiversions.Ge420),
		},
		"junos_graceful_restart": resourceSchema.BoolAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: fmt.Sprintf("**Changing this value may result in a flap of all BGP sessions as "+
				"the sessions are re-negotiated.** When `true`, the bgp graceful restart feature is enabled on Junos "+
				"devices. Requires Apstra %s", apiversions.Ge420),
		},
		"max_evpn_routes_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of EVPN routes to accept on Leaf Switches. " +
				"A positive integer will be rendered into the device BGP configuration as a maximum limit. Using a " +
				" zero will render a `0` into the same line of configuration resulting in platform-specific behavior: " +
				"Either *unlimited routes permitted*, or *no routes permitted* depending on the NOS in use. A `-1` " +
				"can be used to force clear any prior configuration from Apstra, ensuring that no maximum value will " +
				"be rendered into the BGP configuration (default device behavior).",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int64{int64validator.Between(-1, math.MaxUint32)},
		},
		"max_external_routes_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of routes to accept from external routers. " +
				"A positive integer will be rendered into the device BGP configuration as a maximum limit. Using a " +
				" zero will render a `0` into the same line of configuration resulting in platform-specific behavior: " +
				"Either *unlimited routes permitted*, or *no routes permitted* depending on the NOS in use. A `-1` " +
				"can be used to force clear any prior configuration from Apstra, ensuring that no maximum value will " +
				"be rendered into the BGP configuration (default device behavior).",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int64{int64validator.Between(-1, math.MaxUint32)},
		},
		"max_fabric_routes_count": resourceSchema.Int64Attribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "Maximum number of underlay routes permitted between fabric nodes. " +
				"A positive integer will be rendered into the device BGP configuration as a maximum limit. Using a " +
				" zero will render a `0` into the same line of configuration resulting in platform-specific behavior: " +
				"Either *unlimited routes permitted*, or *no routes permitted* depending on the NOS in use. A `-1` " +
				"can be used to force clear any prior configuration from Apstra, ensuring that no maximum value will " +
				"be rendered into the BGP configuration (default device behavior)." +
				"Setting this option may be required in the event of leaking EVPN routes from a Security Zone " +
				"into the default Security Zone (VRF) which may generate a large number of /32 and /128 routes. " +
				"It is suggested that this value be effectively unlimited on all Blueprints to ensure BGP stability in " +
				"the underlay. Unlimited is also suggested for non-EVPN Blueprints considering the impact to traffic if " +
				"spine-leaf sessions go offline.",
			Validators: []validator.Int64{int64validator.Between(-1, math.MaxUint32)},
		},
		"max_mlag_routes_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of routes to accept between MLAG peers. " +
				"A positive integer will be rendered into the device BGP configuration as a maximum limit. Using a " +
				" zero will render a `0` into the same line of configuration resulting in platform-specific behavior: " +
				"Either *unlimited routes permitted*, or *no routes permitted* depending on the NOS in use. A `-1` " +
				"can be used to force clear any prior configuration from Apstra, ensuring that no maximum value will " +
				"be rendered into the BGP configuration (default device behavior).",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int64{int64validator.Between(-1, math.MaxUint32)},
		},
		"optimize_routing_zone_footprint": resourceSchema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("When `true`: routing zones will not be rendered on leafs where "+
				"they are not required, resulting in less resource consumption. Requires Apstra %s", apiversions.Ge420),
			Optional: true,
			Computed: true,
		},
	}
}

func (o *Blueprint) LoadApiData(_ context.Context, in *apstra.BlueprintStatus, _ *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)

	o.Status = types.StringValue(in.Status)
	o.SuperspineCount = types.Int64Value(int64(in.SuperspineCount))
	o.SpineCount = types.Int64Value(int64(in.SpineCount))
	o.LeafCount = types.Int64Value(int64(in.LeafCount))
	o.AccessCount = types.Int64Value(int64(in.AccessCount))
	o.GenericCount = types.Int64Value(int64(in.GenericCount))
	o.ExternalCount = types.Int64Value(int64(in.ExternalRouterCount))
	o.HasUncommittedChanges = types.BoolValue(in.HasUncommittedChanges)
	o.Version = types.Int64Value(int64(in.Version))
	o.BuildErrorsCount = types.Int64Value(int64(in.BuildErrorsCount))
	o.BuildWarningsCount = types.Int64Value(int64(in.BuildWarningsCount))
}

func (o *Blueprint) SetName(ctx context.Context, bpClient *apstra.TwoStageL3ClosClient, state *Blueprint, diags *diag.Diagnostics) {
	if o.Name.Equal(state.Name) {
		// nothing to do
		return
	}

	// struct used for GET and PATCH
	type node struct {
		Label string          `json:"label,omitempty"`
		Id    apstra.ObjectId `json:"id,omitempty"`
	}

	// GET target
	response := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}

	err := bpClient.GetNodes(ctx, apstra.NodeTypeMetadata, response)
	if err != nil {
		diags.AddError(
			fmt.Sprintf(errApiGetWithTypeAndId, "Blueprint Node", bpClient.Id()),
			err.Error(),
		)
		return
	}
	if len(response.Nodes) != 1 {
		diags.AddError(fmt.Sprintf("wrong number of %s nodes", apstra.NodeTypeMetadata.String()),
			fmt.Sprintf("expecting 1 got %d nodes", len(response.Nodes)))
		return
	}

	// pull the only value from the map
	var nodeId apstra.ObjectId
	for _, v := range response.Nodes {
		nodeId = v.Id
	}

	err = bpClient.PatchNode(ctx, nodeId, &node{Label: o.Name.ValueString()}, nil)
	if err != nil {
		diags.AddError(
			fmt.Sprintf(errApiPatchWithTypeAndId, bpClient.Id(), nodeId),
			err.Error(),
		)
		return
	}
}

func (o Blueprint) VersionConstraints() apiversions.Constraints {
	var response apiversions.Constraints

	if utils.Known(o.FabricAddressing) {
		response.AddAttributeConstraints(apiversions.AttributeConstraint{
			Path:        path.Root("fabric_addressing"),
			Constraints: apiversions.Ge411,
		})
	}

	if utils.Known(o.DefaultSviL3Mtu) {
		response.AddAttributeConstraints(apiversions.AttributeConstraint{
			Path:        path.Root("default_svi_l3_mtu"),
			Constraints: apiversions.Ge420,
		})
	}

	if utils.Known(o.FabricMtu) {
		response.AddAttributeConstraints(apiversions.AttributeConstraint{
			Path:        path.Root("fabric_mtu"),
			Constraints: apiversions.Ge420,
		})
	}

	if utils.Known(o.JunosEvpnMaxNexthopAndInterfaceNumber) {
		response.AddAttributeConstraints(apiversions.AttributeConstraint{
			Path:        path.Root("junos_evpn_max_nexthop_and_interface_number"),
			Constraints: apiversions.Ge420,
		})
	}

	if utils.Known(o.JunosEvpnRoutingInstanceModeMacVrf) {
		response.AddAttributeConstraints(apiversions.AttributeConstraint{
			Path:        path.Root("junos_evpn_routing_instance_mode_mac_vrf"),
			Constraints: apiversions.Ge420,
		})
	}

	if utils.Known(o.JunosExOverlayEcmp) {
		response.AddAttributeConstraints(apiversions.AttributeConstraint{
			Path:        path.Root("junos_ex_overlay_ecmp"),
			Constraints: apiversions.Ge420,
		})
	}

	if utils.Known(o.JunosGracefulRestart) {
		response.AddAttributeConstraints(apiversions.AttributeConstraint{
			Path:        path.Root("junos_graceful_restart"),
			Constraints: apiversions.Ge420,
		})
	}

	if utils.Known(o.OptimizeRoutingZoneFootprint) {
		response.AddAttributeConstraints(apiversions.AttributeConstraint{
			Path:        path.Root("optimize_routing_zone_footprint"),
			Constraints: apiversions.Ge420,
		})
	}

	return response
}

func (o *Blueprint) GetFabricSettings(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	fabricSettings, err := bp.GetFabricSettings(ctx)
	if err != nil {
		diags.AddError("failed to retrieve fabric settings", err.Error())
		return
	}

	o.LoadFabricSettings(ctx, fabricSettings, diags)
	if diags.HasError() {
		return
	}
}

func (o *Blueprint) SetFabricSettings(ctx context.Context, bp *apstra.TwoStageL3ClosClient, state *Blueprint, diags *diag.Diagnostics) {
	if !o.fabricSettingsNeedsUpdate(state) {
		return
	}

	planFS := o.FabricSettings(ctx, diags)
	if diags.HasError() {
		return
	}

	err := bp.SetFabricSettings(ctx, planFS)
	if err != nil {
		diags.AddError("failed to set fabric settings", err.Error())
		return
	}
}

func (o *Blueprint) LoadFabricSettings(ctx context.Context, settings *apstra.FabricSettings, diags *diag.Diagnostics) {
	o.AntiAffinityMode = types.StringNull()
	o.AntiAffinityPolicy = types.ObjectNull(new(AntiAffinityPolicy).attrTypes())
	if settings.AntiAffinityPolicy != nil {
		o.AntiAffinityMode = types.StringValue(settings.AntiAffinityPolicy.Mode.String())
		o.LoadAntiAffninityPolicy(ctx, settings.AntiAffinityPolicy, diags)
		if diags.HasError() {
			return
		}
	}

	o.DefaultIpLinksToGenericMtu = int64AttrValueFromPtr(settings.ExternalRouterMtu)
	o.DefaultSviL3Mtu = int64AttrValueFromPtr(settings.DefaultSviL3Mtu)
	o.EsiMacMsb = int64AttrValueFromPtr(settings.EsiMacMsb)
	o.EvpnType5Routes = boolAttrValueFromFeatureswitchEnumPtr(settings.EvpnGenerateType5HostRoutes)
	o.FabricMtu = int64AttrValueFromPtr(settings.FabricL3Mtu)
	o.Ipv6Applications = boolAttrValueFromBoolPtr(settings.Ipv6Enabled)
	o.JunosEvpnMaxNexthopAndInterfaceNumber = boolAttrValueFromFeatureswitchEnumPtr(settings.JunosEvpnMaxNexthopAndInterfaceNumber)
	o.JunosEvpnRoutingInstanceModeMacVrf = boolAttrValueFromFeatureswitchEnumPtr(settings.JunosEvpnRoutingInstanceVlanAware)
	o.JunosExOverlayEcmp = boolAttrValueFromFeatureswitchEnumPtr(settings.JunosExOverlayEcmp)
	o.JunosGracefulRestart = boolAttrValueFromFeatureswitchEnumPtr(settings.JunosGracefulRestart)
	o.MaxEvpnRoutesCount = parseRouteLimit(settings.MaxEvpnRoutes)
	o.MaxExternalRoutesCount = parseRouteLimit(settings.MaxExternalRoutes)
	o.MaxFabricRoutesCount = parseRouteLimit(settings.MaxFabricRoutes)
	o.MaxMlagRoutesCount = parseRouteLimit(settings.MaxMlagRoutes)
	o.OptimizeRoutingZoneFootprint = boolAttrValueFromFeatureswitchEnumPtr(settings.OptimiseSzFootprint)
}

func (o *Blueprint) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.CreateBlueprintFromTemplateRequest {
	result := apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:      apstra.RefDesignTwoStageL3Clos,
		Label:          o.Name.ValueString(),
		TemplateId:     apstra.ObjectId(o.TemplateId.ValueString()),
		FabricSettings: o.FabricSettings(ctx, diags),
	}

	if utils.Known(o.FabricAddressing) {
		result.FabricSettings.SpineLeafLinks = utils.FabricAddressing(ctx, o.FabricAddressing,
			utils.ToPtr(path.Root("fabric_addressing")), diags)
		if diags.HasError() {
			return nil
		}

		result.FabricSettings.SpineSuperspineLinks = utils.FabricAddressing(ctx, o.FabricAddressing,
			utils.ToPtr(path.Root("fabric_addressing")), diags)
		if diags.HasError() {
			return nil
		}
	}

	return &result
}

func (o *Blueprint) LoadAntiAffninityPolicy(ctx context.Context, antiAffinitypolicy *apstra.AntiAffinityPolicy, diags *diag.Diagnostics) {
	var policy AntiAffinityPolicy
	policy.loadApiData(ctx, antiAffinitypolicy, diags)
	if diags.HasError() {
		return
	}

	var d diag.Diagnostics
	o.AntiAffinityPolicy, d = types.ObjectValueFrom(ctx, policy.attrTypes(), policy)
	diags.Append(d...)
}

func (o *Blueprint) FabricSettings(ctx context.Context, diags *diag.Diagnostics) *apstra.FabricSettings {
	var result apstra.FabricSettings

	if utils.Known(o.AntiAffinityMode) && utils.Known(o.AntiAffinityPolicy) {
		var aap AntiAffinityPolicy
		diags.Append(o.AntiAffinityPolicy.As(ctx, &aap, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil
		}

		result.AntiAffinityPolicy = &apstra.AntiAffinityPolicy{
			Algorithm: apstra.AlgorithmHeuristic,
			// Mode:                     0, // handled below
			MaxLinksPerPort:          int(aap.MaxLinksCountPerPort.ValueInt64()),
			MaxLinksPerSlot:          int(aap.MaxLinksCountPerSlot.ValueInt64()),
			MaxPerSystemLinksPerPort: int(aap.MaxLinksCountPerSystemPerPort.ValueInt64()),
			MaxPerSystemLinksPerSlot: int(aap.MaxLinksCountPerSystemPerSlot.ValueInt64()),
		}

		err := result.AntiAffinityPolicy.Mode.FromString(o.AntiAffinityMode.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to process antiafinity mode %s", o.AntiAffinityMode), err.Error())
			return nil
		}
	}

	if utils.Known(o.DefaultSviL3Mtu) {
		result.DefaultSviL3Mtu = utils.ToPtr(uint16(o.DefaultSviL3Mtu.ValueInt64()))
	}

	if utils.Known(o.EsiMacMsb) {
		result.EsiMacMsb = utils.ToPtr(uint8(o.EsiMacMsb.ValueInt64()))
	}

	if utils.Known(o.EvpnType5Routes) {
		result.EvpnGenerateType5HostRoutes = &apstra.FeatureSwitchEnumDisabled
		if o.EvpnType5Routes.ValueBool() {
			result.EvpnGenerateType5HostRoutes = &apstra.FeatureSwitchEnumEnabled
		}
	}

	if utils.Known(o.DefaultIpLinksToGenericMtu) {
		result.ExternalRouterMtu = utils.ToPtr(uint16(o.DefaultIpLinksToGenericMtu.ValueInt64()))
	}

	if utils.Known(o.FabricMtu) {
		result.FabricL3Mtu = utils.ToPtr(uint16(o.FabricMtu.ValueInt64()))
	}

	if utils.Known(o.Ipv6Applications) {
		result.Ipv6Enabled = utils.ToPtr(o.Ipv6Applications.ValueBool())
	}

	if utils.Known(o.JunosEvpnMaxNexthopAndInterfaceNumber) {
		result.JunosEvpnMaxNexthopAndInterfaceNumber = &apstra.FeatureSwitchEnumDisabled
		if o.JunosEvpnMaxNexthopAndInterfaceNumber.ValueBool() {
			result.JunosEvpnMaxNexthopAndInterfaceNumber = &apstra.FeatureSwitchEnumEnabled
		}
	}

	if utils.Known(o.JunosEvpnRoutingInstanceModeMacVrf) {
		result.JunosEvpnRoutingInstanceVlanAware = &apstra.FeatureSwitchEnumDisabled
		if o.JunosEvpnRoutingInstanceModeMacVrf.ValueBool() {
			result.JunosEvpnRoutingInstanceVlanAware = &apstra.FeatureSwitchEnumEnabled
		}
	}

	if utils.Known(o.JunosExOverlayEcmp) {
		result.JunosExOverlayEcmp = &apstra.FeatureSwitchEnumDisabled
		if o.JunosExOverlayEcmp.ValueBool() {
			result.JunosExOverlayEcmp = &apstra.FeatureSwitchEnumEnabled
		}
	}

	if utils.Known(o.JunosGracefulRestart) {
		result.JunosGracefulRestart = &apstra.FeatureSwitchEnumDisabled
		if o.JunosGracefulRestart.ValueBool() {
			result.JunosGracefulRestart = &apstra.FeatureSwitchEnumEnabled
		}
	}

	if utils.Known(o.MaxEvpnRoutesCount) {
		result.MaxEvpnRoutes = utils.ToPtr(uint32(o.MaxEvpnRoutesCount.ValueInt64()))
	}

	if utils.Known(o.MaxExternalRoutesCount) {
		result.MaxExternalRoutes = utils.ToPtr(uint32(o.MaxExternalRoutesCount.ValueInt64()))
	}

	if utils.Known(o.MaxFabricRoutesCount) {
		result.MaxFabricRoutes = utils.ToPtr(uint32(o.MaxFabricRoutesCount.ValueInt64()))
	}

	if utils.Known(o.MaxMlagRoutesCount) {
		result.MaxMlagRoutes = utils.ToPtr(uint32(o.MaxMlagRoutesCount.ValueInt64()))
	}

	if utils.Known(o.OptimizeRoutingZoneFootprint) {
		result.OptimiseSzFootprint = &apstra.FeatureSwitchEnumDisabled
		if o.OptimizeRoutingZoneFootprint.ValueBool() {
			result.OptimiseSzFootprint = &apstra.FeatureSwitchEnumEnabled
		}
	}

	return &result
}

func (o *Blueprint) fabricSettingsNeedsUpdate(state *Blueprint) bool {
	if state == nil {
		return true
	}

	if !o.AntiAffinityMode.Equal(state.AntiAffinityMode) ||
		!o.AntiAffinityPolicy.Equal(state.AntiAffinityPolicy) ||
		!o.DefaultIpLinksToGenericMtu.Equal(state.DefaultIpLinksToGenericMtu) ||
		!o.DefaultSviL3Mtu.Equal(state.DefaultSviL3Mtu) ||
		!o.EsiMacMsb.Equal(state.EsiMacMsb) ||
		!o.EvpnType5Routes.Equal(state.EvpnType5Routes) ||
		!o.FabricAddressing.Equal(state.FabricAddressing) ||
		!o.Ipv6Applications.Equal(state.Ipv6Applications) ||
		!o.JunosEvpnMaxNexthopAndInterfaceNumber.Equal(state.JunosEvpnMaxNexthopAndInterfaceNumber) ||
		!o.JunosEvpnRoutingInstanceModeMacVrf.Equal(state.JunosEvpnRoutingInstanceModeMacVrf) ||
		!o.JunosExOverlayEcmp.Equal(state.JunosExOverlayEcmp) ||
		!o.JunosGracefulRestart.Equal(state.JunosGracefulRestart) ||
		!o.MaxEvpnRoutesCount.Equal(state.MaxEvpnRoutesCount) ||
		!o.MaxExternalRoutesCount.Equal(state.MaxExternalRoutesCount) ||
		!o.MaxFabricRoutesCount.Equal(state.MaxFabricRoutesCount) ||
		!o.MaxMlagRoutesCount.Equal(state.MaxMlagRoutesCount) ||
		!o.OptimizeRoutingZoneFootprint.Equal(state.OptimizeRoutingZoneFootprint) {
		return true
	}

	return false
}

func parseRouteLimit(i *uint32) types.Int64 {
	if i == nil {
		return types.Int64Value(-1)
	}

	return types.Int64Value(int64(*i))
}

func boolAttrValueFromBoolPtr(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}

	return types.BoolValue(*b)
}

func boolAttrValueFromFeatureswitchEnumPtr(fs *apstra.FeatureSwitchEnum) types.Bool {
	if fs == nil {
		return types.BoolNull()
	}

	return types.BoolValue(fs.Value == apstra.FeatureSwitchEnumEnabled.Value)
}

func int64AttrValueFromPtr[A constraints.Integer](a *A) types.Int64 {
	if a == nil {
		return types.Int64Null()
	}

	return types.Int64Value(int64(*a))
}
