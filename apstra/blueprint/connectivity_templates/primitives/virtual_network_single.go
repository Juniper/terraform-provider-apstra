package primitives

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkSingle struct {
	VirtualNetworkId         types.String `tfsdk:"virtual_network_id"`
	Tagged                   types.Bool   `tfsdk:"tagged"`
	BgpPeeringGenericSystems types.Map    `tfsdk:"bgp_peering_generic_systems"`
	StaticRoutes             types.Map    `tfsdk:"static_routes"`
}

func (o VirtualNetworkSingle) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"virtual_network_id":          types.StringType,
		"tagged":                      types.BoolType,
		"bgp_peering_generic_systems": types.MapType{ElemType: types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()}},
		"static_routes":               types.MapType{ElemType: types.ObjectType{AttrTypes: StaticRoute{}.AttrTypes()}},
	}
}

func (o VirtualNetworkSingle) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the desired Virtual Network",
			Required:            true,
		},
		"tagged": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the selected Virtual Network should be presented with an 802.1Q tag",
			Required:            true,
		},
		"bgp_peering_generic_systems": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of BGP Peering (Generic System) primitives",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: BgpPeeringGenericSystem{}.ResourceAttributes(),
			},
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			Optional:   true,
		},
		"static_routes": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Static Route primitives",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: StaticRoute{}.ResourceAttributes(),
			},
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			Optional:   true,
		},
	}
}

func (o VirtualNetworkSingle) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan {
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan{
		Tagged:   o.Tagged.ValueBool(),
		VnNodeId: (*apstra.ObjectId)(o.VirtualNetworkId.ValueStringPointer()),
	}
}

func (o VirtualNetworkSingle) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	return &apstra.ConnectivityTemplatePrimitive{
		// Label:       // set by caller
		Attributes: o.attributes(ctx, diags),
	}
}

func VirtualNetworkSingleSubpolicies(ctx context.Context, virtualNetworkSingleMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var VirtualNetworkSingles map[string]VirtualNetworkSingle
	diags.Append(virtualNetworkSingleMap.ElementsAs(ctx, &VirtualNetworkSingles, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(VirtualNetworkSingles))
	i := 0
	for k, v := range VirtualNetworkSingles {
		subpolicies[i] = v.primitive(ctx, diags)
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newVirtualNetworkSingle(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan, _ *diag.Diagnostics) VirtualNetworkSingle {
	return VirtualNetworkSingle{
		// Name:          // handled by caller
		VirtualNetworkId: types.StringPointerValue((*string)(in.VnNodeId)),
		Tagged:           types.BoolValue(in.Tagged),
	}
}

func VirtualNetworkSinglePrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]VirtualNetworkSingle, len(subpolicies))

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting RoutingPolicy primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newVirtualNetworkSingle(ctx, p, diags)
			newPrimitive.BgpPeeringGenericSystems = BgpPeeringGenericSystemPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.StaticRoutes = StaticRoutePrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: VirtualNetworkSingle{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: VirtualNetworkSingle{}.AttrTypes()}, result, diags)
}
