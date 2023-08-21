package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	"regexp"
	"terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/resources"
	"terraform-provider-apstra/apstra/utils"
)

type DatacenterVirtualNetwork struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	BlueprintId             types.String `tfsdk:"blueprint_id"`
	Type                    types.String `tfsdk:"type"`
	RoutingZoneId           types.String `tfsdk:"routing_zone_id"`
	Vni                     types.Int64  `tfsdk:"vni"`
	HadPriorVniConfig       types.Bool   `tfsdk:"had_prior_vni_config"`
	ReserveVlan             types.Bool   `tfsdk:"reserve_vlan"`
	Bindings                types.Map    `tfsdk:"bindings"`
	DhcpServiceEnabled      types.Bool   `tfsdk:"dhcp_service_enabled"`
	IPv4ConnectivityEnabled types.Bool   `tfsdk:"ipv4_connectivity_enabled"`
	IPv6ConnectivityEnabled types.Bool   `tfsdk:"ipv6_connectivity_enabled"`
	IPv4Subnet              types.String `tfsdk:"ipv4_subnet"`
	IPv6Subnet              types.String `tfsdk:"ipv6_subnet"`
	IPv4GatewayEnabled      types.Bool   `tfsdk:"ipv4_virtual_gateway_enabled"`
	IPv6GatewayEnabled      types.Bool   `tfsdk:"ipv6_virtual_gateway_enabled"`
	IPv4Gateway             types.String `tfsdk:"ipv4_virtual_gateway"`
	IPv6Gateway             types.String `tfsdk:"ipv6_virtual_gateway"`
}

func (o DatacenterVirtualNetwork) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
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
			MarkdownDescription: "Virtual Network Name",
			Optional:            true,
			Validators: []validator.String{stringvalidator.AtLeastOneOf(
				path.MatchRelative(),
				path.MatchRoot("filter").AtName("type"),
				path.MatchRoot("filter").AtName("routing_zone_id"),
				path.MatchRoot("filter").AtName("vni"),
				path.MatchRoot("filter").AtName("reserve_vlan"),
				path.MatchRoot("filter").AtName("dhcp_service_enabled"),
				path.MatchRoot("filter").AtName("ipv4_connectivity_enabled"),
				path.MatchRoot("filter").AtName("ipv6_connectivity_enabled"),
				path.MatchRoot("filter").AtName("ipv4_subnet"),
				path.MatchRoot("filter").AtName("ipv6_subnet"),
				path.MatchRoot("filter").AtName("ipv4_virtual_gateway_enabled"),
				path.MatchRoot("filter").AtName("ipv6_virtual_gateway_enabled"),
				path.MatchRoot("filter").AtName("ipv4_virtual_gateway"),
				path.MatchRoot("filter").AtName("ipv6_virtual_gateway"),
			)},
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Type",
			Optional:            true,
		},
		"routing_zone_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Routing Zone ID (required when `type == %s`", apstra.VnTypeVxlan),
			Optional:            true,
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "EVPN Virtual Network ID to be associated with this Virtual Network.",
			Optional:            true,
		},
		"had_prior_vni_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"reserve_vlan": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "For use only with `%s` type Virtual networks when all `bindings` " +
				"use the same VLAN ID. This option reserves the VLAN fabric-wide, even on switches to " +
				"which the Virtual Network has not yet been deployed.",
			Optional: true,
		},
		"bindings": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: map[string]dataSourceSchema.Attribute{},
			},
		},
		"dhcp_service_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables a DHCP relay agent.",
			Optional:            true,
		},
		"ipv4_connectivity_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv4 within the Virtual Network.",
			Optional:            true,
		},
		"ipv6_connectivity_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv6 within the Virtual Network.",
			Optional:            true,
		},
		"ipv4_subnet": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 subnet associated with the Virtual Network.",
			Optional:            true,
			Validators:          []validator.String{apstravalidator.ParseCidr(true, false)},
		},
		"ipv6_subnet": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 subnet associated with the Virtual Network. " +
				"Note that this attribute will not appear in the `graph_query` output " +
				"because IPv6 zero compression rules are problematic for mechanisms " +
				"which rely on string matching.",
			Optional:   true,
			Validators: []validator.String{apstravalidator.ParseCidr(false, true)},
		},
		"ipv4_virtual_gateway_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Controls and indicates whether the IPv4 gateway within the " +
				"Virtual Network is enabled.",
			Optional: true,
		},
		"ipv6_virtual_gateway_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Controls and indicates whether the IPv6 gateway within the " +
				"Virtual Network is enabled.",
			Optional: true,
		},
		"ipv4_virtual_gateway": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Specifies the IPv4 virtual gateway address within the " +
				"Virtual Network.",
			Optional:   true,
			Validators: []validator.String{apstravalidator.ParseIp(true, false)},
		},
		"ipv6_virtual_gateway": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Specifies the IPv6 virtual gateway address within the " +
				"Virtual Network. Note that this attribute will not appear in the " +
				"`graph_query` output because IPv6 zero compression rules are problematic " +
				"for mechanisms which rely on string matching.",
			Optional:   true,
			Validators: []validator.String{apstravalidator.ParseIp(false, true)},
		},
	}
}

