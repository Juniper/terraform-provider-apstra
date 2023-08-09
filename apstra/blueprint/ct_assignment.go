package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

type ConnectivityTemplateAssignment struct {
	BlueprintId            types.String `tfsdk:"blueprint_id"`
	ConnectivityTemplateId types.String `tfsdk:"connectivity_template_id"`
	ApplicationPointIds    types.Set    `tfsdk:"application_point_ids"`
}

func (o ConnectivityTemplateAssignment) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"connectivity_template_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra node ID of the Connectivity Template.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"application_point_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Application Point IDs to which the Connectivity Template should be assigned.",
			Required:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *ConnectivityTemplateAssignment) AddDelRequest(ctx context.Context, state *ConnectivityTemplateAssignment, diags *diag.Diagnostics) ([]apstra.ObjectId, []apstra.ObjectId) {
	var planIds []apstra.ObjectId
	diags.Append(o.ApplicationPointIds.ElementsAs(ctx, &planIds, false)...)
	if diags.HasError() {
		return nil, nil
	}

	if state == nil { // no state in Create()
		return planIds, nil
	}

	var stateIds []apstra.ObjectId
	diags.Append(o.ApplicationPointIds.ElementsAs(ctx, &stateIds, false)...)
	if diags.HasError() {
		return nil, nil
	}

	return utils.SliceComplementOfA(stateIds, planIds), utils.SliceComplementOfA(planIds, stateIds)
}
