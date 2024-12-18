package primitives

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IpLink struct {
	Name                     types.String `tfsdk:"name"`
	RoutingZoneId            types.String `tfsdk:"routing_zone_id"`
	VlanId                   types.Int64  `tfsdk:"vlan_id"`
	L3Mtu                    types.Int64  `tfsdk:"l3_mtu"`
	Ipv4AddressingType       types.String `tfsdk:"ipv4_addressing_type"`
	Ipv6AddressingType       types.String `tfsdk:"ipv6_addressing_type"`
	BgpPeeringGenericSystems types.Set    `tfsdk:"bgp_peering_generic_systems"`
	BgpPeeringIpEndpoints    types.Set    `tfsdk:"bgp_peering_ip_endpoints"`
	DynamicBgpPeerings       types.Set    `tfsdk:"dynamic_bgp_peerings"`
	StaticRoutes             types.Set    `tfsdk:"static_routes"`
}

func (o IpLink) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                        types.StringType,
		"routing_zone_id":             types.StringType,
		"vlan_id":                     types.Int64Type,
		"l3_mtu":                      types.Int64Type,
		"ipv4_addressing_type":        types.StringType,
		"ipv6_addressing_type":        types.StringType,
		"bgp_peering_generic_systems": types.SetType{ElemType: types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()}},
		"bgp_peering_ip_endpoints":    types.SetType{ElemType: types.ObjectType{AttrTypes: BgpPeeringIpEndpoint{}.AttrTypes()}},
		"dynamic_bgp_peerings":        types.SetType{ElemType: types.ObjectType{AttrTypes: DynamicBgpPeering{}.AttrTypes()}},
		"static_routes":               types.SetType{ElemType: types.ObjectType{AttrTypes: StaticRoute{}.AttrTypes()}},
	}
}

func (o IpLink) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Label used by the web UI on the Primitive \"block\" in the Connectivity Template.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Node ID of the Routing Zone to which this IP Link should belong.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"vlan_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "802.1Q tag number to use for tagged IP Link. Omit for untagged IP Link.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(constants.VlanMinUsable, constants.VlanMaxUsable)}, // min vlan ID is 2
		},
		"l3_mtu": resourceSchema.Int64Attribute{
			// Frankly, I'm not clear what this text is trying to say. It's
			// taken verbatim from the tooltip in 99.2.0-cl-4.2.0-1
			MarkdownDescription: fmt.Sprintf("L3 MTU for sub-interfaces on leaf (spine/superspine) side and "+
				"generic side. Configuration is applicable only when Fabric MTU is enabled. Value must be even "+
				"number rom %d to %d, if not specified - Default IP Links to Generic Systems MTU from Virtual "+
				"Network Policy s used", constants.L3MtuMin, constants.L3MtuMax),
			Optional: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"ipv4_addressing_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of `%s`",
				strings.Join([]string{
					utils.StringersToFriendlyString(apstra.CtPrimitiveIPv4AddressingTypeNone),
					utils.StringersToFriendlyString(apstra.CtPrimitiveIPv4AddressingTypeNumbered),
				}, "`, `"),
			),
			Required: true,
			Validators: []validator.String{stringvalidator.OneOf(
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv4AddressingTypeNone),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv4AddressingTypeNumbered),
			)},
		},
		"ipv6_addressing_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of `%s`",
				strings.Join([]string{
					utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6AddressingTypeNone),
					utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6AddressingTypeLinkLocal),
					utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6AddressingTypeNumbered),
				}, "`, `"),
			),
			Required: true,
			Validators: []validator.String{stringvalidator.OneOf(
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6AddressingTypeNone),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6AddressingTypeLinkLocal),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6AddressingTypeNumbered),
			)},
		},
		"bgp_peering_generic_systems": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of BGP Peering (Generic System) primitives",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: BgpPeeringGenericSystem{}.ResourceAttributes(),
			},
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
			Optional:   true,
		},
		"bgp_peering_ip_endpoints": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of *BGP Peering (IP Endpoint)* Primitives in this Connectivity Template",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: BgpPeeringIpEndpoint{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"dynamic_bgp_peerings": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of *Dynamic BGP Peering* Primitives in this Connectivity Template",
			NestedObject:        resourceSchema.NestedAttributeObject{Attributes: DynamicBgpPeering{}.ResourceAttributes()},
			Optional:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"static_routes": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of network IPv4 or IPv6 destination prefixes reachable via this IP Link",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: StaticRoute{}.ResourceAttributes(),
			},
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
			Optional:   true,
		},
	}
}