func (o DatacenterVirtualNetwork) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Name",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 30),
				stringvalidator.RegexMatches(regexp.MustCompile(design.AlphaNumericRegexp), "valid characters are: "+design.AlphaNumericChars),
			},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Type",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(apstra.VnTypeVxlan.String()),
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{
				// specifically enumerated types - SDK supports additional
				// types which do not make sense in this context.
				stringvalidator.OneOf(apstra.VnTypeVlan.String(), apstra.VnTypeVxlan.String()),
			},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Routing Zone ID (required when `type == %s`", apstra.VnTypeVxlan),
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.StringRequiredWhenValueIs(
					path.MatchRelative().AtParent().AtName("type"),
					types.StringValue(apstra.VnTypeVxlan.String()),
				),
			},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"vni": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("EVPN Virtual Network ID to be associated with this Virtual "+
				"Network.  When omitted, Apstra chooses a VNI from the Resource Pool [allocated]"+
				"(../apstra_datacenter_resource_pool_allocation) to role `%s`.",
				utils.StringersToFriendlyString(apstra.ResourceGroupNameVxlanVnIds)),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(resources.VniMin, resources.VniMax),
				apstravalidator.Int64ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("type"),
					fmt.Sprintf("%q", apstra.VnTypeVlan.String()),
				),
			},
		},
		"had_prior_vni_config": resourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vni` has been removed from the configuration.",
			Computed:            true,
		},
		"reserve_vlan": resourceSchema.BoolAttribute{
			MarkdownDescription: "For use only with `%s` type Virtual networks when all `bindings` " +
				"use the same VLAN ID. This option reserves the VLAN fabric-wide, even on switches to which the" +
				" Virtual Network has not yet been deployed. The only accepted values is `true`.",
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(types.BoolValue(true),
					apstravalidator.ValueAtMustBeBool(
						path.MatchRelative().AtParent().AtName("type"),
						types.StringValue(apstra.VnTypeVxlan.String()),
						false,
					),
				),
			},
		},
		"bindings": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Bindings make a Virtual Network available on Leaf Switches and Access Switches. " +
				"At least one binding entry is required. The value is a map keyed by graph db node IDs of *either* " +
				"Leaf Switches (non-redundant Leaf Switches) or Leaf Switch redundancy groups (redundant Leaf " +
				"Switches). Practitioners are encouraged to consider using the " +
				"[`_datacenter_virtual_network_binding_constructor`]" +
				"(../data-sources/apstra_datacenter_virtual_network_binding_constructor) data source to populate " +
				"this map.",
			Required: true,
			Validators: []validator.Map{
				mapvalidator.SizeAtLeast(1),
				apstravalidator.WhenValueAtMustBeMap(
					path.MatchRelative().AtParent().AtName("type"),
					types.StringValue(apstra.VnTypeVlan.String()),
					mapvalidator.SizeAtMost(1),
				),
			},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: VnBinding{}.ResourceAttributes(),
			},
		},
		"dhcp_service_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables a DHCP relay agent.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(types.BoolValue(true),
					apstravalidator.AlsoRequiresNOf(1,
						path.MatchRelative().AtParent().AtName("ipv4_connectivity_enabled"),
						path.MatchRelative().AtParent().AtName("ipv6_connectivity_enabled"),
					),
				),
			},
		},
		"ipv4_connectivity_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv4 within the Virtual Network. Default: true",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
		},
		"ipv6_connectivity_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv6 within the Virtual Network. Default: false",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"ipv4_subnet": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("IPv4 subnet associated with the "+
				"Virtual Network. When not specified, a prefix from within the IPv4 "+
				"Resource Pool assigned to the `%s` role will be automatically a"+
				"ssigned by Apstra.", apstra.ResourceGroupNameVirtualNetworkSviIpv4),
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				apstravalidator.ParseCidr(true, false),
				apstravalidator.WhenValueSetString(
					apstravalidator.ValueAtMustBeString(
						path.MatchRelative().AtParent().AtName("ipv4_connectivity_enabled"),
						types.BoolValue(true), false,
					),
				),
			},
		},
		"ipv6_subnet": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("IPv6 subnet associated with the "+
				"Virtual Network. When not specified, a prefix from within the IPv6 "+
				"Resource Pool assigned to the `%s` role will be automatically a"+
				"ssigned by Apstra.", apstra.ResourceGroupNameVirtualNetworkSviIpv6),
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				apstravalidator.ParseCidr(false, true),
				apstravalidator.WhenValueSetString(
					apstravalidator.ValueAtMustBeString(
						path.MatchRelative().AtParent().AtName("ipv6_connectivity_enabled"),
						types.BoolValue(true), false,
					),
				),
			},
		},
		"ipv4_virtual_gateway_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Controls and indicates whether the IPv4 gateway within the " +
				"Virtual Network is enabled. Requires `ipv4_connectivity_enabled` to be `true`",
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(
					types.BoolValue(true),
					apstravalidator.ValueAtMustBeBool(
						path.MatchRelative().AtParent().AtName("ipv4_connectivity_enabled"),
						types.BoolValue(true),
						false,
					),
				),
			},
		},
		"ipv6_virtual_gateway_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Controls and indicates whether the IPv6 gateway within the " +
				"Virtual Network is enabled. Requires `ipv6_connectivity_enabled` to be `true`",
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(
					types.BoolValue(true),
					apstravalidator.ValueAtMustBeBool(
						path.MatchRelative().AtParent().AtName("ipv6_connectivity_enabled"),
						types.BoolValue(true),
						false,
					),
				),
			},
		},
		"ipv4_virtual_gateway": resourceSchema.StringAttribute{
			MarkdownDescription: "Specifies the IPv4 virtual gateway address within the " +
				"Virtual Network. The configured value must be a valid IPv4 host address " +
				"configured value within range specified by `ipv4_subnet`",
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				apstravalidator.ParseIp(true, false),
				apstravalidator.FallsWithinCidr(
					path.MatchRelative().AtParent().AtName("ipv4_subnet"),
					false, false),
			},
		},
		"ipv6_virtual_gateway": resourceSchema.StringAttribute{
			MarkdownDescription: "Specifies the IPv6 virtual gateway address within the " +
				"Virtual Network. The configured value must be a valid IPv6 host address " +
				"configured value within range specified by `ipv6_subnet`",
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				apstravalidator.ParseIp(false, true),
				apstravalidator.FallsWithinCidr(
					path.MatchRelative().AtParent().AtName("ipv6_subnet"),
					true, true),
			},
		},
	}
}

func (o *DatacenterVirtualNetwork) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.VirtualNetworkData {
	var vnType apstra.VnType
	err := vnType.FromString(o.Type.ValueString())
	if err != nil {
		diags.Append(
			validatordiag.BugInProviderDiagnostic(
				fmt.Sprintf("error parsing virtual network type %q - %s", o.Type.String(), err.Error())))
		return nil
	}

	b := make(map[string]VnBinding)
	diags.Append(o.Bindings.ElementsAs(ctx, &b, false)...)
	if diags.HasError() {
		return nil
	}
	vnBindings := make([]apstra.VnBinding, len(b))
	var i int
	for leafId, binding := range b {
		vnBindings[i] = *binding.Request(ctx, leafId, diags)
		i++
	}
	if diags.HasError() {
		return nil
	}

	var vnId *apstra.VNI
	if utils.Known(o.Vni) {
		v := apstra.VNI(o.Vni.ValueInt64())
		vnId = &v
	}

	if o.Type.ValueString() == apstra.VnTypeVlan.String() {
		// Exactly one binding is required when type==vlan.
		// Apstra requires vlan == vni when creating a "vlan" type VN.
		if vnBindings[0].VlanId != nil {
			v := apstra.VNI(*vnBindings[0].VlanId)
			vnId = &v
		}
	}

	var reservedVlanId *apstra.Vlan
	if o.ReserveVlan.ValueBool() {
		reservedVlanId = vnBindings[0].VlanId
	}

	var ipv4Subnet, ipv6Subnet *net.IPNet
	if utils.Known(o.IPv4Subnet) {
		_, ipv4Subnet, err = net.ParseCIDR(o.IPv4Subnet.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing attribute ipv4_subnet value %q", o.IPv4Subnet.ValueString()), err.Error())
		}
	}
	if utils.Known(o.IPv6Subnet) {
		_, ipv6Subnet, err = net.ParseCIDR(o.IPv6Subnet.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing attribute ipv6_subnet value %q", o.IPv6Subnet.ValueString()), err.Error())
		}
	}

	var ipv4Gateway, ipv6Gateway net.IP
	if utils.Known(o.IPv4Gateway) {
		ipv4Gateway = net.ParseIP(o.IPv4Gateway.ValueString())
	}
	if utils.Known(o.IPv6Gateway) {
		ipv6Gateway = net.ParseIP(o.IPv6Gateway.ValueString())
	}

	return &apstra.VirtualNetworkData{
		DhcpService:               apstra.DhcpServiceEnabled(o.DhcpServiceEnabled.ValueBool()),
		Ipv4Enabled:               o.IPv4ConnectivityEnabled.ValueBool(),
		Ipv4Subnet:                ipv4Subnet,
		Ipv6Enabled:               o.IPv6ConnectivityEnabled.ValueBool(),
		Ipv6Subnet:                ipv6Subnet,
		Label:                     o.Name.ValueString(),
		ReservedVlanId:            reservedVlanId,
		RouteTarget:               "",
		RtPolicy:                  nil,
		SecurityZoneId:            apstra.ObjectId(o.RoutingZoneId.ValueString()),
		SviIps:                    nil,
		VirtualGatewayIpv4:        ipv4Gateway,
		VirtualGatewayIpv6:        ipv6Gateway,
		VirtualGatewayIpv4Enabled: o.IPv4GatewayEnabled.ValueBool(),
		VirtualGatewayIpv6Enabled: o.IPv6GatewayEnabled.ValueBool(),
		VnBindings:                vnBindings,
		VnId:                      vnId,
		VnType:                    vnType,
		VirtualMac:                nil,
	}
}

func (o *DatacenterVirtualNetwork) LoadApiData(ctx context.Context, in *apstra.VirtualNetworkData, diags *diag.Diagnostics) {
	bindings := make([]attr.Value, len(in.VnBindings))
	for i := range in.VnBindings {
		var binding VnBinding
		binding.LoadApiData(ctx, in.VnBindings[i], diags)
		if diags.HasError() {
			return
		}

		var d diag.Diagnostics
		bindings[i], d = types.ObjectValueFrom(ctx, VnBinding{}.attrTypes(), binding)
		diags.Append(d...)
	}
	if diags.HasError() {
		return
	}

	var virtualGatewayIpv4, virtualGatewayIpv6 string
	if len(in.VirtualGatewayIpv4.To4()) == net.IPv4len {
		virtualGatewayIpv4 = in.VirtualGatewayIpv4.String()
	}
	if len(in.VirtualGatewayIpv6) == net.IPv6len {
		virtualGatewayIpv6 = in.VirtualGatewayIpv6.String()
	}

	o.Name = types.StringValue(in.Label)
	o.Type = types.StringValue(in.VnType.String())
	o.RoutingZoneId = types.StringValue(in.SecurityZoneId.String())
	o.Bindings = newBindingMap(ctx, in.VnBindings, diags)
	o.Vni = utils.Int64ValueOrNull(ctx, in.VnId, diags)
	o.DhcpServiceEnabled = types.BoolValue(bool(in.DhcpService))
	o.IPv4ConnectivityEnabled = types.BoolValue(in.Ipv4Enabled)
	o.IPv6ConnectivityEnabled = types.BoolValue(in.Ipv6Enabled)
	o.ReserveVlan = types.BoolValue(in.ReservedVlanId != nil)
	if in.Ipv4Subnet == nil {
		o.IPv4Subnet = types.StringNull()
	} else {
		o.IPv4Subnet = types.StringValue(in.Ipv4Subnet.String())
	}
	if in.Ipv6Subnet == nil {
		o.IPv6Subnet = types.StringNull()
	} else {
		o.IPv6Subnet = types.StringValue(in.Ipv6Subnet.String())
	}
	o.IPv4GatewayEnabled = types.BoolValue(in.VirtualGatewayIpv4Enabled)
	o.IPv6GatewayEnabled = types.BoolValue(in.VirtualGatewayIpv6Enabled)
	o.IPv4Gateway = utils.StringValueOrNull(ctx, virtualGatewayIpv4, diags)
	o.IPv6Gateway = utils.StringValueOrNull(ctx, virtualGatewayIpv6, diags)
}

func (o *DatacenterVirtualNetwork) Query(vnResultName string) apstra.QEQuery {
	nodeAttributes := []apstra.QEEAttribute{
		apstra.NodeTypeVirtualNetwork.QEEAttribute(),
		{Key: "name", Value: apstra.QEStringVal(vnResultName)},
	}

	if !o.Name.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "label",
			Value: apstra.QEStringVal(o.Name.ValueString()),
		})
	}

	if !o.Type.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "vn_type",
			Value: apstra.QEStringVal(o.Type.ValueString()),
		})
	}

	if !o.Vni.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "vn_id",
			Value: apstra.QEIntVal(int(o.Vni.ValueInt64())),
		})
	}

	if !o.ReserveVlan.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "reserved_vlan_id",
			Value: apstra.QENone(!o.ReserveVlan.ValueBool()),
		})
	}

	if !o.IPv4ConnectivityEnabled.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "ipv4_enabled",
			Value: apstra.QEBoolVal(o.IPv4ConnectivityEnabled.ValueBool()),
		})
	}

	if !o.IPv6ConnectivityEnabled.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "ipv6_enabled",
			Value: apstra.QEBoolVal(o.IPv6ConnectivityEnabled.ValueBool()),
		})
	}

	if !o.IPv4Subnet.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "ipv4_subnet",
			Value: apstra.QEStringVal(o.IPv4Subnet.ValueString()),
		})
	}

	// not handling ipv6 subnet as a string match because of '::' expansion weirdness
	//if !o.IPv6Subnet.IsNull() { nope! }

	if !o.IPv4GatewayEnabled.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "virtual_gateway_ipv4_enabled",
			Value: apstra.QEBoolVal(o.IPv4GatewayEnabled.ValueBool()),
		})
	}

	if !o.IPv6GatewayEnabled.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "virtual_gateway_ipv6_enabled",
			Value: apstra.QEBoolVal(o.IPv6GatewayEnabled.ValueBool()),
		})
	}

	if !o.IPv4Gateway.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "virtual_gateway_ipv4",
			Value: apstra.QEStringVal(o.IPv4Gateway.ValueString()),
		})
	}

	// not handling ipv6 gateway as a string match because of '::' expansion weirdness
	//if !o.IPv6Gateway.IsNull() { nope! }

	// Begin the query with the VN node
	vnQuery := new(apstra.MatchQuery).Match(new(apstra.PathQuery).Node(nodeAttributes))

	if !o.RoutingZoneId.IsNull() {
		// extend the query with a routing zone match
		vnQuery.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetwork.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(vnResultName)},
			}).In([]apstra.QEEAttribute{apstra.RelationshipTypeMemberVNs.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSecurityZone.QEEAttribute(),
				{Key: "id", Value: apstra.QEStringVal(o.RoutingZoneId.ValueString())},
			}))
	}

	if !o.DhcpServiceEnabled.IsNull() {
		vnQuery.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetwork.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(vnResultName)},
			}).Out([]apstra.QEEAttribute{apstra.RelationshipTypeInstantiatedBy.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetworkInstance.QEEAttribute(),
				{Key: "dhcp_enabled", Value: apstra.QEBoolVal(o.DhcpServiceEnabled.ValueBool())},
			}))
	}

	return vnQuery
}

func (o *DatacenterVirtualNetwork) Ipv6Subnet(_ context.Context, path path.Path, diags *diag.Diagnostics) *net.IPNet {
	if o.IPv6Subnet.IsNull() {
		return nil
	}

	_, result, err := net.ParseCIDR(o.IPv6Subnet.ValueString())
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed to parse 'ipv6_subnet' value %s", o.IPv6Subnet), err.Error())
		return nil
	}

	return result
}

func (o *DatacenterVirtualNetwork) Ipv6Gateway(_ context.Context, _ path.Path, _ *diag.Diagnostics) net.IP {
	if o.IPv6Gateway.IsNull() {
		return nil
	}

	return net.ParseIP(o.IPv6Gateway.ValueString())
}
