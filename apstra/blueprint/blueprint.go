package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	apstraplanmodifier "github.com/Juniper/terraform-provider-apstra/apstra/apstra_plan_modifier"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	"strings"
)

type Blueprint struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	TemplateId            types.String `tfsdk:"template_id"`
	FabricAddressing      types.String `tfsdk:"fabric_addressing"`
	Status                types.String `tfsdk:"status"`
	SuperspineCount       types.Int64  `tfsdk:"superspine_count"`
	SpineCount            types.Int64  `tfsdk:"spine_count"`
	LeafCount             types.Int64  `tfsdk:"leaf_switch_count"`
	AccessCount           types.Int64  `tfsdk:"access_switch_count"`
	GenericCount          types.Int64  `tfsdk:"generic_system_count"`
	ExternalCount         types.Int64  `tfsdk:"external_router_count"`
	HasUncommittedChanges types.Bool   `tfsdk:"has_uncommitted_changes"`
	Version               types.Int64  `tfsdk:"version"`
	BuildWarningsCount    types.Int64  `tfsdk:"build_warnings_count"`
	BuildErrorsCount      types.Int64  `tfsdk:"build_errors_count"`
	EsiMacMsb             types.Int64  `tfsdk:"esi_mac_msb"`
	// MTU Settings
	FabricMtu                         types.Int64 `tfsdk:"fabric_mtu"`
	DefaultIPLinksToGenericSystemsMTU types.Int64 `tfsdk:"default_ip_links_to_generic_systems_mtu"`
	DefaultSviL3Mtu                   types.Int64 `tfsdk:"default_svi_l3_mtu"`
	// Fabric Design
	Ipv6Applications                 types.Bool   `tfsdk:"ipv6_applications"`
	RoutingZoneFootprintOptimization types.String `tfsdk:"routing_zone_footprint_optimization"`
	// Route Options
	MaxExternalRoutesCount types.Int64  `tfsdk:"max_external_routes_count"`
	MaxMlagRoutesCount     types.Int64  `tfsdk:"max_mlag_routes_count"`
	MaxEvpnRoutesCount     types.Int64  `tfsdk:"max_evpn_routes_count"`
	MaxFabricRoutesCount   types.Int64  `tfsdk:"max_fabric_routes_count"`
	EvpnType5Routes        types.String `tfsdk:"evpn_type_5_routes"`
	// Vendor Specific
	JunosEvpnRoutingInstanceMode          types.String `tfsdk:"junos_evpn_routing_instance_mode"`
	JunosEvpnMaxNexthopAndInterfaceNumber types.String `tfsdk:"junos_evpn_max_nexthop_and_interface_number"`
	JunosGracefulRestart                  types.String `tfsdk:"junos_graceful_restart"`
	JunosExOverlayEcmp                    types.String `tfsdk:"junos_ex_overlay_ecmp"`
	// Anti Affinity
	//AntiAffinityMode types.String `tfsdk:"anti_affinity_mode"`
}

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
			MarkdownDescription: "Template ID will always be null in 'data source' context.",
			Computed:            true,
		},
		"fabric_addressing": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Addressing scheme for both superspine/spine and spine/leaf links.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status of the Blueprint",
			Computed:            true,
		},
		"superspine_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
			Computed:            true,
		},
		"spine_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of spine devices in the topology.",
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
			MarkdownDescription: "The count of external routers attached to the topology.",
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
		"build_warnings_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build warnings.",
			Computed:            true,
		},
		"build_errors_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build errors.",
			Computed:            true,
		},
		"esi_mac_msb": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "ESI MAC address most significant byte.",
			Computed:            true,
		},
		"ipv6_applications": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables support for IPv6 virtual networks and IPv6 external " +
				"connectivity points. This adds resource requirements and device configurations, " +
				"including IPv6 loopback addresses on leafs, spines and superspines, IPv6 addresses " +
				"for MLAG SVI subnets and IPv6 addresses for leaf L3 peer links.",
			Computed: true,
		},
		"fabric_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "MTU of fabric links. Requires Apstra 4.2 or later.",
			Computed:            true,
		},
		"default_ip_links_to_generic_systems_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Default L3 MTU for IP links to generic systems. A null or empty value " +
				"implies AOS will not render explicit MTU value and system defaults will be used. " +
				"Should be an even number in a range 1280..9216.",
			Computed: true,
		},
		"default_svi_l3_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Default L3 MTU for SVI interfaces. Should be an even number in a range 1280..9216" +
				"Requires Apstra 4.2 or later.",
			Computed: true,
		},
		"routing_zone_footprint_optimization": dataSourceSchema.StringAttribute{
			MarkdownDescription: "When enabled: routing zones will not be rendered on leafs where it is not required," +
				"which results in less resource consumption. Routing zone will only be rendered for systems which have " +
				"other structures configured on top of routing zone, such as virtual networks, protocol sessions, " +
				"static routes, sub-interfaces, etc. Requires Apstra 4.2 or Later",
			Computed: true,
		},
		"max_external_routes_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of routes to accept from external routers. The default (None) will not " +
				"render any maximum-route commands on BGP sessions, implying that only vendor defaults are used." +
				"An integer between 1-2**32-1 will set a maximum limit of routes in BGP config. The value 0 (zero)" +
				"intends the device to never apply a limit to number of EVPN routes (effectively unlimited). " +
				"It is suggested this value is value is effectively unlimited on evpn blueprints, to permit the " +
				"high number of /32 and /128 routes to be advertised and received between VRFs in the event an " +
				"external router is providing a form of route leaking functionality.",
			Computed: true,
		},
		"max_mlag_routes_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of routes to accept across MLAG peer switches. The default (None) will" +
				" not render any maximum-route commands on BGP sessions, implying that only vendor defaults are used. " +
				"An integer between 1-2**32-1 will set a maximum limit of routes in BGP config. The value 0 (zero) " +
				"intends the device to never apply a limit to number of EVPN routes (effectively unlimited). " +
				"Note: Device vendors typically shut down BGP sessions if maximums are exceeded on a session. " +
				"For EVPN blueprints, this should be combined with max_evpn_routes to permit routes across the " +
				"l3 peer link which may contain many /32 and /128 from EVPN type-2 routes that convert into " +
				"BGP route advertisements.",
			Computed: true,
		},
		"max_evpn_routes_count": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Maximum number of EVPN routes to accept on an EVPN switch. The default (None) will " +
				"not render any maximum-route commands on BGP sessions, implying that only vendor defaults are used. " +
				"An integer between 1-2**32-1 will set a maximum limit of routes in BGP config. The value 0 (zero) " +
				"intends the device to never apply a limit to number of EVPN routes (effectively unlimited). " +
				"Note: Device vendors typically shut down BGP sessions if maximums are exceeded on a session.",
		},
		"max_fabric_routes_count": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Maximum number of routes to accept between spine and leaf in the fabric, " +
				"and spine-superspine. This includes the default VRF. Setting this option may be required in the" +
				" event of leaking EVPN routes from a security zone into the default security zone (VRF) which " +
				"could generate a large number of /32 and /128 routes. It is suggested that this value is effectively " +
				"unlimited on all blueprints to ensure the network stability of spine-leaf bgp sessions and evpn " +
				"underlay. Unlimited is also suggested for non-evpn blueprints considering the impact to traffic if " +
				"spine-leaf sessions go offline. An integer between 1-2**32-1 will set a maximum limit of routes in " +
				"BGP config. The value 0 (zero) intends the device to never apply a limit to number of fabric routes " +
				"(effectively unlimited).",
		},
		"evpn_type_5_routes": dataSourceSchema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Default disabled. When enabled all EVPN vteps in the fabric will redistribute " +
				"ARP/IPV6 ND (when possible on NOS type) as EVPN type 5 /32 routes in the routing table. " +
				"Currently, this option is only certified for Juniper JunOS. FRR (SONiC) does this implicitly " +
				"and cannot be disabled. This setting will be ignored. On Arista and Cisco, no configuration is " +
				"rendered and will result in a blueprint warning that it is not supported by AOS. This value is " +
				"disabled by default, as it generates a very large number of routes in the BGP routing table and " +
				"takes large amounts of TCAM allocation space. When these /32 & /128 routes are generated, it assists " +
				"in direct unicast routing to host destinations on VNIs that are not stretched to the ingress vtep, " +
				"and avoids a route lookup to a subnet (eg, /24) that may be hosted on many leafs. The directed host " +
				"route prevents a double lookup to one of many vteps may hosts the /24 and instead routes the " +
				"destination directly to the correct vtep.",
		},
		"junos_evpn_routing_instance_mode": dataSourceSchema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Changing this value will result in a complete restart of all " +
				"EVPN processes on the entire fabric." +
				"In releases before 4.2, Apstra used a single default switch instance as the " +
				"configuration model for Junos. In Apstra 4.2, Apstra transitioned to using MAC-VRF for all new " +
				"blueprints and normalized the configuration of Junos to Junos Evolved. This option allows you to " +
				"transition Junos devices to the MAC-VRF configuration model for any blueprints deployed before the " +
				"4.2 release. All models use the VLAN-Aware service type.",
		},
		"junos_evpn_max_nexthop_and_interface_number": dataSourceSchema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Changing this value will result in a disruptive restart of the PFE." +
				"Enables configuring the maximum number of nexthops and interface numbers reserved " +
				"for use in EVPN-VXLAN overlay network on Junos leaf devices. Default is enabled.",
		},
		"junos_graceful_restart": dataSourceSchema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Changing this value may result in a flap of all BGP sessions as the sessions are re-negotiated" +
				"Enables the Graceful Restart feature on Junos devices",
		},
		"junos_ex_overlay_ecmp": dataSourceSchema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Changing this value will result in a disruptive restart of the PFE on EX-series devices" +
				"Enables VXLAN Overlay ECMP on Junos EX-series devices",
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
			Computed: true,
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
		"superspine_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
			Computed:            true,
		},
		"spine_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of spine devices in the topology.",
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
		"build_warnings_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build warnings.",
			Computed:            true,
		},
		"build_errors_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build errors.",
			Computed:            true,
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
		"ipv6_applications": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables support for IPv6 virtual networks and IPv6 external " +
				"connectivity points. This adds resource requirements and device configurations, " +
				"including IPv6 loopback addresses on leafs, spines and superspines, IPv6 addresses " +
				"for MLAG SVI subnets and IPv6 addresses for leaf L3 peer links. This option cannot " +
				"be disabled without re-creating the Blueprint.",
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
		"fabric_mtu": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("MTU of fabric links. Must be an even number between %d and %d. "+
				"Requires Apstra %s or later.", constants.L3MtuMin, constants.L3MtuMax, apiversions.Apstra420),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"default_ip_links_to_generic_systems_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("Default L3 MTU for IP links to generic systems. A null or empty value "+
				"implies AOS will not render explicit MTU value and system defaults will be used. "+
				"Should be an even number between %d and %d.", constants.L3MtuMin, constants.L3MtuMax),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"default_svi_l3_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("Default L3 MTU for SVI interfaces. Should be an even number in a range %d and %d."+
				"Requires Apstra 4.2 or later.", constants.L3MtuMin, constants.L3MtuMax),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"routing_zone_footprint_optimization": dataSourceSchema.StringAttribute{
			MarkdownDescription: "When enabled: routing zones will not be rendered on leafs where it is not required," +
				"which results in less resource consumption. Routing zone will only be rendered for systems which have " +
				"other structures configured on top of routing zone, such as virtual networks, protocol sessions, " +
				"static routes, sub-interfaces, etc. Requires Apstra 4.2 or Later",
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.OneOf(apstra.FeatureSwitchEnumEnabled.String(), apstra.FeatureSwitchEnumDisabled.String()),
			},
		},
		"max_external_routes_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of routes to accept from external routers. The default (None) will not " +
				"render any maximum-route commands on BGP sessions, implying that only vendor defaults are used." +
				"An integer between 1-2**32-1 will set a maximum limit of routes in BGP config. The value 0 (zero)" +
				"intends the device to never apply a limit to number of EVPN routes (effectively unlimited). " +
				"It is suggested this value is value is effectively unlimited on evpn blueprints, to permit the " +
				"high number of /32 and /128 routes to be advertised and received between VRFs in the event an " +
				"external router is providing a form of route leaking functionality.",
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(0, 4294967295),
			},
		},
		"max_mlag_routes_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of routes to accept across MLAG peer switches. The default (None) will" +
				" not render any maximum-route commands on BGP sessions, implying that only vendor defaults are used. " +
				"An integer between 1-2**32-1 will set a maximum limit of routes in BGP config. The value 0 (zero) " +
				"intends the device to never apply a limit to number of EVPN routes (effectively unlimited). " +
				"Note: Device vendors typically shut down BGP sessions if maximums are exceeded on a session. " +
				"For EVPN blueprints, this should be combined with max_evpn_routes to permit routes across the " +
				"l3 peer link which may contain many /32 and /128 from EVPN type-2 routes that convert into " +
				"BGP route advertisements.",
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(0, 4294967295),
			},
		},
		"max_evpn_routes_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Maximum number of EVPN routes to accept on an EVPN switch. The default (None) will " +
				"not render any maximum-route commands on BGP sessions, implying that only vendor defaults are used. " +
				"An integer between 1-2**32-1 will set a maximum limit of routes in BGP config. The value 0 (zero) " +
				"intends the device to never apply a limit to number of EVPN routes (effectively unlimited). " +
				"Note: Device vendors typically shut down BGP sessions if maximums are exceeded on a session.",
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(0, 4294967295),
			},
		},
		"max_fabric_routes_count": dataSourceSchema.Int64Attribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "Maximum number of routes to accept between spine and leaf in the fabric, " +
				"and spine-superspine. This includes the default VRF. Setting this option may be required in the" +
				" event of leaking EVPN routes from a security zone into the default security zone (VRF) which " +
				"could generate a large number of /32 and /128 routes. It is suggested that this value is effectively " +
				"unlimited on all blueprints to ensure the network stability of spine-leaf bgp sessions and evpn " +
				"underlay. Unlimited is also suggested for non-evpn blueprints considering the impact to traffic if " +
				"spine-leaf sessions go offline. An integer between 1-2**32-1 will set a maximum limit of routes in " +
				"BGP config. The value 0 (zero) intends the device to never apply a limit to number of fabric routes " +
				"(effectively unlimited).",
			Validators: []validator.Int64{
				int64validator.Between(0, 4294967295),
			},
		},
		"evpn_type_5_routes": dataSourceSchema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "Default disabled. When enabled all EVPN vteps in the fabric will redistribute " +
				"ARP/IPV6 ND (when possible on NOS type) as EVPN type 5 /32 routes in the routing table. " +
				"Currently, this option is only certified for Juniper JunOS. FRR (SONiC) does this implicitly " +
				"and cannot be disabled. This setting will be ignored. On Arista and Cisco, no configuration is " +
				"rendered and will result in a blueprint warning that it is not supported by AOS. This value is " +
				"disabled by default, as it generates a very large number of routes in the BGP routing table and " +
				"takes large amounts of TCAM allocation space. When these /32 & /128 routes are generated, it assists " +
				"in direct unicast routing to host destinations on VNIs that are not stretched to the ingress vtep, " +
				"and avoids a route lookup to a subnet (eg, /24) that may be hosted on many leafs. The directed host " +
				"route prevents a double lookup to one of many vteps may hosts the /24 and instead routes the " +
				"destination directly to the correct vtep.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"junos_evpn_routing_instance_mode": dataSourceSchema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "In releases before 4.2, Apstra used a single default switch instance as the " +
				"configuration model for Junos. In Apstra 4.2, Apstra transitioned to using MAC-VRF for all new " +
				"blueprints and normalized the configuration of Junos to Junos Evolved. This option allows you to " +
				"transition Junos devices to the MAC-VRF configuration model for any blueprints deployed before the " +
				"4.2 release. All models use the VLAN-Aware service type.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"junos_evpn_max_nexthop_and_interface_number": dataSourceSchema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "Changing this value will result in a disruptive restart of the PFE." +
				"Enables configuring the maximum number of nexthops and interface numbers reserved " +
				"for use in EVPN-VXLAN overlay network on Junos leaf devices. Default is enabled.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"junos_graceful_restart": dataSourceSchema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "Changing this value may result in a flap of all BGP sessions as the sessions are re-negotiated" +
				"Enables the Graceful Restart feature on Junos devices",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"junos_ex_overlay_ecmp": dataSourceSchema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "Changing this value will result in a disruptive restart of the PFE on EX-series devices" +
				"Enables VXLAN Overlay ECMP on Junos EX-series devices",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	}
}

