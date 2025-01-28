package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
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
	Name               types.String `tfsdk:"name"`
	Condition          types.String `tfsdk:"condition"`
	CatalogConfigletID types.String `tfsdk:"catalog_configlet_id"`
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
				Attributes: ConfigletGenerator{}.DataSourceAttributes(),
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
			MarkdownDescription: "Configlet name. When omitted, the name found in the catalog configlet will be used." +
				" Required when the `generators` attribute is specified.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.RequiredWhenValueNull(path.MatchRoot("catalog_configlet_id")),
			},
		},
		"condition": resourceSchema.StringAttribute{
			MarkdownDescription: "Condition determines where the Configlet is applied.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"catalog_configlet_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Id of the catalog configlet to be imported. " +
				"This is an alternative to specifying the `generators` attribute",
			Optional: true,
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
			MarkdownDescription: "Ordered list of Generators. " +
				"This is an alternative to specifying the `catalog_configlet_id` attribute",
			Optional: true,
			Computed: true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ConfigletGenerator{}.ResourceAttributes(),
			},
			PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			Validators:    []validator.List{listvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *DatacenterConfiglet) LoadApiData(ctx context.Context, in *apstra.TwoStageL3ClosConfigletData, diags *diag.Diagnostics) {
	generators := make([]ConfigletGenerator, len(in.Data.Generators))
	for i, generator := range in.Data.Generators {
		generators[i].LoadApiData(ctx, &generator, diags)
	}

	o.Condition = types.StringValue(in.Condition)
	o.Name = types.StringValue(in.Label)
	o.Generators = utils.ListValueOrNull(ctx, types.ObjectType{AttrTypes: ConfigletGenerator{}.AttrTypes()}, generators, diags)
}

func (o *DatacenterConfiglet) LoadCatalogConfigletData(ctx context.Context, in *apstra.ConfigletData, diags *diag.Diagnostics) {
	if o.Name.IsUnknown() {
		o.Name = types.StringValue(in.DisplayName)
	}

	generators := make([]ConfigletGenerator, len(in.Generators))
	for i, generator := range in.Generators {
		generators[i].LoadApiData(ctx, &generator, diags)
	}
	if diags.HasError() {
		return
	}

	o.Generators = utils.ListValueOrNull(ctx, types.ObjectType{AttrTypes: ConfigletGenerator{}.AttrTypes()}, generators, diags)
	if diags.HasError() {
		return
	}
}

func (o *DatacenterConfiglet) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.TwoStageL3ClosConfigletData {
	result := apstra.TwoStageL3ClosConfigletData{
		Label:     o.Name.ValueString(),
		Condition: o.Condition.ValueString(),
		Data: &apstra.ConfigletData{
			DisplayName: o.Name.ValueString(),
			Generators:  nil, //  handled below
		},
	}

	var generators []ConfigletGenerator
	diags.Append(o.Generators.ElementsAs(ctx, &generators, false)...)
	if diags.HasError() {
		return nil
	}

	result.Data.Generators = make([]apstra.ConfigletGenerator, len(generators))
	for i, generator := range generators {
		result.Data.Generators[i] = *generator.Request(ctx, diags)
	}

	return &result
}
