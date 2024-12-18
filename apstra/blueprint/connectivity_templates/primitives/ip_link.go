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
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IpLink struct {
	Id                       types.String `tfsdk:"id"`
	BatchId                  types.String `tfsdk:"batch_id"`
	PipelineId               types.String `tfsdk:"pipeline_id"`
	RoutingZoneId            types.String `tfsdk:"routing_zone_id"`
	VlanId                   types.Int64  `tfsdk:"vlan_id"`
	L3Mtu                    types.Int64  `tfsdk:"l3_mtu"`
	Ipv4AddressingType       types.String `tfsdk:"ipv4_addressing_type"`
	Ipv6AddressingType       types.String `tfsdk:"ipv6_addressing_type"`
	BgpPeeringGenericSystems types.Map    `tfsdk:"bgp_peering_generic_systems"`
	BgpPeeringIpEndpoints    types.Map    `tfsdk:"bgp_peering_ip_endpoints"`
	DynamicBgpPeerings       types.Map    `tfsdk:"dynamic_bgp_peerings"`
	StaticRoutes             types.Map    `tfsdk:"static_routes"`
}

func (o IpLink) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                          types.StringType,
		"batch_id":                    types.StringType,
		"pipeline_id":                 types.StringType,
		"routing_zone_id":             types.StringType,
		"vlan_id":                     types.Int64Type,
		"l3_mtu":                      types.Int64Type,
		"ipv4_addressing_type":        types.StringType,
		"ipv6_addressing_type":        types.StringType,
		"bgp_peering_generic_systems": types.MapType{ElemType: types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()}},
		"bgp_peering_ip_endpoints":    types.MapType{ElemType: types.ObjectType{AttrTypes: BgpPeeringIpEndpoint{}.AttrTypes()}},
		"dynamic_bgp_peerings":        types.MapType{ElemType: types.ObjectType{AttrTypes: DynamicBgpPeering{}.AttrTypes()}},
		"static_routes":               types.MapType{ElemType: types.ObjectType{AttrTypes: StaticRoute{}.AttrTypes()}},
	}
}

func (o IpLink) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive element",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"batch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's downstream collection",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{ipLinkBatchPlanModifier{}},
		},
		"pipeline_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's upstream pipeline",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
		"bgp_peering_generic_systems": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of BGP Peering (Generic System) primitives",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: BgpPeeringGenericSystem{}.ResourceAttributes(),
			},
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			Optional:   true,
		},
		"bgp_peering_ip_endpoints": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of *BGP Peering (IP Endpoint)* Primitives in this Connectivity Template",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: BgpPeeringIpEndpoint{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
		"dynamic_bgp_peerings": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of *Dynamic BGP Peering* Primitives in this Connectivity Template",
			NestedObject:        resourceSchema.NestedAttributeObject{Attributes: DynamicBgpPeering{}.ResourceAttributes()},
			Optional:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
		"static_routes": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of network IPv4 or IPv6 destination prefixes reachable via this IP Link",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: StaticRoute{}.ResourceAttributes(),
			},
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
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
		SecurityZone:       (*apstra.ObjectId)(o.RoutingZoneId.ValueStringPointer()),
		Tagged:             !o.VlanId.IsNull(),
		Vlan:               vlan,
		IPv4AddressingType: ipv4AddressingType,
		IPv6AddressingType: ipv6AddressingType,
		L3Mtu:              l3Mtu,
	}
}

func (o IpLink) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	result := apstra.ConnectivityTemplatePrimitive{Attributes: o.attributes(ctx, diags)}

	if !o.PipelineId.IsUnknown() {
		result.PipelineId = (*apstra.ObjectId)(o.PipelineId.ValueStringPointer()) // nil when null
	}
	if !o.Id.IsUnknown() {
		result.Id = (*apstra.ObjectId)(o.Id.ValueStringPointer()) // nil when null
	}
	if !o.BatchId.IsUnknown() {
		result.BatchId = (*apstra.ObjectId)(o.BatchId.ValueStringPointer()) // nil when null
	}

	result.Subpolicies = append(result.Subpolicies, BgpPeeringGenericSystemSubpolicies(ctx, o.BgpPeeringGenericSystems, diags)...)
	result.Subpolicies = append(result.Subpolicies, BgpPeeringIpEndpointSubpolicies(ctx, o.BgpPeeringIpEndpoints, diags)...)
	result.Subpolicies = append(result.Subpolicies, DynamicBgpPeeringSubpolicies(ctx, o.DynamicBgpPeerings, diags)...)
	result.Subpolicies = append(result.Subpolicies, StaticRouteSubpolicies(ctx, o.StaticRoutes, diags)...)

	return &result
}

