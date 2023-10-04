package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IbaDashboard struct {
	BlueprintId         types.String `tfsdk:"blueprint_id"`
	Id                  types.String `tfsdk:"id"`
	Label               types.String `tfsdk:"label"`
	Description         types.String `tfsdk:"description"`
	Default             types.Bool   `tfsdk:"default"`
	WidgetGrid          types.List   `tfsdk:"widget_grid"`
	PredefinedDashboard types.String `tfsdk:"predefined_dashboard"`
	UpdatedBy           types.String `tfsdk:"updated_by"`
}

func (o IbaDashboard) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
					path.MatchRoot("label"),
				}...),
			},
		},
		"label": dataSourceSchema.StringAttribute{
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
		"default": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "True if Default Dashboard",
			Computed:            true,
		},
		"predefined_dashboard": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Id of predefined dashboard if any",
			Computed:            true,
		},
		"updated_by": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The user who updated the dashboard last",
			Computed:            true,
		},
		"widget_grid": resourceSchema.ListAttribute{
			MarkdownDescription: fmt.Sprintf("Grid of Widgets to be displayed in the dashboard"),
			Computed:            true,
			ElementType: types.ListType{
				ElemType: types.StringType,
			},
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
	}
}

func (o *IbaDashboard) LoadApiData(ctx context.Context, in *apstra.IbaDashboard, diag *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Label = types.StringValue(in.Data.Label)
	o.Description = types.StringValue(in.Data.Description)
	o.Default = types.BoolValue(in.Data.Default)
	o.PredefinedDashboard = types.StringValue(in.Data.PredefinedDashboard)
	o.UpdatedBy = types.StringValue(in.Data.UpdatedBy)
	o.WidgetGrid = utils.ListValueOrNull(ctx, types.ListType{ElemType: types.StringType}, in.Data.IbaWidgetGrid, diag)
}
