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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StaticRoute struct {
	Network customtypes.IPv46Prefix `tfsdk:"network"`
	Shared  types.Bool              `tfsdk:"share_ip_endpoint"`
}

func (o StaticRoute) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"network":           customtypes.IPv46PrefixType{},
		"share_ip_endpoint": types.BoolType,
	}
}

func (o StaticRoute) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"network": resourceSchema.StringAttribute{
			MarkdownDescription: "Destination network in CIDR notation",
			CustomType:          customtypes.IPv46PrefixType{},
			Required:            true,
		},
		"share_ip_endpoint": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the next-hop IP address is shared across " +
				"multiple remote systems. Default:  Default: `false`",
			Required: true,
		},
	}
}

func (o StaticRoute) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachStaticRoute {
	_, network, _ := net.ParseCIDR(o.Network.ValueString())

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachStaticRoute{
		Network:         network,
		ShareIpEndpoint: o.Shared.ValueBool(),
	}
}

func (o StaticRoute) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	return &apstra.ConnectivityTemplatePrimitive{
		// Label:       // set by caller
		Attributes: o.attributes(ctx, diags),
	}
}

func StaticRouteSubpolicies(ctx context.Context, StaticRouteMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var StaticRoutes map[string]StaticRoute
	diags.Append(StaticRouteMap.ElementsAs(ctx, &StaticRoutes, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(StaticRoutes))
	i := 0
	for k, v := range StaticRoutes {
		subpolicies[i] = v.primitive(ctx, diags)
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newStaticRoute(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachStaticRoute, _ *diag.Diagnostics) StaticRoute {
	return StaticRoute{
		// Name: // handled by caller
		Network: customtypes.NewIPv46PrefixNetPointerValue(in.Network),
		Shared:  types.BoolValue(in.ShareIpEndpoint),
	}
}

func StaticRoutePrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]StaticRoute, len(subpolicies))

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachStaticRoute); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting RoutingPolicy primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newStaticRoute(ctx, p, diags)
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: StaticRoute{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: StaticRoute{}.AttrTypes()}, result, diags)
}
