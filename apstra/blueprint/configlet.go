package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterConfiglet struct {
	BlueprintId        types.String `tfsdk:"blueprint_id"`
	Id                 types.String `tfsdk:"id"`
	Condition          types.String `tfsdk:"condition"`
	CatalogConfigletID types.String `tfsdk:"catalog_configlet_id"`
	Name               types.String `tfsdk:"name"`
	Generators         types.List   `tfsdk:"generators"`
}

func (o DatacenterConfiglet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the Configlet belongs to.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Configlet by ID. Required when `name` is omitted.",
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
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"condition": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Condition determines where the Configlet is applied.",
			Computed:            true,
		},
		"catalog_configlet_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Will be null in the data source",
			Computed:            true,
		},
		"generators": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered list of Generators",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: design.ConfigletGenerator{}.DataSourceAttributesNested(),
			},
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
			MarkdownDescription: "Configlet ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Configlet name. When omitted, the name found in the catalog will be used.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"condition": resourceSchema.StringAttribute{
			MarkdownDescription: "Condition determines where the Configlet is applied.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"catalog_configlet_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Id of the catalog Configlet to be imported",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(
					path.MatchRelative(),
					path.MatchRoot("generators"),
				),
				stringvalidator.AtLeastOneOf(
					path.MatchRelative(),
					path.MatchRoot("name"),
				),
			},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"generators": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered list of Generators",
			Optional:            true,
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: design.ConfigletGenerator{}.ResourceAttributesNested(),
			},
			PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			Validators:    []validator.List{listvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *DatacenterConfiglet) LoadApiData(ctx context.Context, in *apstra.TwoStageL3ClosConfigletData, diags *diag.Diagnostics) {
	var configlet design.Configlet
	configlet.LoadApiData(ctx, in.Data, diags)
	if diags.HasError() {
		return
	}

	o.Condition = types.StringValue(in.Condition)
	o.Name = types.StringValue(in.Label)
	o.Generators = configlet.Generators
}

func (o *DatacenterConfiglet) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.TwoStageL3ClosConfigletData {
	var c apstra.TwoStageL3ClosConfigletData

	c.Label = o.Name.ValueString()
	c.Condition = o.Condition.ValueString()

	cfglet := design.Configlet{
		Name:       o.Name,
		Generators: o.Generators,
	}
	c.Data = cfglet.Request(ctx, diags)

	if diags.HasError() {
		return nil
	}
	return &c
}
