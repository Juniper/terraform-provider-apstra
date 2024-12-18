package primitives

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CustomStaticRoute struct {
	Id            types.String             `tfsdk:"id"`
	BatchId       types.String             `tfsdk:"batch_id"`
	PipelineId    types.String             `tfsdk:"pipeline_id"`
	RoutingZoneId types.String             `tfsdk:"routing_zone_id"`
	Network       customtypes.IPv46Prefix  `tfsdk:"network"`
	NextHop       customtypes.IPv46Address `tfsdk:"next_hop"`
}

func (o CustomStaticRoute) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.StringType,
		"batch_id":        types.StringType,
		"pipeline_id":     types.StringType,
		"routing_zone_id": types.StringType,
		"network":         customtypes.IPv46PrefixType{},
		"next_hop":        customtypes.IPv46AddressType{},
	}
}

func (o CustomStaticRoute) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive element",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"batch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's downstream collection",
			Computed:            true,
		},
		"pipeline_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's upstream pipeline",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Routing Zone ID where this route should be installed",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"network": resourceSchema.StringAttribute{
			MarkdownDescription: "Destination network in CIDR notation",
			CustomType:          customtypes.IPv46PrefixType{},
			Required:            true,
		},
		"next_hop": resourceSchema.StringAttribute{
			MarkdownDescription: "Next-hop router address",
			CustomType:          customtypes.IPv46AddressType{},
			Required:            true,
		},
	}
}

func (o *CustomStaticRoute) ValidateConfig(_ context.Context, path path.Path, diags *diag.Diagnostics) {
	if o.Network.IsUnknown() || o.NextHop.IsUnknown() {
		return
	}

	if (o.Network.Is4() && o.NextHop.Is6()) || (o.Network.Is6() && o.NextHop.Is4()) {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			path,
			fmt.Sprintf("attributes \"network\" and \"next-hop\" must both be IPv4 or both be IPv6, got %s -> %s", o.Network, o.NextHop)),
		)
	}
}

func (o CustomStaticRoute) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute {
	_, network, _ := net.ParseCIDR(o.Network.ValueString())
	nextHop := net.ParseIP(o.NextHop.ValueString())

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute{
		Network:      network,
		NextHop:      nextHop,
		SecurityZone: (*apstra.ObjectId)(o.RoutingZoneId.ValueStringPointer()),
	}
}

func (o CustomStaticRoute) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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

	return &result
}

func CustomStaticRouteSubpolicies(ctx context.Context, customStaticRouteMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var customStaticRoutes map[string]CustomStaticRoute
	diags.Append(customStaticRouteMap.ElementsAs(ctx, &customStaticRoutes, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(customStaticRoutes))
	i := 0
	for k, v := range customStaticRoutes {
		subpolicies[i] = v.primitive(ctx, diags)
		if diags.HasError() {
			return nil
		}
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newCustomStaticRoute(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute, _ *diag.Diagnostics) CustomStaticRoute {
	return CustomStaticRoute{
		// Name:       // handled by caller
		RoutingZoneId: types.StringPointerValue((*string)(in.SecurityZone)),
		Network:       customtypes.NewIPv46PrefixNetPointerValue(in.Network),
		NextHop:       customtypes.NewIPv46PrefixIpValue(in.NextHop),
	}
}

func CustomStaticRoutePrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]CustomStaticRoute)

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting CustomStaticRoute primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newCustomStaticRoute(ctx, p, diags)
			newPrimitive.PipelineId = types.StringPointerValue((*string)(subpolicy.PipelineId))
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			newPrimitive.BatchId = types.StringPointerValue((*string)(subpolicy.BatchId))
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: CustomStaticRoute{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: CustomStaticRoute{}.AttrTypes()}, result, diags)
}

func LoadIDsIntoCustomStaticRouteMap(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, inMap types.Map, diags *diag.Diagnostics) types.Map {
	result := make(map[string]CustomStaticRoute, len(inMap.Elements()))
	inMap.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: CustomStaticRoute{}.AttrTypes()})
	}

	for _, p := range subpolicies {
		if _, ok := p.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute); !ok {
			continue // wrong type and nil value both wind up getting skipped
		}

		if v, ok := result[p.Label]; ok {
			v.PipelineId = types.StringPointerValue((*string)(p.PipelineId))
			v.Id = types.StringPointerValue((*string)(p.Id))
			v.BatchId = types.StringPointerValue((*string)(p.BatchId))
			result[p.Label] = v
		}
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: CustomStaticRoute{}.AttrTypes()}, result, diags)
}
