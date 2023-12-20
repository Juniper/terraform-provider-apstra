package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectivityTemplatesAssignment struct {
	BlueprintId             types.String `tfsdk:"blueprint_id"`
	ConnectivityTemplateIds types.Set    `tfsdk:"connectivity_template_ids"`
	ApplicationPointId      types.String `tfsdk:"application_point_id"`
}

func (o ConnectivityTemplatesAssignment) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"application_point_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra node ID of the Interface or System where the Connectivity Templates " +
				"should be applied.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"connectivity_template_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Connectivity Template IDs which should be applied to the Application Point.",
			Required:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *ConnectivityTemplatesAssignment) AddDelRequest(ctx context.Context, state *ConnectivityTemplatesAssignment, diags *diag.Diagnostics) ([]apstra.ObjectId, []apstra.ObjectId) {
	var planIds, stateIds []apstra.ObjectId

	if o != nil { // o will be nil in Delete()
		diags.Append(o.ConnectivityTemplateIds.ElementsAs(ctx, &planIds, false)...)
		if diags.HasError() {
			return nil, nil
		}
	}

	if state != nil { // state will be nil in Create()
		diags.Append(state.ConnectivityTemplateIds.ElementsAs(ctx, &stateIds, false)...)
		if diags.HasError() {
			return nil, nil
		}
	}

	return utils.SliceComplementOfA(stateIds, planIds), utils.SliceComplementOfA(planIds, stateIds)
}
