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
)

type ConnectivityTemplateAssignments struct {
	BlueprintId            types.String `tfsdk:"blueprint_id"`
	ConnectivityTemplateId types.String `tfsdk:"connectivity_template_id"`
	ApplicationPointIds    types.Set    `tfsdk:"application_point_ids"`
}

func (o ConnectivityTemplateAssignments) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"connectivity_template_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Connectivity Template ID which should be applied to the Application Points.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"application_point_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Apstra node IDs of the Interfaces or Systems where the Connectivity " +
				"Template should be applied.",
			Required:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *ConnectivityTemplateAssignments) Request(ctx context.Context, state *ConnectivityTemplateAssignments, diags *diag.Diagnostics) map[apstra.ObjectId]map[apstra.ObjectId]bool {
	var desired, current []apstra.ObjectId // Application Point IDs

	diags.Append(o.ApplicationPointIds.ElementsAs(ctx, &desired, false)...)
	if diags.HasError() {
		return nil
	}
	desiredMap := make(map[apstra.ObjectId]bool, len(desired))
	for _, apId := range desired {
		desiredMap[apId] = true
	}

	if state != nil {
		diags.Append(state.ApplicationPointIds.ElementsAs(ctx, &current, false)...)
		if diags.HasError() {
			return nil
		}
	}
	currentMap := make(map[apstra.ObjectId]bool, len(current))
	for _, apId := range current {
		currentMap[apId] = true
	}

	result := make(map[apstra.ObjectId]map[apstra.ObjectId]bool)
	ctId := apstra.ObjectId(o.ConnectivityTemplateId.ValueString())

	for _, ApplicationPointId := range desired {
		if _, ok := currentMap[ApplicationPointId]; !ok {
			// desired Application Point not found in currentMap -- need to add
			result[ApplicationPointId] = map[apstra.ObjectId]bool{ctId: true} // causes CT to be added
		}
	}

	for _, ApplicationPointId := range current {
		if _, ok := desiredMap[ApplicationPointId]; !ok {
			// current Application Point not found in desiredMap -- need to remove
			result[ApplicationPointId] = map[apstra.ObjectId]bool{ctId: false} // causes CT to be added
		}
	}

	return result
}
