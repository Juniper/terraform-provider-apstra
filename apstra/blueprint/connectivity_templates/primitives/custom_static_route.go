package primitives

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CustomStaticRoute struct {
	Name          types.String             `tfsdk:"name"`
	RoutingZoneId types.String             `tfsdk:"routing_zone_id"`
	Network       customtypes.IPv46Prefix  `tfsdk:"network"`
	NextHop       customtypes.IPv46Address `tfsdk:"next_hop"`
}

func (o CustomStaticRoute) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":            types.StringType,
		"routing_zone_id": types.StringType,
		"network":         customtypes.IPv46PrefixType{},
		"next_hop":        customtypes.IPv46AddressType{},
	}
}

func (o CustomStaticRoute) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Label used on the Primitive \"block\" in the Connectivity Template",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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

func (o CustomStaticRoute) attributes() *apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute {
	_, network, _ := net.ParseCIDR(o.Network.ValueString())
	nextHop := net.ParseIP(o.NextHop.ValueString())

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute{
		Label:        o.NextHop.ValueString(),
		Network:      network,
		NextHop:      nextHop,
		SecurityZone: (*apstra.ObjectId)(o.RoutingZoneId.ValueStringPointer()),
	}
}

func (o CustomStaticRoute) Request() *apstra.ConnectivityTemplatePrimitive {
	return &apstra.ConnectivityTemplatePrimitive{
		Label:      o.Name.ValueString(),
		Attributes: o.attributes(),
	}
}

func newCustomStaticRoute(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute, _ *diag.Diagnostics) CustomStaticRoute {
	result := CustomStaticRoute{
		// Name:          utils.StringValueOrNull(ctx, in.Label, diags),
		RoutingZoneId: types.StringPointerValue((*string)(in.SecurityZone)),
		Network:       customtypes.NewIPv46PrefixNull(),
		NextHop:       customtypes.NewIPv46AddressNull(),
	}

	if in.Network != nil {
		result.Network = customtypes.NewIPv46PrefixValue(in.Network.String())
	}

	if in.NextHop != nil {
		result.NextHop = customtypes.NewIPv46AddressValue(in.NextHop.String())
	}

	return result
}

func CustomStaticRoutePrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	var result []CustomStaticRoute

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting RoutingPolicy primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newCustomStaticRoute(ctx, p, diags)
			newPrimitive.Name = utils.StringValueOrNull(ctx, subpolicy.Label, diags)
			result = append(result, newPrimitive)
		}
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: CustomStaticRoute{}.AttrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: CustomStaticRoute{}.AttrTypes()}, result, diags)
}
