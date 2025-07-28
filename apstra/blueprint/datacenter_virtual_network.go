package blueprint

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
)

type DatacenterVirtualNetwork struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	BlueprintId             types.String `tfsdk:"blueprint_id"`
	Type                    types.String `tfsdk:"type"`
	RoutingZoneId           types.String `tfsdk:"routing_zone_id"`
	Vni                     types.Int64  `tfsdk:"vni"`
	HadPriorVniConfig       types.Bool   `tfsdk:"had_prior_vni_config"`
	ReserveVlan             types.Bool   `tfsdk:"reserve_vlan"`
	ReservedVlanId          types.Int64  `tfsdk:"reserved_vlan_id"`
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
	L3Mtu                   types.Int64  `tfsdk:"l3_mtu"`
	ImportRouteTargets      types.Set    `tfsdk:"import_route_targets"`
	ExportRouteTargets      types.Set    `tfsdk:"export_route_targets"`
	SviIps                  types.Set    `tfsdk:"svi_ips"`
	Tags                    types.Set    `tfsdk:"tags"`
}

func (o DatacenterVirtualNetwork) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The id of the Virtual Network",
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
			MarkdownDescription: "The blueprint ID where the Virtual Network is present.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Name",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Description",
			Computed:            true,
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Type",
			Computed:            true,
		},
		"routing_zone_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Routing Zone ID (only applies when `type == %s`", enum.VnTypeVxlan),
			Computed:            true,
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "EVPN Virtual Network ID to be associated with this Virtual Network.",
			Computed:            true,
		},
		"had_prior_vni_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Not applicable in data source context. Ignore.",
			Computed:            true,
		},
		"reserve_vlan": dataSourceSchema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("For use only with `%s` type Virtual networks when all `bindings` "+
				"use the same VLAN ID. This option reserves the VLAN fabric-wide, even on switches to "+
				"which the Virtual Network has not yet been deployed.", enum.VnTypeVxlan),
			Computed: true,
		},
		"reserved_vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Reserved VLAN ID, if any.",
			Computed:            true,
		},
		"bindings": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Details availability of the virtual network on leaf and access switches",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: VnBinding{}.DataSourceAttributes(),
			},
		},
		"dhcp_service_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables a DHCP relay agent.",
			Computed:            true,
		},
		"ipv4_connectivity_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv4 within the Virtual Network.",
			Computed:            true,
		},
		"ipv6_connectivity_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv6 within the Virtual Network.",
			Computed:            true,
		},
		"ipv4_subnet": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 subnet associated with the Virtual Network.",
			Computed:            true,
		},
		"ipv6_subnet": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 subnet associated with the Virtual Network. " +
				"Note that this attribute will not appear in the `graph_query` output " +
				"because IPv6 zero compression rules are problematic for mechanisms " +
				"which rely on string matching.",
			Computed: true,
		},
		"ipv4_virtual_gateway_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Controls and indicates whether the IPv4 gateway within the " +
				"Virtual Network is enabled.",
			Computed: true,
		},
		"ipv6_virtual_gateway_enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Controls and indicates whether the IPv6 gateway within the " +
				"Virtual Network is enabled.",
			Computed: true,
		},
		"ipv4_virtual_gateway": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Specifies the IPv4 virtual gateway address within the " +
				"Virtual Network.",
			Computed: true,
		},
		"ipv6_virtual_gateway": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Specifies the IPv6 virtual gateway address within the " +
				"Virtual Network. Note that this attribute will not appear in the " +
				"`graph_query` output because IPv6 zero compression rules are problematic " +
				"for mechanisms which rely on string matching.",
			Computed: true,
		},
		"l3_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "L3 MTU used by the L3 switch interfaces participating in the Virtual Network. " +
				"Requires Apstra 4.2 or later.",
			Computed: true,
		},
		"import_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Import RTs for this Virtual Network.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"export_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Export RTs for this Virtual Network.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Tags for this Virtual Network.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
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
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Description",
			Optional:            true,
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Type",
			Optional:            true,
		},
		"routing_zone_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Routing Zone ID (required when `type == %s`)", enum.VnTypeVxlan),
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
			MarkdownDescription: "Selects only virtual networks with the *Reserve across blueprint* box checked.",
			Optional:            true,
		},
		"reserved_vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Selects only virtual networks with the *Reserve across blueprint* box checked and this value selected.",
			Optional:            true,
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
		"l3_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "L3 MTU used by the L3 switch interfaces participating in the Virtual Network. " +
				"Requires Apstra 4.2 or later.",
			Optional: true,
		},
		"import_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "This is a set of *required* import RTs, not an exact-match list.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators:          []validator.Set{setvalidator.ValueStringsAre(apstravalidator.ParseRT())},
		},
		"export_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "This is a set of *required* export RTs, not an exact-match list.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators:          []validator.Set{setvalidator.ValueStringsAre(apstravalidator.ParseRT())},
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags. All tags supplied here are used to match the Virtual Network, " +
				"but a matching Virtual Network may have additional tags not enumerated in this set.",
			Optional:    true,
			ElementType: types.StringType,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
	}
}

