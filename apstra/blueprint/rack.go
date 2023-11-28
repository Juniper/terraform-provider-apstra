package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Rack struct {
	Id          types.String `tfsdk:"id"`
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Name        types.String `tfsdk:"name"`
	PodId       types.String `tfsdk:"pod_id"`
	RackTypeId  types.String `tfsdk:"rack_type_id"`
}

func (o Rack) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Blueprint where the Rack should be created.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name of the Rack.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"pod_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of Pod (3-stage topology) where the new rack should be created. " +
				"Required only in Pod-Based (5-stage) Blueprints.",
			Optional:      true,
			Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"rack_type_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Global Catalog Rack Type design object to use as a template for this Rack.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
	}
}

func (o Rack) Request() *apstra.TwoStageL3ClosRackRequest {
	return &apstra.TwoStageL3ClosRackRequest{
		PodId:      apstra.ObjectId(o.PodId.ValueString()),
		RackTypeId: apstra.ObjectId(o.RackTypeId.ValueString()),
	}
}

func (o Rack) SetName(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	// data structure to use when calling PatchNode
	patch := struct {
		Label string `json:"label"`
	}{
		Label: o.Name.ValueString(),
	}

	err := client.PatchNode(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), apstra.ObjectId(o.Id.ValueString()), &patch, nil)
	if err != nil {
		diags.AddError("Unable to create Datacenter Configlet", err.Error())
		// do not return - we must set the state below
	}
}
