package primitives

import (
	"context"
	"fmt"
	"net"

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

func (o *CustomStaticRoute) loadApiData(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	customStaticRoute := (in.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute)

	o.Name = utils.StringValueOrNull(ctx, in.Label, diags)
	o.RoutingZoneId = types.StringPointerValue((*string)(customStaticRoute.SecurityZone))
	if customStaticRoute.Network != nil {
		o.Network = customtypes.NewIPv46PrefixValue(customStaticRoute.Network.String())
	} else {
		o.Network = customtypes.NewIPv46PrefixNull()
	}
	if customStaticRoute.NextHop != nil {
		o.NextHop = customtypes.NewIPv46AddressValue(customStaticRoute.NextHop.String())
	} else {
		o.NextHop = customtypes.NewIPv46AddressNull()
	}
}

func NewSetCustomStaticRoutes(ctx context.Context, in []apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	customStaticRoutes := make([]CustomStaticRoute, len(in))
	for i, primitive := range in {
		_, ok := (primitive.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute)
		if !ok {
			diags.AddError(
				constants.ErrProviderBug,
				fmt.Sprintf("primitive %d should have type %T, got %T", i, apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute{}, primitive.Attributes))
			continue
		}
		customStaticRoutes[i].loadApiData(ctx, primitive, diags)
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: CustomStaticRoute{}.AttrTypes()}, customStaticRoutes, diags)
}
