package iba

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IbaWidget struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Stage       types.String `tfsdk:"stage"`
	ProbeId     types.String `tfsdk:"probe_id"`
}

func (o IbaWidget) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the IBA Widget belongs to.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a IBA Widget by ID. Required when `name` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a IBA Widget by name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Description of the IBA Widget",
			Computed:            true,
		},
		"stage": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Stage of IBA Probe used by this widget",
			Computed:            true,
		},
		"probe_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Id of IBA Probe used by this widget",
			Computed:            true,
		},
	}
}

func (o IbaWidget) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Apstra Blueprint where the IBA Widget will be created",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "IBA Widget ID",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "IBA Widget Name",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "IBA Widget Description",
			Required:            true,
		},
		"probe_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Id of IBA Probe used by this widget",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"stage": resourceSchema.StringAttribute{
			MarkdownDescription: "Stage of IBA Probe used by this widget",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	}
}

func (o *IbaWidget) LoadApiData(_ context.Context, in *apstra.IbaWidget, _ *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Data.Label)
	o.Description = types.StringValue(in.Data.Description)
	o.Stage = types.StringValue(in.Data.StageName)
	o.ProbeId = types.StringValue(in.Data.ProbeId.String())
}

func (o *IbaWidget) Request(ctx context.Context, d *diag.Diagnostics) *apstra.IbaWidgetData {

	return &apstra.IbaWidgetData{
		StageName:   o.Stage.ValueString(),
		Description: o.Description.ValueString(),
		ProbeId:     apstra.ObjectId(o.ProbeId.ValueString()),
		Label:       o.Name.ValueString(),
		Type:        apstra.IbaWidgetTypeStage,
	}
}
