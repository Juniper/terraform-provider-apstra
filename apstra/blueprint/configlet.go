package blueprint

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

type DatacenterConfiglet struct {
	BlueprintId        types.String `tfsdk:"blueprint_id"`
	Id                 types.String `tfsdk:"id"`
	CatalogConfigletID types.String `tfsdk:"catalog_configlet_id"`
	Condition          types.String `tfsdk:"condition"`
	Name               types.String `tfsdk:"name"`
}

func (o DatacenterConfiglet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint that the Configlet belongs to.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Configlet by Id. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up a Configlet by name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true},
		"condition": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Condition that decides how the configlet is applied",
			Computed:            true,
		},
		"catalog_configlet_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Will be null in the data source",
			Computed:            true,
		},
	}
}

func (o DatacenterConfiglet) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the Configlet belongs to.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Configlet Id",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Configlet name.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"condition": resourceSchema.StringAttribute{
			MarkdownDescription: "Condition that determines when the configlet is applied",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"catalog_configlet_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Id of the Catalog Configlet that is to be imported",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o *DatacenterConfiglet) LoadApiData(ctx context.Context, in *apstra.TwoStageL3ClosConfiglet, diags *diag.Diagnostics) {
	o.Condition = types.StringValue(in.Data.Condition)
	o.Name = types.StringValue(in.Data.Label)
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Data.Label)
}
