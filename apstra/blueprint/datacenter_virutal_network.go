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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ReserveVlan             types.Bool   `tfsdk:"reserve_vlan"`
	Bindings                types.Map    `tfsdk:"bindings"`
	DhcpServiceEnabled      types.Bool   `tfsdk:"dhcp_service_enabled"`
	IPv4ConnectivityEnabled types.Bool   `tfsdk:"ipv4_connectivity_enabled"`
	IPv6ConnectivityEnabled types.Bool   `tfsdk:"ipv6_connectivity_enabled"`
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
			MarkdownDescription: fmt.Sprintf("EVPN Virtual Network ID to be associatd with this Virtual "+
				"Network.  When omitted, Apstra chooses a VNI from the Resource Pool [allocated]"+
				"(../apstra_datacenter_resource_pool_allocation) to role `%s`.", apstra.ResourceGroupNameEvpnL3Vni),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(resources.VniMin-1, resources.VniMax+1),
				apstravalidator.Int64ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("type"),
					fmt.Sprintf("%q", apstra.VnTypeVlan.String()),
				),
			},
		},
		"reserve_vlan": resourceSchema.BoolAttribute{
			MarkdownDescription: "For use only with `%s` type Virtual networks when all `bindings` " +
				"use the same VLAN ID. This option reserves the VLAN fabric-wide, even on switches to which the" +
				" Virtual Network has not yet been deployed. The only accepted values is `true`.",
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(types.BoolValue(true),
					apstravalidator.ReserveVlanOK(),
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
				apstravalidator.ExactlyOneBindingWhenVnTypeVlan(),
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
			MarkdownDescription: "Enables IPv4 within the Virtual Network",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"ipv6_connectivity_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv6 within the Virtual Network",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
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
		v := apstra.VNI(*vnBindings[0].VlanId)
		vnId = &v
	}

	var reservedVlanId *apstra.Vlan
	if o.ReserveVlan.ValueBool() {
		reservedVlanId = vnBindings[0].VlanId
	}

	return &apstra.VirtualNetworkData{
		DhcpService:               apstra.DhcpServiceEnabled(o.DhcpServiceEnabled.ValueBool()),
		Ipv4Enabled:               o.IPv4ConnectivityEnabled.ValueBool(),
		Ipv4Subnet:                nil,
		Ipv6Enabled:               o.IPv6ConnectivityEnabled.ValueBool(),
		Ipv6Subnet:                nil,
		Label:                     o.Name.ValueString(),
		ReservedVlanId:            reservedVlanId,
		RouteTarget:               "",
		RtPolicy:                  nil,
		SecurityZoneId:            apstra.ObjectId(o.RoutingZoneId.ValueString()),
		SviIps:                    nil,
		VirtualGatewayIpv4:        nil,
		VirtualGatewayIpv6:        nil,
		VirtualGatewayIpv4Enabled: false,
		VirtualGatewayIpv6Enabled: false,
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

	o.Name = types.StringValue(in.Label)
	o.Type = types.StringValue(in.VnType.String())
	o.RoutingZoneId = types.StringValue(in.SecurityZoneId.String())
	o.Bindings = newBindingMap(ctx, in.VnBindings, diags)
	o.Vni = utils.Int64ValueOrNull(ctx, in.VnId, diags)
	o.DhcpServiceEnabled = types.BoolValue(bool(in.DhcpService))
	o.IPv4ConnectivityEnabled = types.BoolValue(in.Ipv4Enabled)
	o.IPv6ConnectivityEnabled = types.BoolValue(in.Ipv6Enabled)
	o.ReserveVlan = types.BoolValue(in.ReservedVlanId != nil)
}
