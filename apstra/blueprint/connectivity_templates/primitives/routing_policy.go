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
	RoutingPolicyId types.String `tfsdk:"routing_policy_id"`
}

func (o RoutingPolicy) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"routing_policy_id": types.StringType,
	}
}

func (o RoutingPolicy) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"routing_policy_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Routing Policy ID to be applied",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o RoutingPolicy) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy {
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy{
		RpToAttach: (*apstra.ObjectId)(o.RoutingPolicyId.ValueStringPointer()),
	}
}

func (o RoutingPolicy) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	return &apstra.ConnectivityTemplatePrimitive{
		// Label:       // set by caller
		Attributes: o.attributes(ctx, diags),
	}
}

func RoutingPolicySubpolicies(ctx context.Context, routingPolicyMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var routingPolicies map[string]RoutingPolicy
	diags.Append(routingPolicyMap.ElementsAs(ctx, &routingPolicies, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(routingPolicies))
	i := 0
	for k, v := range routingPolicies {
		subpolicies[i] = v.primitive(ctx, diags)
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newRoutingPolicy(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy, _ *diag.Diagnostics) RoutingPolicy {
	return RoutingPolicy{
		// Name:         // handled by caller
		RoutingPolicyId: types.StringPointerValue((*string)(in.RpToAttach)),
	}
}

func RoutingPolicyPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]RoutingPolicy, len(subpolicies))

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
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}, result, diags)
}