func IpLinkSubpolicies(ctx context.Context, ipLinkMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var ipLinks map[string]IpLink
	diags.Append(ipLinkMap.ElementsAs(ctx, &ipLinks, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(ipLinks))
	i := 0
	for k, v := range ipLinks {
		subpolicies[i] = v.primitive(ctx, diags)
		if diags.HasError() {
			return nil
		}
		subpolicies[i].Label = k
		i++
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

func IpLinkPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]IpLink)

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
			newPrimitive.PipelineId = types.StringPointerValue((*string)(subpolicy.PipelineId))
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			newPrimitive.BatchId = types.StringPointerValue((*string)(subpolicy.BatchId))
			newPrimitive.BgpPeeringGenericSystems = BgpPeeringGenericSystemPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.BgpPeeringIpEndpoints = BgpPeeringIpEndpointPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.DynamicBgpPeerings = DynamicBgpPeeringPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.StaticRoutes = StaticRoutePrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: IpLink{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: IpLink{}.AttrTypes()}, result, diags)
}

func LoadIDsIntoIpLinkMap(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, inMap types.Map, diags *diag.Diagnostics) types.Map {
	result := make(map[string]IpLink, len(inMap.Elements()))
	inMap.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: IpLink{}.AttrTypes()})
	}

	for _, p := range subpolicies {
		if _, ok := p.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink); !ok {
			continue // wrong type and nil value both wind up getting skipped
		}

		if v, ok := result[p.Label]; ok {
			v.PipelineId = types.StringPointerValue((*string)(p.PipelineId))
			v.Id = types.StringPointerValue((*string)(p.Id))
			v.BatchId = types.StringPointerValue((*string)(p.BatchId))
			v.BgpPeeringGenericSystems = LoadIDsIntoBgpPeeringGenericSystemMap(ctx, p.Subpolicies, v.BgpPeeringGenericSystems, diags)
			v.BgpPeeringIpEndpoints = LoadIDsIntoBgpPeeringIpEndpointMap(ctx, p.Subpolicies, v.BgpPeeringIpEndpoints, diags)
			v.DynamicBgpPeerings = LoadIDsIntoDynamicBgpPeeringMap(ctx, p.Subpolicies, v.DynamicBgpPeerings, diags)
			v.StaticRoutes = LoadIDsIntoStaticRouteMap(ctx, p.Subpolicies, v.StaticRoutes, diags)
			result[p.Label] = v
		}
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: IpLink{}.AttrTypes()}, result, diags)
}

var _ planmodifier.String = (*ipLinkBatchPlanModifier)(nil)

type ipLinkBatchPlanModifier struct{}

func (o ipLinkBatchPlanModifier) Description(_ context.Context) string {
	return "preserves the the state value unless all child primitives have been removed, in which case null is planned"
}

func (o ipLinkBatchPlanModifier) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ipLinkBatchPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var plan IpLink
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, req.Path.ParentPath(), &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// do we have any children?
	if len(plan.BgpPeeringGenericSystems.Elements())+
		len(plan.BgpPeeringIpEndpoints.Elements())+
		len(plan.DynamicBgpPeerings.Elements())+
		len(plan.StaticRoutes.Elements()) == 0 {
		resp.PlanValue = types.StringNull() // with no children the batch id should be null
		return
	}

	// are we a new object?
	if plan.Id.IsUnknown() {
		resp.PlanValue = types.StringUnknown() // we are a new object. the batch id is not knowable
	}

	// we're not new, and we have children. use the old value
	resp.PlanValue = req.StateValue
}
