package primitives

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkMultiple struct {
	Id           types.String `tfsdk:"id"`
	UntaggedVnId types.String `tfsdk:"untagged_vn_id"`
	TaggedVnIds  types.Set    `tfsdk:"tagged_vn_ids"`
}

func (o VirtualNetworkMultiple) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.StringType,
		"untagged_vn_id": types.StringType,
		"tagged_vn_ids":  types.SetType{ElemType: types.StringType},
	}
}

func (o VirtualNetworkMultiple) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive element",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
	if !utils.HasValue(o.Id) {
		o.Id = utils.NewUuidStringVal(diags)
		if diags.HasError() {
			return nil
		}
	}

	return &apstra.ConnectivityTemplatePrimitive{
		Id: (*apstra.ObjectId)(o.Id.ValueStringPointer()),
		// Label:       // set by caller
		Attributes: o.attributes(ctx, diags),
	}
}

func VirtualNetworkMultipleSubpolicies(ctx context.Context, virtualNetworkMultipleMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var VirtualNetworkMultiples map[string]VirtualNetworkMultiple
	diags.Append(virtualNetworkMultipleMap.ElementsAs(ctx, &VirtualNetworkMultiples, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(VirtualNetworkMultiples))
	i := 0
	for k, v := range VirtualNetworkMultiples {
		subpolicies[i] = v.primitive(ctx, diags)
		if diags.HasError() {
			return nil
		}
		subpolicies[i].Label = k
		i++
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

func VirtualNetworkMultiplePrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]VirtualNetworkMultiple)

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
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: VirtualNetworkMultiple{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: VirtualNetworkMultiple{}.AttrTypes()}, result, diags)
}
