package design

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
	"terraform-provider-apstra/apstra/utils"
)

type ConfigletGenerator struct {
	ConfigStyle          types.String `tfsdk:"config_style"`
	Section              types.String `tfsdk:"section"`
	TemplateText         types.String `tfsdk:"template_text"`
	NegationTemplateText types.String `tfsdk:"negation_template_text"`
	FileName             types.String `tfsdk:"filename"`
}

func (o ConfigletGenerator) DataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"config_style": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Indicates Platform Specific Configuration Style"),
			Computed:            true,
		},
		"section": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Config Section",
			Computed:            true,
		},
		"template_text": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Template Text",
			Computed:            true,
		},
		"negation_template_text": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Negation Template Text",
			Computed:            true,
		},
		"filename": dataSourceSchema.StringAttribute{
			MarkdownDescription: "FileName",
			Computed:            true,
		},
	}
}

func (o ConfigletGenerator) ResourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"config_style": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Specifies the switch platform, must be one of '%s'.",
				strings.Join(utils.AllPlatformOSNames(), "', '")),
			Required:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllPlatformOSNames()...)}},
		"section": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Specifies where in the target device the configlet"+
				"should be applied. Must be one of '%s", strings.Join(utils.AllConfigletSectionNames(), "', '")),
			Required:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllConfigletSectionNames()...)},
		},
		"template_text": resourceSchema.StringAttribute{
			MarkdownDescription: "Template Text",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"negation_template_text": resourceSchema.StringAttribute{
			MarkdownDescription: "Negation Template Text",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"filename": resourceSchema.StringAttribute{
			MarkdownDescription: "FileName",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o ConfigletGenerator) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"config_style":           types.StringType,
		"section":                types.StringType,
		"template_text":          types.StringType,
		"negation_template_text": types.StringType,
		"filename":               types.StringType,
	}
}

func (o *ConfigletGenerator) LoadApiData(ctx context.Context, in *goapstra.ConfigletGenerator, diags *diag.Diagnostics) {
	o.ConfigStyle = types.StringValue(in.ConfigStyle.String())
	o.Section = types.StringValue(utils.StringersToFriendlyString(in.Section, in.ConfigStyle))
	o.TemplateText = types.StringValue(in.TemplateText)
	o.NegationTemplateText = utils.StringValueOrNull(ctx, in.NegationTemplateText, diags)
	o.FileName = utils.StringValueOrNull(ctx, in.Filename, diags)
}

func (o *ConfigletGenerator) Request(_ context.Context, diags *diag.Diagnostics) *goapstra.ConfigletGenerator {
	var err error

	var configStyle goapstra.PlatformOS
	err = configStyle.FromString(o.ConfigStyle.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing configlet config_style %q", o.ConfigStyle.ValueString()), err.Error())
	}

	var section goapstra.ConfigletSection
	//err = section.FromString(o.Section.ValueString())
	err = utils.FriendlyStringToAPIStringer(&section, o.Section.ValueString(), o.ConfigStyle.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing configlet section %q", o.Section.ValueString()), err.Error())
	}

	return &goapstra.ConfigletGenerator{
		ConfigStyle:          configStyle,
		Section:              section,
		TemplateText:         o.TemplateText.ValueString(),
		NegationTemplateText: o.NegationTemplateText.ValueString(),
		Filename:             o.FileName.ValueString(),
	}
}
