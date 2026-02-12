package blueprint

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterRoutingZone struct {
	Id                   types.String `tfsdk:"id"`
	BlueprintId          types.String `tfsdk:"blueprint_id"`
	Name                 types.String `tfsdk:"name"`
	VrfName              types.String `tfsdk:"vrf_name"`
	VlanId               types.Int64  `tfsdk:"vlan_id"`
	HadPriorVlanIdConfig types.Bool   `tfsdk:"had_prior_vlan_id_config"`
	Vni                  types.Int64  `tfsdk:"vni"`
	HadPriorVniConfig    types.Bool   `tfsdk:"had_prior_vni_config"`
	DhcpServers          types.Set    `tfsdk:"dhcp_servers"`
	RoutingPolicyId      types.String `tfsdk:"routing_policy_id"`
	ImportRouteTargets   types.Set    `tfsdk:"import_route_targets"`
	ExportRouteTargets   types.Set    `tfsdk:"export_route_targets"`
	JunosEvpnIrbMode     types.String `tfsdk:"junos_evpn_irb_mode"`
	IPAddressingType     types.String `tfsdk:"ip_addressing_type"`
	DisableIPv4          types.Bool   `tfsdk:"disable_ipv4"`
}

func (o DatacenterRoutingZone) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID. Required when `name` is omitted.",
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
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 17),
				stringvalidator.RegexMatches(apstraregexp.AlphaNumW2HLConstraint, apstraregexp.AlphaNumW2HLConstraintMsg),
			},
		},
		"vrf_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "VRF name.",
			Computed:            true,
		},
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Used for VLAN tagged Layer 3 links on external connections. " +
				"Leave this field blank to have it automatically assigned from a static pool in the " +
				"range of 2-4094), or enter a specific value.",
			Computed: true,
		},
		"had_prior_vlan_id_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vlan_id` has been removed from the " +
				"configuration in managed resource context, this attribute will always be `null` and should be " +
				"ignored in data source context.",
			Computed: true,
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "VxLAN VNI associated with the routing zone. Leave this field blank to have it " +
				"automatically assigned from an allocated resource pool, or enter a specific value.",
			Computed: true,
		},
		"had_prior_vni_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vni` has been removed from the " +
				"configuration in managed resource context, this attribute will always be `null` and should be " +
				"ignored in data source context.",
			Computed: true,
		},
		"dhcp_servers": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of DHCP server IPv4 or IPv6 addresses of DHCP servers.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"routing_policy_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Non-EVPN blueprints must use the default policy, so this field must be null. " +
				"Set this attribute in an EVPN blueprint to use a non-default policy.",
			Computed: true,
		},
		"import_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Used to import routes into the EVPN VRF.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"export_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Used to export routes from the EVPN VRF.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"junos_evpn_irb_mode": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Symmetric IRB Routing for EVPN on Junos devices makes use of an L3 VNI for " +
				"inter-subnet routing which is embedded into EVPN Type2-routes to support better scaling for " +
				"networks with large amounts of VLANs.",
			Computed: true,
		},
		"ip_addressing_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Defines if the according routing zone addresses resources with ipv4, ipv4+ipv6, or " +
				"ipv6-only. Errors are raised if resources are created within the routing zone and that resource violates " +
				"this addressing support value. Note that ipv4 is still permitted in an ipv6-only network, in which " +
				"case disable_ipv4 can be set to True to disallow ipv4 completely.",
			Computed: true,
		},
		"disable_ipv4": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Only valid for `ip_addressing_type=ipv6`. When this is set `true`, pure ipv6 routing " +
				"zones will not render IPv4 SAFIs and other ipv4-over-ipv6/RFC5549 related configuration will be removed. " +
				"User-defined IPv4 resources will not be permitted in the blueprint. An IPv4 loopback is still required " +
				"in order to derive BGP Router IDs and Route Distinguishers and it will not participate in routing.",
			Computed: true,
		},
	}
}

