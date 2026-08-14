package design

import (
	"context"

	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagDefinition struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (td tagDefinition) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
	}
}

func (td tagDefinition) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Tag name",
			Computed:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Tag description",
			Computed:            true,
		},
	}
}

func (td tagDefinition) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag name",
			Computed:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag description",
			Computed:            true,
		},
	}
}

func (td tagDefinition) resourceAttributesEmbedded() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag name",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 64),
				stringvalidator.RegexMatches(apstraregexp.NoLeadingOrTrailingWhitespace, apstraregexp.NoLeadingOrTrailingWhitespaceMsg),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag description",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (td tagDefinition) asObject(ctx context.Context, diags *diag.Diagnostics) types.Object {
	result, d := types.ObjectValueFrom(ctx, td.attrTypes(), td)
	diags.Append(d...)
	return result
}
