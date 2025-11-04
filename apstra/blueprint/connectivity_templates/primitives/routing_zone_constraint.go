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

type RoutingZoneConstraint struct {
	Id                      types.String `tfsdk:"id"`
	PipelineId              types.String `tfsdk:"pipeline_id"`
	RoutingZoneConstraintId types.String `tfsdk:"routing_zone_constraint_id"`
}

func (o RoutingZoneConstraint) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                         types.StringType,
		"pipeline_id":                types.StringType,
		"routing_zone_constraint_id": types.StringType,
	}
}

func (o RoutingZoneConstraint) ResourceAttributes() map[string]resourceSchema.Attribute {
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
		"routing_zone_constraint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Routing Zone Constraint ID to be applied",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o RoutingZoneConstraint) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint {
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint{
		RoutingZoneConstraint: (*apstra.ObjectId)(o.RoutingZoneConstraintId.ValueStringPointer()),
	}
}

func (o RoutingZoneConstraint) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	result := apstra.ConnectivityTemplatePrimitive{Attributes: o.attributes(ctx, diags)}

	if !o.PipelineId.IsUnknown() {
		result.PipelineId = (*apstra.ObjectId)(o.PipelineId.ValueStringPointer()) // nil when null
	}
	if !o.Id.IsUnknown() {
		result.Id = (*apstra.ObjectId)(o.Id.ValueStringPointer()) // nil when null
	}

	return &result
}

func RoutingZoneConstraintSubpolicies(ctx context.Context, routingZoneConstraintMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var routingZoneConstraints map[string]RoutingZoneConstraint
	diags.Append(routingZoneConstraintMap.ElementsAs(ctx, &routingZoneConstraints, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(routingZoneConstraints))
	i := 0
	for k, v := range routingZoneConstraints {
		subpolicies[i] = v.primitive(ctx, diags)
		if diags.HasError() {
			return nil
		}
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newRoutingZoneConstraint(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint, _ *diag.Diagnostics) RoutingZoneConstraint {
	return RoutingZoneConstraint{
		// Name:         // handled by caller
		RoutingZoneConstraintId: types.StringPointerValue((*string)(in.RoutingZoneConstraint)),
	}
}

func RoutingZoneConstraintPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]RoutingZoneConstraint)

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting RoutingZoneConstraint primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newRoutingZoneConstraint(ctx, p, diags)
			newPrimitive.PipelineId = types.StringPointerValue((*string)(subpolicy.PipelineId))
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: RoutingZoneConstraint{}.AttrTypes()})
	}

	return value.MapOrNull(ctx, types.ObjectType{AttrTypes: RoutingZoneConstraint{}.AttrTypes()}, result, diags)
}

func LoadIDsIntoRoutingZoneConstraintMap(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, inMap types.Map, diags *diag.Diagnostics) types.Map {
	result := make(map[string]RoutingZoneConstraint, len(inMap.Elements()))
	inMap.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: RoutingZoneConstraint{}.AttrTypes()})
	}

	for _, p := range subpolicies {
		if _, ok := p.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint); !ok {
			continue // wrong type and nil value both wind up getting skipped
		}

		if v, ok := result[p.Label]; ok {
			v.PipelineId = types.StringPointerValue((*string)(p.PipelineId))
			v.Id = types.StringPointerValue((*string)(p.Id))
			result[p.Label] = v
		}
	}

	return value.MapOrNull(ctx, types.ObjectType{AttrTypes: RoutingZoneConstraint{}.AttrTypes()}, result, diags)
}