func (o DatacenterRoutingZone) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI.",
			Optional:            true,
		},
		"vrf_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "VRF name.",
			Optional:            true,
		},
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Used for VLAN tagged Layer 3 links on external connections.",
			Optional:            true,
		},
		"had_prior_vlan_id_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "VxLAN VNI associated with the routing zone.",
			Optional:            true,
		},
		"had_prior_vni_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"dhcp_servers": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of addresses of DHCP servers (IPv4 or IPv6) which must be configured " +
				"in the Routing Zone. This is a list of *required* servers, not an exact-match list.",
			ElementType: types.StringType,
			Optional:    true,
		},
		"routing_policy_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Non-EVPN blueprints must use the default policy, so this field must be null. " +
				"Set this attribute in an EVPN blueprint to use a non-default policy.",
			Optional: true,
		},
		"import_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "This is a set of *required* RTs, not an exact-match list.",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"export_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "This is a set of *required* RTs, not an exact-match list.",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"junos_evpn_irb_mode": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Symmetric IRB Routing for EVPN on Junos devices makes use of an L3 VNI for " +
				"inter-subnet routing which is embedded into EVPN Type2-routes to support better scaling for " +
				"networks with large amounts of VLANs.",
			Optional: true,
		},
		"ip_addressing_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Defines if the according routing zone addresses resources with "+
				"ipv4, ipv4+ipv6, or ipv6-only. Errors are raised if resources are created within the routing zone "+
				"and that resource violates this addressing support value. Note that ipv4 is still permitted in an "+
				"ipv6-only network, in which case disable_ipv4 can be set to True to disallow ipv4 completely.\n"+
				"Must be one of `['%s']`.\n"+
				"Requires Apstra version %s",
				strings.Join(enum.AddressingSchemes.Values(), "','"), compatibility.RoutingZoneAddressingTypeOK),
			Optional: true,
		},
		"disable_ipv4": dataSourceSchema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("Only valid for `ip_addressing_type=ipv6`. When this is set "+
				"`true`, pure ipv6 routing zones will not render IPv4 SAFIs and other ipv4-over-ipv6/RFC5549 related "+
				"configuration will be removed. User-defined IPv4 resources will not be permitted in the blueprint. "+
				"An IPv4 loopback is still required in order to derive BGP Router IDs and Route Distinguishers and it "+
				"will not participate in routing. Requires Apstra version %s", compatibility.RoutingZoneAddressingTypeOK),
			Optional: true,
		},
	}
}

