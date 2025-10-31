package primitives

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoutingPolicy struct {
	Id              types.String `tfsdk:"id"`
	PipelineId      types.String `tfsdk:"pipeline_id"`
	RoutingPolicyId types.String `tfsdk:"routing_policy_id"`
}

func (o RoutingPolicy) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.StringType,
		"pipeline_id":       types.StringType,
		"routing_policy_id": types.StringType,
	}
}

func (o RoutingPolicy) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive element",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"pipeline_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's upstream pipeline",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
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
	result := apstra.ConnectivityTemplatePrimitive{Attributes: o.attributes(ctx, diags)}

	if !o.PipelineId.IsUnknown() {
		result.PipelineId = (*apstra.ObjectId)(o.PipelineId.ValueStringPointer()) // nil when null
	}
	if !o.Id.IsUnknown() {
		result.Id = (*apstra.ObjectId)(o.Id.ValueStringPointer()) // nil when null
	}

	return &result
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
		if diags.HasError() {
			return nil
		}
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
	result := make(map[string]RoutingPolicy)

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
			newPrimitive.PipelineId = types.StringPointerValue((*string)(subpolicy.PipelineId))
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()})
	}

	return value.MapOrNull(ctx, types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}, result, diags)
}

func LoadIDsIntoRoutingPolicyMap(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, inMap types.Map, diags *diag.Diagnostics) types.Map {
	result := make(map[string]RoutingPolicy, len(inMap.Elements()))
	inMap.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()})
	}

	for _, p := range subpolicies {
		if _, ok := p.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy); !ok {
			continue // wrong type and nil value both wind up getting skipped
		}

		if v, ok := result[p.Label]; ok {
			v.PipelineId = types.StringPointerValue((*string)(p.PipelineId))
			v.Id = types.StringPointerValue((*string)(p.Id))
			result[p.Label] = v
		}
	}

	return value.MapOrNull(ctx, types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}, result, diags)
}
