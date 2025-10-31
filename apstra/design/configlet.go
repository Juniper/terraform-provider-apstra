package design

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

type Configlet struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Generators types.List   `tfsdk:"generators"`
}

func (o Configlet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
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

func (o Configlet) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID number of Configlet",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
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
			},
			Validators: []validator.List{listvalidator.SizeAtLeast(1)},
		},
	}
}

func (o Configlet) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ConfigletData {
	var tfGenerators []ConfigletGenerator
	diags.Append(o.Generators.ElementsAs(ctx, &tfGenerators, false)...)
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
		RefArchs:    []enum.RefDesign{enum.RefDesignDatacenter}, // We only use the Datacenter Reference Design
		Generators:  generators,
	}
}

func (o *Configlet) LoadApiData(ctx context.Context, in *apstra.ConfigletData, diags *diag.Diagnostics) {
	configletGenerators := make([]ConfigletGenerator, len(in.Generators))
	for i := range in.Generators {
		configletGenerators[i].LoadApiData(ctx, &in.Generators[i], diags)
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Generators = value.ListOrNull(ctx, types.ObjectType{AttrTypes: ConfigletGenerator{}.AttrTypes()}, configletGenerators, diags)
}