func (o DatacenterRoutingZone) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(apstraregexp.AlphaNumW2HLConstraint, apstraregexp.AlphaNumW2HLConstraintMsg),
				stringvalidator.LengthBetween(1, 15),
			},
		},
		"vrf_name": resourceSchema.StringAttribute{
			MarkdownDescription: "VRF name. Copied from the `name` field on initial create.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"vlan_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "Used for VLAN tagged Layer 3 links on external connections. " +
				"Leave this field blank to have it automatically assigned from a static pool in the " +
				"range of 2-4094, or enter a specific value.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int64{int64validator.Between(design.VlanMin, design.VlanMax)},
		},
		"had_prior_vlan_id_config": resourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vlan_id` has been removed from the " +
				"configuration, this attribute can be ignored.",
			Computed: true,
		},
		"vni": resourceSchema.Int64Attribute{
			MarkdownDescription: "VxLAN VNI associated with the routing zone. Leave this field blank to have it " +
				"automatically assigned from an allocated resource pool, or enter a specific value.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int64{int64validator.Between(constants.VniMin, constants.VniMax)},
		},
		"had_prior_vni_config": resourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vni` has been removed from the " +
				"configuration, this attribute can be ignored.",
			Computed: true,
		},
		"dhcp_servers": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of DHCP server IPv4 or IPv6 addresses of DHCP servers.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseIp(false, false)),
			},
		},
		"routing_policy_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Non-EVPN blueprints must use the default policy, so this field must be null. " +
				"Set this attribute in an EVPN blueprint to use a non-default policy.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"import_route_targets": resourceSchema.SetAttribute{
			MarkdownDescription: "Used to import routes into the EVPN VRF.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseRT()),
			},
		},
		"export_route_targets": resourceSchema.SetAttribute{
			MarkdownDescription: "Used to export routes from the EVPN VRF.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseRT()),
			},
		},
		"junos_evpn_irb_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "Symmetric IRB Routing for EVPN on Junos devices makes use of an L3 VNI for " +
				"inter-subnet routing which is embedded into EVPN Type2-routes to support better scaling for " +
				"networks with large amounts of VLANs.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.OneOf(enum.JunosEVPNIRBModes.Values()...)},
			Default:    stringdefault.StaticString(enum.JunosEVPNIRBModeAsymmetric.String()),
		},
		"ip_addressing_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Defines if the according routing zone addresses resources with "+
				"ipv4, ipv4+ipv6, or ipv6-only. Errors are raised if resources are created within the routing zone "+
				"and that resource violates this addressing support value. Note that ipv4 is still permitted in an "+
				"ipv6-only network, in which case disable_ipv4 can be set to True to disallow ipv4 completely.\n"+
				"Must be one of `['%s']`.\n"+
				"Requires Apstra version %s",
				strings.Join(enum.AddressingSchemes.Values(), "','"), compatibility.RoutingZoneAddressingTypeOK),
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.OneOf(enum.AddressingSchemes.Values()...)},
		},
		"disable_ipv4": resourceSchema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("Only valid for `ip_addressing_type=ipv6`. When this is set "+
				"`true`, pure ipv6 routing zones will not render IPv4 SAFIs and other ipv4-over-ipv6/RFC5549 related "+
				"configuration will be removed. User-defined IPv4 resources will not be permitted in the blueprint. "+
				"An IPv4 loopback is still required in order to derive BGP Router IDs and Route Distinguishers and it "+
				"will not participate in routing. Requires Apstra version %s", compatibility.RoutingZoneAddressingTypeOK),
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				boolvalidator.AlsoRequires(path.MatchRoot("ip_addressing_type")),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("ip_addressing_type"), types.StringValue(enum.AddressingSchemeIPv4.String())),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("ip_addressing_type"), types.StringValue(enum.AddressingSchemeIPv46.String())),
			},
		},
	}
}

func (o *DatacenterRoutingZone) Request(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) apstra.SecurityZone {
	var result apstra.SecurityZone
	result.SetID(o.Id.ValueString()) // will set empty string if null or unknown -- this is fine

	if !o.VlanId.IsNull() && !o.VlanId.IsUnknown() {
		result.VLAN = pointer.To(apstra.VLAN(o.VlanId.ValueInt64()))
	}

	if !o.Vni.IsNull() && !o.Vni.IsUnknown() {
		result.VNI = pointer.To(int(o.Vni.ValueInt64()))
	}

	ocp, err := client.BlueprintOverlayControlProtocol(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		diags.AddError(fmt.Sprintf("API error querying for blueprint %q Overlay Control Protocol",
			o.BlueprintId.ValueString()), err.Error())
		return result
	}

	if ocp != apstra.OverlayControlProtocolEvpn {
		diags.AddAttributeError(
			path.Root("blueprint_id"),
			constants.ErrInvalidConfig,
			fmt.Sprintf("cannot create routing zone in blueprints with overlay control protocol %q", rosetta.StringersToFriendlyString(ocp)))
		return result
	}

	result.RTPolicy = new(apstra.RTPolicy)
	diags.Append(o.ImportRouteTargets.ElementsAs(ctx, &result.RTPolicy.ImportRTs, false)...)
	diags.Append(o.ExportRouteTargets.ElementsAs(ctx, &result.RTPolicy.ExportRTs, false)...)
	if diags.HasError() {
		return apstra.SecurityZone{}
	}

	result.Type = enum.SecurityZoneTypeEVPN
	result.VRFName = o.VrfName.ValueString()
	result.Label = o.Name.ValueString()
	result.RoutingPolicyID = o.RoutingPolicyId.ValueString()
	result.JunosEVPNIRBMode = enum.JunosEVPNIRBModes.Parse(o.JunosEvpnIrbMode.ValueString())

	if utils.HasValue(o.IPAddressingType) {
		var as enum.AddressingScheme
		err = as.FromString(o.IPAddressingType.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("ip_addressing_type"), constants.ErrInvalidConfig, err.Error())
		}

		result.AddressingSupport = &as
	}

	if utils.HasValue(o.DisableIPv4) {
		result.DisableIPv4 = o.DisableIPv4.ValueBoolPointer()
	}

	return result
}

