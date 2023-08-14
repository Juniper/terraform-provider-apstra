package design

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

type ConfigletData struct {
	Name       types.String `tfsdk:"name"`
	Generators types.List   `tfsdk:"generators"`
}

func (o ConfigletData) DataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Configlet by name. Required when `id` is omitted.",
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"generators": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "Ordered list of Generators",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: ConfigletGenerator{}.DataSourceAttributesNested(),
			},
		},
	}
}

func (o ConfigletData) ResourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Configlet name displayed in the Apstra web UI",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"generators": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Generators organized by Network OS",
			Required:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ConfigletGenerator{}.ResourceAttributesNested(),
				Validators: []validator.Object{ValidateConfigletGenerator()},
			},
			Validators: []validator.List{listvalidator.SizeAtLeast(1)},
		},
	}
}

func (o ConfigletData) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":       types.StringType,
		"generators": types.ListType{ElemType: types.ObjectType{AttrTypes: ConfigletGenerator{}.AttrTypes()}},
	}
}

func (o *ConfigletData) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ConfigletData {
	var d diag.Diagnostics
	// We only use the Datacenter Reference Design
	refArchs := []apstra.RefDesign{apstra.RefDesignTwoStageL3Clos}
	// Extract configlet generators
	tfGenerators := make([]ConfigletGenerator, len(o.Generators.Elements()))
	d = o.Generators.ElementsAs(ctx, &tfGenerators, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	// Convert configlet generators to apstra types
	generators := make([]apstra.ConfigletGenerator, len(tfGenerators))
	for i, gen := range tfGenerators {
		generators[i] = *gen.Request(ctx, diags)
	}
	if diags.HasError() {
		return nil
	}
	return &apstra.ConfigletData{
		DisplayName: o.Name.ValueString(),
		RefArchs:    refArchs,
		Generators:  generators,
	}
}

func (o *ConfigletData) LoadApiData(ctx context.Context, in *apstra.ConfigletData, diags *diag.Diagnostics) {
	configletGenerators := make([]ConfigletGenerator, len(in.Generators))
	for i := range in.Generators {
		configletGenerators[i].LoadApiData(ctx, &in.Generators[i], diags)
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Generators = utils.ListValueOrNull(ctx, types.ObjectType{AttrTypes: ConfigletGenerator{}.AttrTypes()}, configletGenerators, diags)
}
