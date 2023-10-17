package iba

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IbaProbe struct {
	BlueprintId          types.String `tfsdk:"blueprint_id"`
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	PredefinedIbaProbeId types.String `tfsdk:"predefined_probe_id"`
	IbaProbeConfig       types.String `tfsdk:"probe_config"`
	Stages               types.Set    `tfsdk:"stages"`
}

func (o IbaProbe) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the IBA Probe belongs to.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "IBA Probe ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "IBA Probe Name.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Description of the IBA Probe",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"stages": resourceSchema.SetAttribute{
			MarkdownDescription: "Description of the IBA Probe",
			Computed:            true,
			PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			ElementType:         types.StringType,
		},
		"predefined_probe_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Id of predefined IBA Probe",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"probe_config": resourceSchema.StringAttribute{
			MarkdownDescription: "Configuration elements for the IBA Probe",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
	}
}

// func (o IbaProbe) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
// 	return map[string]dataSourceSchema.Attribute{
// 		"blueprint_id": dataSourceSchema.StringAttribute{
// 			MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the IBA Widget belongs to.",
// 			Required:            true,
// 			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
// 		},
// 		"id": dataSourceSchema.StringAttribute{
// 			MarkdownDescription: "Populate this field to look up a IBA Widget by ID. Required when `name` is omitted.",
// 			Optional:            true,
// 			Computed:            true,
// 			Validators: []validator.String{
// 				stringvalidator.LengthAtLeast(1),
// 				stringvalidator.ExactlyOneOf(path.Expressions{
// 					path.MatchRelative(),
// 					path.MatchRoot("name"),
// 				}...),
// 			},
// 		},
// 		"name": dataSourceSchema.StringAttribute{
// 			MarkdownDescription: "Populate this field to look up a IBA Widget by name. Required when `id` is omitted.",
// 			Optional:            true,
// 			Computed:            true,
// 			Validators: []validator.String{
// 				stringvalidator.LengthAtLeast(1),
// 			},
// 		},
// 		"description": dataSourceSchema.StringAttribute{
// 			MarkdownDescription: "Description of the IBA Widget",
// 			Computed:            true,
// 		},
// 		"default": dataSourceSchema.BoolAttribute{
// 			MarkdownDescription: "True if Default IbaProbe",
// 			Computed:            true,
// 		},
// 		"predefined_dashboard": dataSourceSchema.StringAttribute{
// 			MarkdownDescription: "Id of predefined dashboard if any",
// 			Computed:            true,
// 		},
// 		"updated_by": dataSourceSchema.StringAttribute{
// 			MarkdownDescription: "The user who updated the dashboard last",
// 			Computed:            true,
// 		},
// 		"widget_grid": dataSourceSchema.ListAttribute{
// 			MarkdownDescription: "Grid of Widgets to be displayed in the dashboard",
// 			Computed:            true,
// 			ElementType: types.ListType{
// 				ElemType: types.StringType,
// 			},
// 			Validators: []validator.List{
// 				listvalidator.SizeAtLeast(1),
// 			},
// 		},
// 	}
// }

func (o *IbaProbe) LoadApiData(ctx context.Context, in *apstra.IbaProbe, diag *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)
	o.Description = types.StringValue(in.Description)
	s := make([]string, len(in.Stages))
	for i, j := range in.Stages {
		s[i] = j["name"].(string)
	}
	o.Stages = utils.SetValueOrNull(ctx, types.StringType, s, diag)
}

func (o *IbaProbe) Request(ctx context.Context, d *diag.Diagnostics) *apstra.IbaPredefinedProbeRequest {
	return &apstra.IbaPredefinedProbeRequest{
		Name: o.PredefinedIbaProbeId.ValueString(),
		Data: []byte(o.IbaProbeConfig.ValueString()),
	}
}