func (o *DatacenterRoutingZone) DhcpServerRequest(_ context.Context, _ *diag.Diagnostics) []net.IP {
	dhcpServers := o.DhcpServers.Elements()
	request := make([]net.IP, len(dhcpServers))
	for i, dhcpServer := range dhcpServers {
		request[i] = net.ParseIP(dhcpServer.(types.String).ValueString())
	}
	return request
}

func (o *DatacenterRoutingZone) LoadApiData(ctx context.Context, sz apstra.SecurityZone, diags *diag.Diagnostics) {
	if !utils.HasValue(o.Name) { // required attribute
		o.Name = types.StringValue(sz.Label)
	}

	if !utils.HasValue(o.VrfName) { // computed attribute
		o.VrfName = types.StringValue(sz.VRFName)
	}

	if !utils.HasValue(o.VlanId) { // optional + computed attribute
		if sz.VLAN == nil {
			o.VlanId = types.Int64Null()
		} else {
			o.VlanId = types.Int64Value(int64(*sz.VLAN))
		}
	}

	if !utils.HasValue(o.RoutingPolicyId) { // optional + computed attribute
		if sz.RoutingPolicyID == "" {
			o.RoutingPolicyId = types.StringNull()
		} else {
			o.RoutingPolicyId = types.StringValue(sz.RoutingPolicyID)
		}
	}

	if !utils.HasValue(o.Vni) {
		if sz.VNI == nil {
			o.Vni = types.Int64Null()
		} else {
			o.Vni = types.Int64Value(int64(*sz.VNI))
		}
	}

	if !utils.HasValue(o.ImportRouteTargets) {
		if sz.RTPolicy == nil || sz.RTPolicy.ImportRTs == nil {
			o.ImportRouteTargets = types.SetNull(types.StringType)
		} else {
			o.ImportRouteTargets = value.SetOrNull(ctx, types.StringType, sz.RTPolicy.ImportRTs, diags)
		}
	}

	if !utils.HasValue(o.ExportRouteTargets) {
		if sz.RTPolicy == nil || sz.RTPolicy.ExportRTs == nil {
			o.ExportRouteTargets = types.SetNull(types.StringType)
		} else {
			o.ExportRouteTargets = value.SetOrNull(ctx, types.StringType, sz.RTPolicy.ExportRTs, diags)
		}
	}

	if !utils.HasValue(o.JunosEvpnIrbMode) {
		if sz.JunosEVPNIRBMode == nil {
			o.JunosEvpnIrbMode = types.StringNull()
		} else {
			o.JunosEvpnIrbMode = types.StringValue(sz.JunosEVPNIRBMode.Value)
		}
	}

	if !utils.HasValue(o.IPAddressingType) {
		if sz.AddressingSupport != nil {
			o.IPAddressingType = types.StringValue(sz.AddressingSupport.Value)
		} else {
			o.IPAddressingType = types.StringNull()
		}
	}

	if !utils.HasValue(o.DisableIPv4) {
		// only read "disable_ipv4" value when addressing scheme is ipv6
		if o.IPAddressingType.ValueString() == enum.AddressingSchemeIPv6.String() {
			o.DisableIPv4 = types.BoolPointerValue(sz.DisableIPv4)
		} else {
			o.DisableIPv4 = types.BoolNull()
		}
	}
}