func (o Blueprint) Request(_ context.Context, diags *diag.Diagnostics) *apstra.CreateBlueprintFromTemplateRequest {
	// start with a nil pointer for fabric addressing policy
	var fabricAddressingPolicy *apstra.BlueprintRequestFabricAddressingPolicy

	// if the user supplied either value, we must create the fabric addressing policy object
	if !o.FabricAddressing.IsUnknown() || !o.FabricMtu.IsUnknown() {
		fabricAddressingPolicy = &apstra.BlueprintRequestFabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4, // sensible default
			SpineLeafLinks:       apstra.AddressingSchemeIp4, // sensible default
		}

		if utils.Known(o.FabricAddressing) {
			var fabricAddressing apstra.AddressingScheme
			err := fabricAddressing.FromString(o.FabricAddressing.ValueString())
			if err != nil {
				diags.AddError(
					constants.ErrProviderBug,
					fmt.Sprintf("error parsing fabric_addressing %q - %s",
						o.FabricAddressing.ValueString(), err.Error()))
				return nil
			}
			fabricAddressingPolicy.SpineSuperspineLinks = fabricAddressing
			fabricAddressingPolicy.SpineLeafLinks = fabricAddressing
		}

		if utils.Known(o.FabricMtu) {
			fabricMtu := uint16(o.FabricMtu.ValueInt64())
			fabricAddressingPolicy.FabricL3Mtu = &fabricMtu
		}
	}

	return &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:              apstra.RefDesignTwoStageL3Clos,
		Label:                  o.Name.ValueString(),
		TemplateId:             apstra.ObjectId(o.TemplateId.ValueString()),
		FabricAddressingPolicy: fabricAddressingPolicy,
	}
}

