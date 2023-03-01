package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

type configlet struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RefArchs   types.Set    `tfsdk:"ref_archs"`
	Generators types.List   `tfsdk:"generators"`
}

func (o configlet) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Configlet by ID. Required when `name`is omitted.",
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
			MarkdownDescription: "Populate this field to look up a Configlet by name. Required when `id`is omitted.",
			Optional:            true,
			Computed:            true,
		},
		"ref_archs": dataSourceSchema.SetAttribute{
			MarkdownDescription: "List of architectures",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"generators": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "Generators organized by Network OS",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: configletGenerator{}.dataSourceAttributes(),
			},
		},
	}
}

func (o *configlet) loadApiData(ctx context.Context, in *goapstra.ConfigletData, diags *diag.Diagnostics) {
	refArchs := make([]string, len(in.RefArchs))
	for i, refArch := range in.RefArchs {
		refArchs[i] = refArch.String()
	}

	configletGenerators := make([]configletGenerator, len(in.Generators))
	for i := range in.Generators {
		configletGenerators[i].loadApiData(ctx, &in.Generators[i], diags)
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.RefArchs = utils.SetValueOrNull(ctx, types.StringType, refArchs, diags)
	o.Generators = utils.ListValueOrNull(ctx, types.ObjectType{AttrTypes: configletGenerator{}.attrTypes()}, configletGenerators, diags)
}
