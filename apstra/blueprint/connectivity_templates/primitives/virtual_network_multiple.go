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

type VirtualNetworkMultiple struct {
	Name         types.String `tfsdk:"name"`
	UntaggedVnId types.String `tfsdk:"untagged_vn_id"`
	TaggedVnIds  types.Set    `tfsdk:"tagged_vn_ids"`
}

func (o VirtualNetworkMultiple) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"untagged_vn_id": types.StringType,
		"tagged_vn_ids":  types.SetType{ElemType: types.StringType},
	}
}

func (o VirtualNetworkMultiple) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Label used on the Primitive \"block\" in the Connectivity Template",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"untagged_vn_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Virtual Network which should be untagged on the link",
			Optional:            true,
		},
		"tagged_vn_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "IDs of the Virtual Networks which should be tagged on the link",
			Optional:            true,
			ElementType:         types.StringType,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o VirtualNetworkMultiple) attributes(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachMultipleVlan {
	var taggedVnNodeIds []apstra.ObjectId
	diags.Append(o.TaggedVnIds.ElementsAs(ctx, &taggedVnNodeIds, false)...)
	if diags.HasError() {
		return nil
	}

	// Don't send `null` to the API. Send `[]` instead.
	if taggedVnNodeIds == nil {
		taggedVnNodeIds = []apstra.ObjectId{}
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachMultipleVlan{
		UntaggedVnNodeId: (*apstra.ObjectId)(o.UntaggedVnId.ValueStringPointer()),
		TaggedVnNodeIds:  taggedVnNodeIds,
	}
}

func (o VirtualNetworkMultiple) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	return &apstra.ConnectivityTemplatePrimitive{
		Label:      o.Name.ValueString(),
		Attributes: o.attributes(ctx, diags),
	}
}

func VirtualNetworkMultipleSubpolicies(ctx context.Context, virtualNetworkMultipleSet types.Set, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var VirtualNetworkMultiples []VirtualNetworkMultiple
	diags.Append(virtualNetworkMultipleSet.ElementsAs(ctx, &VirtualNetworkMultiples, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(VirtualNetworkMultiples))
	for i, virtualNetworkMultiple := range VirtualNetworkMultiples {
		subpolicies[i] = virtualNetworkMultiple.primitive(ctx, diags)
	}

	return subpolicies
}

func newVirtualNetworkMultiple(ctx context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachMultipleVlan, diags *diag.Diagnostics) VirtualNetworkMultiple {
	return VirtualNetworkMultiple{
		// Name: // handled by caller
		UntaggedVnId: types.StringPointerValue((*string)(in.UntaggedVnNodeId)),
		TaggedVnIds:  utils.SetValueOrNull(ctx, types.StringType, in.TaggedVnNodeIds, diags),
	}
}

func VirtualNetworkMultiplePrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	var result []VirtualNetworkMultiple

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachMultipleVlan); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting RoutingPolicy primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newVirtualNetworkMultiple(ctx, p, diags)
			newPrimitive.Name = utils.StringValueOrNull(ctx, subpolicy.Label, diags)
			result = append(result, newPrimitive)
		}
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: VirtualNetworkMultiple{}.AttrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: VirtualNetworkMultiple{}.AttrTypes()}, result, diags)
}
