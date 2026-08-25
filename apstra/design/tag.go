package design

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/design"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Tag struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Definition  types.Object `tfsdk:"definition"`
}

func (t Tag) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"definition":  types.ObjectType{AttrTypes: tagDefinition{}.attrTypes()},
	}
}

func (t Tag) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Tag. Required when `name` is omitted.",
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
			MarkdownDescription: "Web UI name of the Tag. Required when `id` is omitted.",
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
		"definition": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Used in nested contexts.",
			Computed:            true,
			Attributes:          tagDefinition{}.dataSourceAttributes(),
		},
	}
}

func (t Tag) DataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in nested contexts.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Tag name.",
			Computed:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Tag description.",
			Computed:            true,
		},
		"definition": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Used in nested contexts.",
			Computed:            true,
			Attributes:          tagDefinition{}.dataSourceAttributes(),
		},
	}
}

func (t Tag) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Tag.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag name field as seen in the web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 64),
				stringvalidator.RegexMatches(apstraregexp.NoLeadingOrTrailingWhitespace, apstraregexp.NoLeadingOrTrailingWhitespaceMsg),
			},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, // {"errors":{"label":"Tag label cannot be changed"}}

		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag description field as seen in the web UI.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"definition": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Used in nested contexts.",
			Computed:            true,
			Attributes:          tagDefinition{}.resourceAttributes(),
		},
	}
}

func (t Tag) ResourceAttributesNested() map[string]resourceSchema.Attribute {
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
		"definition": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Used in nested contexts.",
			Computed:            true,
			Attributes:          tagDefinition{}.resourceAttributes(),
		},
	}
}

func (o *Tag) LoadApiDataLegacy(ctx context.Context, in *apstra.DesignTagData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Description = value.StringOrNull(ctx, in.Description, diags)
}

func (t *Tag) LoadApiData(ctx context.Context, in design.Tag, diags *diag.Diagnostics) {
	t.ID = types.StringPointerValue(in.ID())
	t.Name = types.StringValue(in.Label)
	t.Description = value.StringOrNull(ctx, in.Description, diags)
	t.Definition = t.DefinitionAsObject(ctx, diags)
}

func (t Tag) Request(_ context.Context, _ *diag.Diagnostics) design.Tag {
	var result design.Tag
	if utils.HasValue(t.ID) {
		result = design.NewTag(t.ID.ValueString())
	}

	result.Label = t.Name.ValueString()
	result.Description = t.Description.ValueString()

	return result
}

func (t Tag) DefinitionAsObject(ctx context.Context, diags *diag.Diagnostics) types.Object {
	return tagDefinition{
		Name:        t.Name,
		Description: t.Description,
	}.asObject(ctx, diags)
}

func NewTagSetLegacy(ctx context.Context, in []apstra.DesignTagData, diags *diag.Diagnostics) types.Set {
	if len(in) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: Tag{}.attrTypes()})
	}

	tags := make([]Tag, len(in))
	for i, t := range in {
		tags[i].ID = types.StringNull()
		tags[i].LoadApiDataLegacy(ctx, &t, diags)
	}

	return value.SetOrNull(ctx, types.ObjectType{AttrTypes: Tag{}.attrTypes()}, tags, diags)
}

func NewTagSet(ctx context.Context, in []design.Tag, diags *diag.Diagnostics) types.Set {
	if len(in) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: Tag{}.attrTypes()})
	}

	tags := make([]Tag, len(in))
	for i, tag := range in {
		tags[i].ID = types.StringNull()
		tags[i].LoadApiData(ctx, tag, diags)
	}

	return value.SetOrNull(ctx, types.ObjectType{AttrTypes: Tag{}.attrTypes()}, tags, diags)
}
