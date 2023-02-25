package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
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

type tag struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (o tag) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Tag by ID. Required when `name`is omitted.",
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
			MarkdownDescription: "Populate this field to look up a Tag by name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The description of the returned Tag.",
			Computed:            true,
		},
	}
}

func (o tag) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Tag.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag name field as seen in the web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()}, // {"errors":{"label":"Tag label cannot be changed"}}

		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag description field as seen in the web UI.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o tag) resourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in nested contexts.",
			Computed:            true,
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag name field as seen in the web UI.",
			Computed:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag description field as seen in the web UI.",
			Computed:            true,
		},
	}
}

func (o *tag) loadApiData(_ context.Context, in *goapstra.DesignTagData, _ *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Description = types.StringValue(in.Description)
}

func (o *tag) request(_ context.Context, _ *diag.Diagnostics) *goapstra.DesignTagRequest {
	return &goapstra.DesignTagRequest{
		Label:       o.Name.ValueString(),
		Description: o.Description.ValueString(),
	}
}