func (o *DatacenterRoutingZone) LoadApiDhcpServers(ctx context.Context, IPs []net.IP, diags *diag.Diagnostics) {
	dhcpServers := make([]string, len(IPs))
	for i, ip := range IPs {
		dhcpServers[i] = ip.String()
	}
	o.DhcpServers = value.SetOrNull(ctx, types.StringType, dhcpServers, diags)
}

func (o *DatacenterRoutingZone) Query(szResultName string) *apstra.MatchQuery {
	matchQuery := new(apstra.MatchQuery)
	nodeQuery := new(apstra.PathQuery).Node(o.szNodeQueryAttributes(szResultName))
	matchQuery.Match(nodeQuery)

	if utils.HasValue(o.RoutingPolicyId) {
		q := new(apstra.PathQuery)
		q.Node([]apstra.QEEAttribute{
			{Key: "name", Value: apstra.QEStringVal(szResultName)},
		})
		q.Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.RelationshipTypePolicy.String())},
		})
		q.Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeRoutingPolicy.String())},
			{Key: "id", Value: apstra.QEStringVal(o.RoutingPolicyId.ValueString())},
		})
		matchQuery.Match(q)
	}

	if utils.HasValue(o.DhcpServers) {
		q := new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				{Key: "name", Value: apstra.QEStringVal(szResultName)},
			}).
			Out([]apstra.QEEAttribute{
				{Key: "type", Value: apstra.QEStringVal(apstra.RelationshipTypePolicy.String())},
			}).
			Node([]apstra.QEEAttribute{
				{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypePolicy.String())},
				{Key: "policy_type", Value: apstra.QEStringVal("dhcp_relay")},
				{Key: "name", Value: apstra.QEStringVal("dhcp_policy")},
				{Key: "dhcp_servers", Value: apstra.QENone(false)},
			})

		for _, dhcpServerVal := range o.DhcpServers.Elements() {
			q.Where(fmt.Sprintf(
				"lambda %s: '%s' in %s.dhcp_servers",
				"dhcp_policy",
				dhcpServerVal.(types.String).ValueString(),
				"dhcp_policy",
			))
		}

		matchQuery.Match(q)
	}

	if utils.HasValue(o.ImportRouteTargets) || utils.HasValue(o.ExportRouteTargets) {
		rtPolicyNodeAttrs := []apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeRouteTargetPolicy.String())},
			{Key: "name", Value: apstra.QEStringVal("rt_policy")},
		}

		if utils.HasValue(o.ImportRouteTargets) {
			rtPolicyNodeAttrs = append(rtPolicyNodeAttrs, apstra.QEEAttribute{
				Key:   "import_RTs",
				Value: apstra.QENone(false),
			})
		}

		if utils.HasValue(o.ExportRouteTargets) {
			rtPolicyNodeAttrs = append(rtPolicyNodeAttrs, apstra.QEEAttribute{
				Key:   "export_RTs",
				Value: apstra.QENone(false),
			})
		}

		q := new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				{Key: "name", Value: apstra.QEStringVal(szResultName)},
			}).
			Out([]apstra.QEEAttribute{
				{Key: "type", Value: apstra.QEStringVal(apstra.RelationshipTypeRouteTargetPolicy.String())},
			}).
			Node(rtPolicyNodeAttrs)

		for _, importRT := range o.ImportRouteTargets.Elements() {
			q.Where(fmt.Sprintf(
				"lambda %s: '%s' in %s.import_RTs",
				"rt_policy",
				importRT.(types.String).ValueString(),
				"rt_policy",
			))
		}

		for _, exportRT := range o.ExportRouteTargets.Elements() {
			q.Where(fmt.Sprintf(
				"lambda %s: '%s' in %s.export_RTs",
				"rt_policy",
				exportRT.(types.String).ValueString(),
				"rt_policy",
			))
		}

		matchQuery.Match(q)
	}

	return matchQuery
}