func (o IpLink) attributes(_ context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink {
	var vlan *apstra.Vlan
	if !o.VlanId.IsNull() {
		vlan = utils.ToPtr(apstra.Vlan(o.VlanId.ValueInt64()))
	}

	var err error

	var ipv4AddressingType apstra.CtPrimitiveIPv4AddressingType
	err = utils.ApiStringerFromFriendlyString(&ipv4AddressingType, o.Ipv4AddressingType.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse ipv4_addressing_type value %s", o.Ipv4AddressingType), err.Error())
		return nil
	}

	var ipv6AddressingType apstra.CtPrimitiveIPv6AddressingType
	err = utils.ApiStringerFromFriendlyString(&ipv6AddressingType, o.Ipv6AddressingType.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse ipv6_addressing_type value %s", o.Ipv6AddressingType), err.Error())
		return nil
	}

	var l3Mtu *uint16
	if !o.L3Mtu.IsNull() {
		l3Mtu = utils.ToPtr(uint16(o.L3Mtu.ValueInt64()))
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink{
		Label:              o.Name.ValueString(), // todo is this necessary?
		SecurityZone:       (*apstra.ObjectId)(o.RoutingZoneId.ValueStringPointer()),
		Tagged:             !o.VlanId.IsNull(),
		Vlan:               vlan,
		IPv4AddressingType: ipv4AddressingType,
		IPv6AddressingType: ipv6AddressingType,
		L3Mtu:              l3Mtu,
	}
}

func (o IpLink) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	result := apstra.ConnectivityTemplatePrimitive{
		Label:      o.Name.ValueString(),
		Attributes: o.attributes(ctx, diags),
		// Subpolicies: // set below
	}

	result.Subpolicies = append(result.Subpolicies, BgpPeeringGenericSystemSubpolicies(ctx, o.BgpPeeringGenericSystems, diags)...)
	result.Subpolicies = append(result.Subpolicies, BgpPeeringIpEndpointSubpolicies(ctx, o.BgpPeeringIpEndpoints, diags)...)
	result.Subpolicies = append(result.Subpolicies, DynamicBgpPeeringSubpolicies(ctx, o.DynamicBgpPeerings, diags)...)
	result.Subpolicies = append(result.Subpolicies, StaticRouteSubpolicies(ctx, o.StaticRoutes, diags)...)

	return &result
}

func IpLinkSubpolicies(ctx context.Context, ipLinkSet types.Set, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var ipLinks []IpLink
	diags.Append(ipLinkSet.ElementsAs(ctx, &ipLinks, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(ipLinks))
	for i, ipLink := range ipLinks {
		subpolicies[i] = ipLink.primitive(ctx, diags)
	}

	return subpolicies
}

func newIpLink(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink, _ *diag.Diagnostics) IpLink {
	result := IpLink{
		// Name:       // handled by caller
		RoutingZoneId: types.StringPointerValue((*string)(in.SecurityZone)),
		// VlanId:      // handled below
		// L3Mtu:       // handled below
		Ipv4AddressingType: types.StringValue(utils.StringersToFriendlyString(in.IPv4AddressingType)),
		Ipv6AddressingType: types.StringValue(utils.StringersToFriendlyString(in.IPv6AddressingType)),
		// StaticRoutes:             handled by caller
		// BgpPeeringGenericSystems: handled by caller
		// BgpPeeringIpEndpoints:    handled by caller
		// DynamicBgpPeerings:       handled by caller
	}

	if in.Vlan != nil {
		result.VlanId = types.Int64Value(int64(*in.Vlan))
	}

	if in.L3Mtu != nil {
		result.L3Mtu = types.Int64Value(int64(*in.L3Mtu))
	}

	return result
}

func IpLinkPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	var result []IpLink

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting IpLink primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newIpLink(ctx, p, diags)
			newPrimitive.Name = utils.StringValueOrNull(ctx, subpolicy.Label, diags)
			newPrimitive.BgpPeeringGenericSystems = BgpPeeringGenericSystemPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.BgpPeeringIpEndpoints = BgpPeeringIpEndpointPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.DynamicBgpPeerings = DynamicBgpPeeringPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.StaticRoutes = StaticRoutePrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result = append(result, newPrimitive)
		}
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: IpLink{}.AttrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: IpLink{}.AttrTypes()}, result, diags)
}
