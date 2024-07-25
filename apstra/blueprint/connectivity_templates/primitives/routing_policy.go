package primitives

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoutingPolicy struct {
	Name            types.String `tfsdk:"name"`
	RoutingPolicyId types.String `tfsdk:"routing_policy_id"`
}

func (o RoutingPolicy) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":              types.StringType,
		"routing_policy_id": types.StringType,
	}
}

func (o RoutingPolicy) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Label used on the Primitive \"block\" in the Connectivity Template",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"routing_policy_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Routing Policy ID to be applied",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o RoutingPolicy) attributes() *apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy {
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy{
		RpToAttach: (*apstra.ObjectId)(o.RoutingPolicyId.ValueStringPointer()),
	}
}

func (o RoutingPolicy) Request(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	return &apstra.ConnectivityTemplatePrimitive{
		Label:      o.Name.ValueString(),
		Attributes: o.attributes(),
	}
}

func newRoutingPolicy(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy, _ *diag.Diagnostics) RoutingPolicy {
	return RoutingPolicy{
		// Name:            utils.StringValueOrNull(ctx, in.Label, diags),
		RoutingPolicyId: types.StringPointerValue((*string)(in.RpToAttach)),
	}
}

func RoutingPolicyPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	var result []RoutingPolicy

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting RoutingPolicy primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newRoutingPolicy(ctx, p, diags)
			newPrimitive.Name = utils.StringValueOrNull(ctx, subpolicy.Label, diags)
			result = append(result, newPrimitive)
		}
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}, result, diags)
}
