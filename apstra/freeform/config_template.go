package freeform

import (
	"context"
	"regexp"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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

type ConfigTemplate struct {
	Id          types.String `tfsdk:"id"`
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Name        types.String `tfsdk:"name"`
	Text        types.String `tfsdk:"text"`
	Tags        types.Set    `tfsdk:"tags"`
	AssignedTo  types.Set    `tfsdk:"assigned_to"`
}

func (o ConfigTemplate) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Config Template lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Config Template by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up an imported Config Template by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"text": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Configuration Jinja2 template text",
			Computed:            true,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"assigned_to": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of System IDs to which the ConfigTemplate is assigned",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

func (o ConfigTemplate) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Config Template.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Config Template name as shown in the Web UI. Must end with `.jinja`.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(7),
				stringvalidator.RegexMatches(regexp.MustCompile(".jinja$"), "must end with '.jinja'"),
			},
		},
		"text": resourceSchema.StringAttribute{
			MarkdownDescription: "Configuration Jinja2 template text",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"assigned_to": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of System IDs to which the ConfigTemplate is assigned",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *ConfigTemplate) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ConfigTemplateData {
	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil
	}

	return &apstra.ConfigTemplateData{
		Label: o.Name.ValueString(),
		Text:  o.Text.ValueString(),
		Tags:  tags,
	}
}

func (o *ConfigTemplate) LoadApiData(ctx context.Context, in *apstra.ConfigTemplateData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Text = types.StringValue(in.Text)
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags) // safe to ignore diagnostic here
}

func (o ConfigTemplate) NeedsUpdate(state ConfigTemplate) bool {
	switch {
	case !o.Name.Equal(state.Name):
		return true
	case !o.Text.Equal(state.Text):
		return true
	case !o.Tags.Equal(state.Tags):
		return true
	}

	return false
}