func (o *DatacenterRoutingZone) szNodeQueryAttributes(name string) []apstra.QEEAttribute {
	result := []apstra.QEEAttribute{
		{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeSecurityZone.String())},
	}

	if name != "" {
		result = append(result, apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal(name)})
	}

	if utils.HasValue(o.Name) {
		result = append(result, apstra.QEEAttribute{Key: "label", Value: apstra.QEStringVal(o.Name.ValueString())})
	}

	if utils.HasValue(o.VrfName) {
		result = append(result, apstra.QEEAttribute{Key: "vrf_name", Value: apstra.QEStringVal(o.VrfName.ValueString())})
	}

	if utils.HasValue(o.Vni) {
		result = append(result, apstra.QEEAttribute{Key: "vni_id", Value: apstra.QEIntVal(int(o.Vni.ValueInt64()))})
	}

	if utils.HasValue(o.VlanId) {
		result = append(result, apstra.QEEAttribute{Key: "vlan_id", Value: apstra.QEIntVal(int(o.VlanId.ValueInt64()))})
	}

	if utils.HasValue(o.JunosEvpnIrbMode) {
		result = append(result, apstra.QEEAttribute{Key: "junos_evpn_irb_mode", Value: apstra.QEStringVal(o.JunosEvpnIrbMode.ValueString())})
	}

	if utils.HasValue(o.IPAddressingType) {
		result = append(result, apstra.QEEAttribute{Key: "addressing_support", Value: apstra.QEStringVal(o.IPAddressingType.ValueString())})
	}

	if utils.HasValue(o.DisableIPv4) {
		result = append(result, apstra.QEEAttribute{Key: "disable_ipv4", Value: apstra.QEBoolVal(o.DisableIPv4.ValueBool())})
	}

	return result
}

func (o *DatacenterRoutingZone) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) error {
	if utils.HasValue(o.VlanId) &&
		utils.HasValue(o.Vni) &&
		utils.HasValue(o.DhcpServers) &&
		utils.HasValue(o.RoutingPolicyId) &&
		utils.HasValue(o.ImportRouteTargets) &&
		utils.HasValue(o.ExportRouteTargets) &&
		utils.HasValue(o.JunosEvpnIrbMode) &&
		utils.HasValue(o.IPAddressingType) &&
		utils.HasValue(o.DisableIPv4) {
		return nil // we are in Create() or Update() and have no need for an API call
	}

	o.BlueprintId = types.StringValue(bp.Id().String())

	sz, err := bp.GetSecurityZone(ctx, o.Id.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return err
		}
		diags.AddError("failed while reading routing zone from API", err.Error())
		return nil
	}

	o.LoadApiData(ctx, sz, diags)
	if diags.HasError() {
		return nil
	}

	if !utils.HasValue(o.DhcpServers) {
		dhcpServerIPs, err := bp.GetSecurityZoneDhcpServers(ctx, o.Id.ValueString())
		if err != nil {
			diags.AddError("failed retrieving security zone DCHP servers", err.Error())
			return nil
		}

		o.LoadApiDhcpServers(ctx, dhcpServerIPs, diags)
		if diags.HasError() {
			return nil
		}
	}

	return nil
}

func (o DatacenterRoutingZone) VersionConstraints(_ context.Context, _ *diag.Diagnostics) compatibility.ConfigConstraints {
	var response compatibility.ConfigConstraints

	if !o.IPAddressingType.IsUnknown() && !o.IPAddressingType.IsNull() {
		response.AddAttributeConstraints(compatibility.AttributeConstraint{
			Path:        path.Root("ip_addressing_type"),
			Constraints: compatibility.RoutingZoneAddressingTypeOK,
		})
	}

	if !o.DisableIPv4.IsUnknown() && !o.DisableIPv4.IsNull() {
		response.AddAttributeConstraints(compatibility.AttributeConstraint{
			Path:        path.Root("disable_ipv4"),
			Constraints: compatibility.RoutingZoneAddressingTypeOK,
		})
	}

	return response
}