func (o DatacenterVirtualNetwork) ResourceAttributes() map[string]resourceSchema.Attribute {
	attrs := map[string]resourceSchema.Attribute{
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
				stringvalidator.RegexMatches(apstraregexp.AlphaNumW2HLConstraint, apstraregexp.AlphaNumW2HLConstraintMsg),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Description",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 222),
				stringvalidator.RegexMatches(regexp.MustCompile(`^[^"<>\\?]+$`), `must not contain the following characters: ", <, >, \, ?`),
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
			Default:             stringdefault.StaticString(enum.VnTypeVxlan.String()),
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{
				// specifically enumerated types - SDK supports additional
				// types which do not make sense in this context.
				stringvalidator.OneOf(enum.VnTypeVlan.String(), enum.VnTypeVxlan.String()),
			},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Routing Zone ID (required when `type == %s`", enum.VnTypeVxlan),
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.RequiredWhenValueIs(
					path.MatchRelative().AtParent().AtName("type"),
					types.StringValue(enum.VnTypeVxlan.String()),
				),
				apstravalidator.RequiredWhenValueNull(
					path.MatchRelative().AtParent().AtName("type"),
				),
			},
		},
		"vni": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("EVPN Virtual Network ID to be associated with this Virtual "+
				"Network.  When omitted, Apstra chooses a VNI from the Resource Pool [allocated]"+
				"(../resources/datacenter_resource_pool_allocation) to role `%s`.",
				utils.StringersToFriendlyString(apstra.ResourceGroupNameVxlanVnIds)),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.VniMin, constants.VniMax),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("type"),
					types.StringValue(enum.VnTypeVlan.String()),
				),
			},
		},
		"had_prior_vni_config": resourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vni` has been removed from the configuration.",
			Computed:            true,
		},
		"reserve_vlan": resourceSchema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("For use only with `%s` type Virtual networks when all "+
				"`bindings` use the same VLAN ID. This option reserves the VLAN fabric-wide, even on "+
				"switches to which the Virtual Network has not yet been deployed.", enum.VnTypeVxlan.String()),
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(
					types.BoolValue(true),
					apstravalidator.ForbiddenWhenValueIs(
						path.MatchRelative().AtParent().AtName("type"),
						types.StringValue(enum.VnTypeVlan.String()),
					),
				),
				apstravalidator.AlsoRequiresNOf(1,
					path.MatchRoot("bindings"),
					path.MatchRoot("reserved_vlan_id"),
				),
			},
		},
		"reserved_vlan_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "Used to specify the reserved VLAN ID without specifying any *bindings*.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("reserve_vlan"), types.BoolNull()),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("reserve_vlan"), types.BoolValue(false)),
				int64validator.ConflictsWith(path.MatchRoot("bindings")),
				int64validator.Between(design.VlanMin, design.VlanMax),
			},
		},
		"bindings": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Bindings make a Virtual Network available on Leaf Switches and Access Switches. " +
				"At least one binding entry is required with Apstra 4.x. With Apstra 5.x, a Virtual Network with " +
				"no bindings can be created by omitting (or setting `null`) this attribute. The value is a map " +
				"keyed by graph db node IDs of *either* Leaf Switches (non-redundant Leaf Switches) or Leaf Switch " +
				"redundancy groups (redundant Leaf Switches). Practitioners are encouraged to consider using the " +
				"[`apstra_datacenter_virtual_network_binding_constructor`]" +
				"(../data-sources/datacenter_virtual_network_binding_constructor) data source to populate " +
				"this map.",
			Optional: true,
			Validators: []validator.Map{
				mapvalidator.SizeAtLeast(1),
				apstravalidator.WhenValueAtMustBeMap(
					path.MatchRelative().AtParent().AtName("type"),
					types.StringValue(enum.VnTypeVlan.String()),
					mapvalidator.SizeAtMost(1),
				),
			},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: VnBinding{}.ResourceAttributes(),
			},
		},
		"dhcp_service_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables a DHCP relay agent. Note that configuring this feature without configuring " +
				"any `bindings` may lead to state churn because a VN with no bindings does not retain the " +
				"`dhcp_service_enabled` state.",
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
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
		"l3_mtu": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("L3 MTU used by the L3 switch interfaces participating in the"+
				" Virtual Network. Must be an even number between %d and %d. Requires Apstra %s or later.",
				constants.L3MtuMin, constants.L3MtuMax, apiversions.Apstra420),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"import_route_targets": resourceSchema.SetAttribute{
			MarkdownDescription: "Import RTs for this Virtual Network.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseRT()),
			},
		},
		"export_route_targets": resourceSchema.SetAttribute{
			MarkdownDescription: "Export RTs for this Virtual Network.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseRT()),
			},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of tags for this Virtual Network",
			Optional:            true,
			ElementType:         types.StringType,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"svi_ips": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of SVI IP addresses for this Virtual Network",
			Optional:            true,
			ElementType:         types.StringType,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *DatacenterVirtualNetwork) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.VirtualNetworkData {
	var vnType enum.VnType
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
	if utils.HasValue(o.Vni) {
		v := apstra.VNI(o.Vni.ValueInt64())
		vnId = &v
	}

	if o.Type.ValueString() == enum.VnTypeVlan.String() {
		// Maximum of one binding is required when type==vlan.
		// Apstra requires vlan == vni when creating a "vlan" type VN.
		// VNI attribute is forbidden when type == VLAN
		if len(vnBindings) > 0 && vnBindings[0].VlanId != nil {
			v := apstra.VNI(*vnBindings[0].VlanId)
			vnId = &v
		}
	}

	var reservedVlanId *apstra.Vlan
	if o.ReserveVlan.ValueBool() {
		if utils.HasValue(o.ReservedVlanId) {
			reservedVlanId = utils.ToPtr(apstra.Vlan(o.ReservedVlanId.ValueInt64()))
		} else {
			reservedVlanId = vnBindings[0].VlanId
		}
	}

	var ipv4Subnet, ipv6Subnet *net.IPNet
	if utils.HasValue(o.IPv4Subnet) {
		_, ipv4Subnet, err = net.ParseCIDR(o.IPv4Subnet.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing attribute ipv4_subnet value %q", o.IPv4Subnet.ValueString()), err.Error())
		}
	}
	if utils.HasValue(o.IPv6Subnet) {
		_, ipv6Subnet, err = net.ParseCIDR(o.IPv6Subnet.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing attribute ipv6_subnet value %q", o.IPv6Subnet.ValueString()), err.Error())
		}
	}

	var ipv4Gateway, ipv6Gateway net.IP
	if utils.HasValue(o.IPv4Gateway) {
		ipv4Gateway = net.ParseIP(o.IPv4Gateway.ValueString())
	}
	if utils.HasValue(o.IPv6Gateway) {
		ipv6Gateway = net.ParseIP(o.IPv6Gateway.ValueString())
	}

	var l3Mtu *int
	if utils.HasValue(o.L3Mtu) {
		i := int(o.L3Mtu.ValueInt64())
		l3Mtu = &i
	}

	var rtPolicy *apstra.RtPolicy
	if !o.ImportRouteTargets.IsNull() || !o.ExportRouteTargets.IsNull() {
		rtPolicy = new(apstra.RtPolicy)
		if !o.ImportRouteTargets.IsNull() {
			diags.Append(o.ImportRouteTargets.ElementsAs(ctx, &rtPolicy.ImportRTs, false)...)
		}
		if !o.ExportRouteTargets.IsNull() {
			diags.Append(o.ExportRouteTargets.ElementsAs(ctx, &rtPolicy.ExportRTs, false)...)
		}
	}

	return &apstra.VirtualNetworkData{
		Description:               o.Description.ValueString(),
		DhcpService:               apstra.DhcpServiceEnabled(o.DhcpServiceEnabled.ValueBool()),
		Ipv4Enabled:               o.IPv4ConnectivityEnabled.ValueBool(),
		Ipv4Subnet:                ipv4Subnet,
		Ipv6Enabled:               o.IPv6ConnectivityEnabled.ValueBool(),
		Ipv6Subnet:                ipv6Subnet,
		L3Mtu:                     l3Mtu,
		Label:                     o.Name.ValueString(),
		ReservedVlanId:            reservedVlanId,
		RouteTarget:               "",
		RtPolicy:                  rtPolicy,
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
	var virtualGatewayIpv4, virtualGatewayIpv6 string
	if len(in.VirtualGatewayIpv4.To4()) == net.IPv4len {
		virtualGatewayIpv4 = in.VirtualGatewayIpv4.String()
	}
	if len(in.VirtualGatewayIpv6) == net.IPv6len {
		virtualGatewayIpv6 = in.VirtualGatewayIpv6.String()
	}

	o.Name = types.StringValue(in.Label)
	o.Description = utils.StringValueOrNull(ctx, in.Description, diags)
	o.Type = types.StringValue(in.VnType.String())
	o.RoutingZoneId = types.StringValue(in.SecurityZoneId.String())
	o.Bindings = newBindingMap(ctx, in.VnBindings, diags)
	o.Vni = utils.Int64ValueOrNull(ctx, in.VnId, diags)
	o.DhcpServiceEnabled = types.BoolValue(bool(in.DhcpService))
	o.IPv4ConnectivityEnabled = types.BoolValue(in.Ipv4Enabled)
	o.IPv6ConnectivityEnabled = types.BoolValue(in.Ipv6Enabled)
	o.ReserveVlan = types.BoolValue(in.ReservedVlanId != nil)
	if in.ReservedVlanId == nil {
		o.ReservedVlanId = types.Int64Null()
	} else {
		o.ReservedVlanId = types.Int64Value(int64(*in.ReservedVlanId))
	}
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
	o.L3Mtu = utils.Int64ValueOrNull(ctx, in.L3Mtu, diags)

	if in.RtPolicy == nil {
		o.ImportRouteTargets = types.SetNull(types.StringType)
		o.ExportRouteTargets = types.SetNull(types.StringType)
	} else {
		o.ImportRouteTargets = utils.SetValueOrNull(ctx, types.StringType, in.RtPolicy.ImportRTs, diags)
		o.ExportRouteTargets = utils.SetValueOrNull(ctx, types.StringType, in.RtPolicy.ExportRTs, diags)
	}

	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags)
}

func (o *DatacenterVirtualNetwork) Query(resultName string) apstra.QEQuery {
	nodeAttributes := []apstra.QEEAttribute{
		apstra.NodeTypeVirtualNetwork.QEEAttribute(),
		{Key: "name", Value: apstra.QEStringVal(resultName)},
	}

	if !o.Name.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "label",
			Value: apstra.QEStringVal(o.Name.ValueString()),
		})
	}

	if !o.Description.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "description",
			Value: apstra.QEStringVal(o.Description.ValueString()),
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
			Value: apstra.QEStringVal(strconv.Itoa(int(o.Vni.ValueInt64()))),
		})
	}

	if !o.ReserveVlan.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "reserved_vlan_id",
			Value: apstra.QENone(!o.ReserveVlan.ValueBool()),
		})
	}

	if !o.ReservedVlanId.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "reserved_vlan_id",
			Value: apstra.QEIntVal(o.ReservedVlanId.ValueInt64()),
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
	// if !o.IPv6Subnet.IsNull() { nope! }

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

	if !o.L3Mtu.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{
			Key:   "l3_mtu",
			Value: apstra.QEIntVal(o.L3Mtu.ValueInt64()),
		})
	}

	// not handling ipv6 gateway as a string match because of '::' expansion weirdness
	// if !o.IPv6Gateway.IsNull() { nope! }

	// Begin the query with the VN node
	vnQuery := new(apstra.MatchQuery).Match(new(apstra.PathQuery).Node(nodeAttributes))

	if !o.RoutingZoneId.IsNull() {
		// extend the query with a routing zone match
		vnQuery.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetwork.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(resultName)},
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
				{Key: "name", Value: apstra.QEStringVal(resultName)},
			}).Out([]apstra.QEEAttribute{apstra.RelationshipTypeInstantiatedBy.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetworkInstance.QEEAttribute(),
				{Key: "dhcp_enabled", Value: apstra.QEBoolVal(o.DhcpServiceEnabled.ValueBool())},
			}))
	}

	// Add RT import/export matchers for the route_target_policy node as needed
	if !o.ImportRouteTargets.IsNull() || !o.ExportRouteTargets.IsNull() {
		nodeName := "n_route_target_policy"
		rtQuery := new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetwork.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(resultName)},
			}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeRouteTargetPolicy.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeRouteTargetPolicy.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(nodeName)},
			})

		for _, attrVal := range o.ImportRouteTargets.Elements() {
			iRT := attrVal.(types.String).ValueString()
			where := fmt.Sprintf("lambda %s: '%s' in (%s.import_RTs or [])", nodeName, iRT, nodeName)
			rtQuery.Where(where)
		}

		for _, attrVal := range o.ExportRouteTargets.Elements() {
			eRT := attrVal.(types.String).ValueString()
			where := fmt.Sprintf("lambda %s: '%s' in (%s.export_RTs or [])", nodeName, eRT, nodeName)
			rtQuery.Where(where)
		}

		vnQuery.Match(rtQuery)
	}

	for _, tag := range o.Tags.Elements() {
		tagQuery := new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetwork.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(resultName)},
			}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeTag.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeTag.QEEAttribute(),
				{Key: "label", Value: apstra.QEStringVal(tag.(types.String).ValueString())},
			})

		vnQuery.Match(tagQuery)
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

// ValidateConfigBindingsReservation ensures that all bindings use the same VLAN
// ID when `reserve_vlan` is true.
func (o DatacenterVirtualNetwork) ValidateConfigBindingsReservation(ctx context.Context, diags *diag.Diagnostics) {
	// validation only possible when reserve_vlan is set "true"
	if !o.ReserveVlan.ValueBool() {
		return // skip 'false', 'unknown', 'null' values
	}

	// validation not possible when bindings are unknown
	if o.Bindings.IsUnknown() {
		return
	}

	// validation not possible when any individual binding is unknown
	for _, v := range o.Bindings.Elements() {
		if v.IsUnknown() {
			return
		}
	}

	// extract bindings as a map
	var bindings map[string]VnBinding
	diags.Append(o.Bindings.ElementsAs(ctx, &bindings, false)...)
	if diags.HasError() {
		return
	}

	// validate each binding
	invalidConfigDueToNullVlan := false
	reservedVlanIds := make(map[int64]struct{})
	for _, binding := range bindings {
		if binding.VlanId.IsUnknown() {
			continue // skip further validation of unknown vlans
		}
		if binding.VlanId.IsNull() {
			invalidConfigDueToNullVlan = true
			break
		}
		reservedVlanIds[binding.VlanId.ValueInt64()] = struct{}{}
	}

	// null vlan is invalid
	if invalidConfigDueToNullVlan {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			path.Root("bindings"),
			"'vlan_id' must be specified in each binding when 'reserve_vlan' is true"))
	}

	// we should have at most one VLAN ID across all bindings (zero when they're unknown)
	if len(reservedVlanIds) > 1 {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			path.Root("bindings"),
			"'vlan_id' attributes must match when 'reserve_vlan' is true"))
		return
	}
}

func (o DatacenterVirtualNetwork) VersionConstraints() compatibility.ConfigConstraints {
	var response compatibility.ConfigConstraints

	if !o.Bindings.IsUnknown() && len(o.Bindings.Elements()) == 0 {
		response.AddAttributeConstraints(
			compatibility.AttributeConstraint{
				Path:        path.Root("bindings"),
				Constraints: compatibility.VnEmptyBindingsOk,
			},
		)
	}

	if utils.HasValue(o.Description) {
		response.AddAttributeConstraints(
			compatibility.AttributeConstraint{
				Path:        path.Root("description"),
				Constraints: compatibility.VnDescriptionOk,
			},
		)
	}

	if utils.HasValue(o.Tags) {
		response.AddAttributeConstraints(
			compatibility.AttributeConstraint{
				Path:        path.Root("tags"),
				Constraints: compatibility.VnTagsOk,
			})
	}
	return response
}
