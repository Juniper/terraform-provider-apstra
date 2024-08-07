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

type RoutingZoneConstraint struct {
	Name                    types.String `tfsdk:"name"`
	RoutingZoneConstraintId types.String `tfsdk:"routing_zone_constraint_id"`
}

func (o RoutingZoneConstraint) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                       types.StringType,
		"routing_zone_constraint_id": types.StringType,
	}
}

func (o RoutingZoneConstraint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Label used on the Primitive \"block\" in the Connectivity Template",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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
	return &apstra.ConnectivityTemplatePrimitive{
		Label:      o.Name.ValueString(),
		Attributes: o.attributes(ctx, diags),
	}
}

func RoutingZoneConstraintSubpolicies(ctx context.Context, routingZoneConstraintSet types.Set, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var routingZoneConstraints []RoutingZoneConstraint
	diags.Append(routingZoneConstraintSet.ElementsAs(ctx, &routingZoneConstraints, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(routingZoneConstraints))
	for i, routingZoneConstraint := range routingZoneConstraints {
		subpolicies[i] = routingZoneConstraint.primitive(ctx, diags)
	}

	return subpolicies
}

func newRoutingZoneConstraint(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint, _ *diag.Diagnostics) RoutingZoneConstraint {
	return RoutingZoneConstraint{
		// Name:         // handled by caller
		RoutingZoneConstraintId: types.StringPointerValue((*string)(in.RoutingZoneConstraint)),
	}
}

func RoutingZoneConstraintPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Set {
	var result []RoutingZoneConstraint

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
			newPrimitive.Name = utils.StringValueOrNull(ctx, subpolicy.Label, diags)
			result = append(result, newPrimitive)
		}
	}
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: RoutingZoneConstraint{}.AttrTypes()})
	}

	return utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: RoutingZoneConstraint{}.AttrTypes()}, result, diags)
}