func (o Blueprint) FabricAddressingRequest(_ context.Context, _ *diag.Diagnostics) *apstra.TwoStageL3ClosFabricAddressingPolicy {
	var result apstra.TwoStageL3ClosFabricAddressingPolicy

	if utils.Known(o.Ipv6Applications) {
		ipv6Enabled := o.Ipv6Applications.ValueBool()
		result.Ipv6Enabled = &ipv6Enabled
	}

	if utils.Known(o.EsiMacMsb) {
		esiMacMsb := uint8(o.EsiMacMsb.ValueInt64())
		result.EsiMacMsb = &esiMacMsb
	}

	if utils.Known(o.FabricMtu) {
		fabricMtu := uint16(o.FabricMtu.ValueInt64())
		result.FabricL3Mtu = &fabricMtu
	}

	return &result
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

func (o *Blueprint) LoadFabricAddressingPolicy(_ context.Context, in *apstra.TwoStageL3ClosFabricAddressingPolicy, _ *diag.Diagnostics) {
	o.EsiMacMsb = types.Int64Null()
	if in.EsiMacMsb != nil {
		o.EsiMacMsb = types.Int64Value(int64(*in.EsiMacMsb))
	}

	o.FabricMtu = types.Int64Null()
	if in.FabricL3Mtu != nil {
		o.FabricMtu = types.Int64Value(int64(*in.FabricL3Mtu))
	}

	o.Ipv6Applications = types.BoolValue(false)
	if in.Ipv6Enabled != nil {
		o.Ipv6Applications = types.BoolValue(*in.Ipv6Enabled)
	}
}

func (o *Blueprint) SetName(ctx context.Context, bpClient *apstra.TwoStageL3ClosClient, state *Blueprint, diags *diag.Diagnostics) {
	if o.Name.Equal(state.Name) {
		// nothing to do
		return
	}

	type node struct {
		Label string          `json:"label,omitempty"`
		Id    apstra.ObjectId `json:"id,omitempty"`
	}
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

func (o *Blueprint) SetFabricAddressingPolicy(ctx context.Context, bpClient *apstra.TwoStageL3ClosClient, state *Blueprint, diags *diag.Diagnostics) {
	switch {
	case utils.Known(o.EsiMacMsb): // we have a value; do not return in default action
	case utils.Known(o.FabricMtu): // we have a value; do not return in default action
	case utils.Known(o.Ipv6Applications): // we have a value; do not return in default action
	default:
		return // no relevant values set in the plan
	}

	if state != nil {
		switch {
		case utils.Known(o.EsiMacMsb) && !o.EsiMacMsb.Equal(state.EsiMacMsb): // plan and state not in agreement
		case utils.Known(o.FabricMtu) && !o.FabricMtu.Equal(state.FabricMtu): // plan and state not in agreement
		case utils.Known(o.Ipv6Applications) && !o.Ipv6Applications.Equal(state.Ipv6Applications): // plan and state not in agreement
		default:
			return // no plan values represent changes from the current state
		}
	}

	fapRequest := o.FabricAddressingRequest(ctx, diags)
	if diags.HasError() {
		return
	}

	if fapRequest == nil {
		// nothing to do
		return
	}

	err := bpClient.SetFabricAddressingPolicy(ctx, fapRequest)
	if err != nil {
		diags.AddError("failed setting blueprint fabric addressing policy", err.Error())
		return
	}
}

func (o *Blueprint) GetFabricLinkAddressing(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	query := new(apstra.PathQuery).
		SetClient(bp.Client()).
		SetBlueprintId(bp.Id()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeFabricAddressingPolicy.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_fabric_addressing_policy")},
		})

	var result struct {
		Items []struct {
			FabricAddressingPolicy struct {
				SpineLeafLinks       string `json:"spine_leaf_links"`
				SpineSuperspineLinks string `json:"spine_superspine_links"`
			} `json:"n_fabric_addressing_policy"`
		} `json:"items"`
	}

	err := query.Do(ctx, &result)
	if err != nil {
		diags.AddError("failed querying for blueprint fabric addressing policy", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		diags.AddError(
			"failed querying for blueprint fabric addressing policy",
			fmt.Sprintf("query produced no results: %q", query.String()))
		return
	case 1:
		// expected case handled below
	default:
		diags.AddError(
			"failed querying for blueprint fabric addressing policy",
			fmt.Sprintf("query produced %d results (expected 1): %q", len(result.Items), query.String()))
		return
	}

	if result.Items[0].FabricAddressingPolicy.SpineLeafLinks != result.Items[0].FabricAddressingPolicy.SpineSuperspineLinks {
		diags.AddError(
			"failed querying for blueprint fabric addressing policy",
			fmt.Sprintf("spine_leaf_links addressing does not match spine_superspine_links addressing:\n"+
				"%q vs. %q\nquery: %q",
				result.Items[0].FabricAddressingPolicy.SpineLeafLinks,
				result.Items[0].FabricAddressingPolicy.SpineSuperspineLinks,
				query.String()))
		return
	}

	var addressingScheme apstra.AddressingScheme
	err = addressingScheme.FromString(result.Items[0].FabricAddressingPolicy.SpineLeafLinks)
	if err != nil {
		diags.AddError("failed to parse fabric addressing", err.Error())
		return
	}

	o.FabricAddressing = types.StringValue(addressingScheme.String())
}

func (o Blueprint) VersionConstraints() apiversions.Constraints {
	var response apiversions.Constraints

	if utils.Known(o.FabricAddressing) {
		response.AddAttributeConstraints(
			apiversions.AttributeConstraint{
				Path:        path.Root("fabric_addressing"),
				Constraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra411)),
			},
		)
	}

	if utils.Known(o.FabricMtu) {
		response.AddAttributeConstraints(
			apiversions.AttributeConstraint{
				Path:        path.Root("fabric_mtu"),
				Constraints: version.MustConstraints(version.NewConstraint(">=" + apiversions.Apstra420)),
			},
		)
	}

	if utils.Known(o.DefaultIPLinksToGenericSystemsMTU) {
		response.AddAttributeConstraints(
			path.Root("default_ip_links_to_generic_systems_mtu"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}

	if utils.Known(o.DefaultSviL3Mtu) {
		response.AddAttributeConstraints(
			path.Root("default_svi_l3_mtu"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}

	if utils.Known(o.RoutingZoneFootprintOptimization) {
		response.AddAttributeConstraints(
			path.Root("routing_zone_footprint_optimization"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}

	if utils.Known(o.MaxExternalRoutesCount) {
		response.AddAttributeConstraints(
			path.Root("max_external_routes_count"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}

	if utils.Known(o.MaxMlagRoutesCount) {
		response.AddAttributeConstraints(
			path.Root("max_mlag_routes_count"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}

	if utils.Known(o.MaxEvpnRoutesCount) {
		response.AddAttributeConstraints(
			path.Root("max_evpn_routes_count"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}

	if utils.Known(o.MaxFabricRoutesCount) {
		response.AddAttributeConstraints(
			path.Root("max_fabric_routes_count"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}

	if utils.Known(o.EvpnType5Routes) {
		response.AddAttributeConstraints(
			path.Root("evpn_type_5_routes"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}
	if utils.Known(o.JunosEvpnRoutingInstanceMode) {
		response.AddAttributeConstraints(
			path.Root("junos_evpn_routing_instance_mode"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}
	if utils.Known(o.JunosEvpnMaxNexthopAndInterfaceNumber) {
		response.AddAttributeConstraints(
			path.Root("junos_evpn_max_nexthop_and_interface_number"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}
	if utils.Known(o.JunosGracefulRestart) {
		response.AddAttributeConstraints(
			path.Root("junos_graceful_restart"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}
	if utils.Known(o.JunosExOverlayEcmp) {
		response.AddAttributeConstraints(
			path.Root("junos_ex_overlay_ecmp"),
			version.MustConstraints(version.NewConstraint(">="+apiversions.Apstra420)))
	}
	return response
}

func (o Blueprint) fabricSettings() *apstra.FabricSettings {
	var fabricSettings apstra.FabricSettings
	var valueFound bool
	if utils.Known(o.DefaultIPLinksToGenericSystemsMTU) {
		fabricSettings.ExternalRouterMtu = utils.ToPtr(uint16(o.DefaultIPLinksToGenericSystemsMTU.ValueInt64()))
		valueFound = true
	}
	if utils.Known(o.EsiMacMsb) {
		fabricSettings.EsiMacMsb = utils.ToPtr(uint8(o.EsiMacMsb.ValueInt64()))
		valueFound = true
	}
	if utils.Known(o.DefaultSviL3Mtu) {
		fabricSettings.DefaultSviL3Mtu = utils.ToPtr(uint16(o.DefaultSviL3Mtu.ValueInt64()))
		valueFound = true
	}
	if utils.Known(o.RoutingZoneFootprintOptimization) {
		fabricSettings.OptimiseSzFootprint = utils.ToPtr(apstra.FeatureSwitchEnum{o.RoutingZoneFootprintOptimization})
		valueFound = true
	}
	if utils.Known(o.MaxExternalRoutesCount) {
		fabricSettings.MaxExternalRoutes = utils.ToPtr(uint32(o.MaxExternalRoutesCount.ValueInt64()))
		valueFound = true
	}
	if utils.Known(o.MaxMlagRoutesCount) {
		fabricSettings.MaxMlagRoutes = utils.ToPtr(uint32(o.MaxMlagRoutesCount.ValueInt64()))
		valueFound = true
	}
	if utils.Known(o.MaxEvpnRoutesCount) {
		fabricSettings.MaxEvpnRoutes = utils.ToPtr(uint32(o.MaxEvpnRoutesCount.ValueInt64()))
		valueFound = true
	}
	if utils.Known(o.MaxFabricRoutesCount) {
		fabricSettings.MaxFabricRoutes = utils.ToPtr(uint32(o.MaxFabricRoutesCount.ValueInt64()))
		valueFound = true
	}
	if utils.Known(o.EvpnType5Routes) {
		fabricSettings.EvpnGenerateType5HostRoutes = utils.ToPtr(apstra.FeatureSwitchEnum{Value: o.EvpnType5Routes})
	}
	if utils.Known(o.JunosEvpnRoutingInstanceMode) {
		if o.JunosEvpnRoutingInstanceMode == types.String("enabled") {
			fabricSettings.JunosEvpnRoutingInstanceVlanAware = utils.ToPtr(apstra.FeatureSwitchEnumEnabled)
		}
		fabricSettings.JunosEvpnRoutingInstanceVlanAware = utils.ToPtr(o.JunosEvpnRoutingInstanceMode.ValueString())
	}
	if utils.Known(o.JunosEvpnMaxNexthopAndInterfaceNumber) {
		fabricSettings.JunosEvpnMaxNexthopAndInterfaceNumber = utils.ToPtr(o.JunosEvpnMaxNexthopAndInterfaceNumber)
	}
	if utils.Known(o.JunosGracefulRestart) {
		fabricSettings.JunosGracefulRestart = utils.ToPtr(o.JunosGracefulRestart)
	}

	if valueFound {
		return &fabricSettings
	}
	return nil
}

func (o *Blueprint) SetFabricSettings(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if version.MustConstraints(version.NewConstraint(">" + apiversions.Apstra412)).Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
		fabricSettings := o.fabricSettings()
		if fabricSettings == nil {
			return
		}
		bp.SetFabricSettings(ctx, &apstra.FabricSettings{
			//JunosEvpnDuplicateMacRecoveryTime:     nil,
			//MaxExternalRoutes:                     nil,
			//EsiMacMsb:                             nil, // formerly fabric_settings
			//JunosGracefulRestart:                  nil,
			//OptimiseSzFootprint:                   nil,
			//JunosEvpnRoutingInstanceVlanAware:     nil,
			//EvpnGenerateType5HostRoutes:           nil,
			//MaxFabricRoutes:                       nil,
			//MaxMlagRoutes:                         nil,
			//JunosExOverlayEcmp:                    nil,
			//DefaultSviL3Mtu:                       nil,
			//JunosEvpnMaxNexthopAndInterfaceNumber: nil,
			//FabricL3Mtu:                           nil, // formerly fabric_settings
			//Ipv6Enabled:                           nil, // formerly fabric_settings
			//OverlayControlProtocol:                nil,
			//ExternalRouterMtu:                     nil,
			//MaxEvpnRoutes:                         nil,
			//AntiAffinityPolicy:                    nil,
		})
	} else {

	}
}
