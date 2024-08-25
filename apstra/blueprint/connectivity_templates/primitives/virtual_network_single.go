package primitives

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type VirtualNetworkSingle struct {
	Name                     types.String `tfsdk:"name"`
	VirtualNetworkId         types.String `tfsdk:"virtual_network_id"`
	Tagged                   types.Bool   `tfsdk:"tagged"`
	BgpPeeringGenericSystems types.Set    `tfsdk:"bgp_peering_generic_systems"`
	StaticRoutes             types.Set    `tfsdk:"static_routes"`
}

func (o VirtualNetworkSingle) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                        types.StringType,
		"virtual_network_id":          types.StringType,
		"tagged":                      types.BoolType,
		"bgp_peering_generic_systems": types.SetType{ElemType: types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()}},
		"static_routes":               types.SetType{ElemType: types.ObjectType{AttrTypes: StaticRoute{}.AttrTypes()}},
	}
}

func (o VirtualNetworkSingle) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Label used on the Primitive \"block\" in the Connectivity Template",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the desired Virtual Network",
			Required:            true,
		},
		"tagged": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the selected Virtual Network should be presented with an 802.1Q tag",
			Required:            true,
		},
		"bgp_peering_generic_systems": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of BGP Peering (Generic System) primitives",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: BgpPeeringGenericSystem{}.ResourceAttributes(),
			},
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
			Optional:   true,
		},
		"static_routes": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Static Route primitives",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: StaticRoute{}.ResourceAttributes(),
			},
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
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
		Label:      o.Name.ValueString(),
		Attributes: o.attributes(ctx, diags),
	}
}

func VirtualNetworkSingleSubpolicies(ctx context.Context, virtualNetworkSingleSet types.Set, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var VirtualNetworkSingles []VirtualNetworkSingle
	diags.Append(virtualNetworkSingleSet.ElementsAs(ctx, &VirtualNetworkSingles, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(VirtualNetworkSingles))
	for i, virtualNetworkSingle := range VirtualNetworkSingles {
		subpolicies[i] = virtualNetworkSingle.primitive(ctx, diags)
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

func VirtualNetworkSinglePrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	var result []VirtualNetworkSingle

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
			newPrimitive.Name = utils.StringValueOrNull(ctx, subpolicy.Label, diags)
			newPrimitive.BgpPeeringGenericSystems = BgpPeeringGenericSystemPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.StaticRoutes = StaticRoutePrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result = append(result, newPrimitive)
		}
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: VirtualNetworkSingle{}.AttrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: VirtualNetworkSingle{}.AttrTypes()}, result, diags)
}
