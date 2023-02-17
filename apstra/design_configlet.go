package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dConfiglet struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RefArch    types.Set    `tfsdk:"ref_archs"`
	Generators types.List   `tfsdk:"generators"`
}

func (o *dConfiglet) loadApiResponse(ctx context.Context, in *goapstra.Configlet, diags *diag.Diagnostics) {
	var d diag.Diagnostics

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)

	refArchs := make([]string, len(in.Data.RefArchs))
	for i, refArch := range in.Data.RefArchs {
		refArchs[i] = refArch.String()
	}
	o.RefArch, d = types.SetValueFrom(ctx, types.StringType, refArchs)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	configletGenerators := make([]dConfigletGenerator, len(in.Data.Generators))
	for i, configletGenerator := range in.Data.Generators {
		configletGenerators[i].loadApiResponse(ctx, &configletGenerator, diags)
		if diags.HasError() {
			return
		}
	}

	o.Generators, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: dConfigletGenerator{}.attrTypes()}, configletGenerators)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
}

type dConfigletGenerator struct {
	ConfigStyle          types.String `tfsdk:"config_style"`
	Section              types.String `tfsdk:"section"`
	TemplateText         types.String `tfsdk:"template_text"`
	NegationTemplateText types.String `tfsdk:"negation_template_text"`
	FileName             types.String `tfsdk:"filename"`
}

func (o dConfigletGenerator) attributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"config_style": datasourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(""),
			Computed:            true,
		},
		"section": datasourceSchema.StringAttribute{
			MarkdownDescription: "Config Section",
			Computed:            true,
		},
		"template_text": datasourceSchema.StringAttribute{
			MarkdownDescription: "Template Text",
			Computed:            true,
		},
		"negation_template_text": datasourceSchema.StringAttribute{
			MarkdownDescription: "Negation Template Text",
			Computed:            true,
		},
		"filename": datasourceSchema.StringAttribute{
			MarkdownDescription: "FileName",
			Computed:            true,
		},
	}
}

func (o dConfigletGenerator) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"config_style":           types.StringType,
		"section":                types.StringType,
		"template_text":          types.StringType,
		"negation_template_text": types.StringType,
		"filename":               types.StringType,
	}
}

func (o *dConfigletGenerator) loadApiResponse(_ context.Context, in *goapstra.ConfigletGenerator, _ *diag.Diagnostics) {
	o.ConfigStyle = types.StringValue(in.ConfigStyle.String())
	o.Section = types.StringValue(in.Section.String())
	o.TemplateText = types.StringValue(in.TemplateText)

	if in.NegationTemplateText == "" {
		o.NegationTemplateText = types.StringNull()
	} else {
		o.NegationTemplateText = types.StringValue(in.NegationTemplateText)
	}

	if in.Filename == "" {
		o.FileName = types.StringNull()
	} else {
		o.FileName = types.StringValue(in.Filename)
	}
}
