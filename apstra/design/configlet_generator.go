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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
			MarkdownDescription: fmt.Sprintf("Specifies where in the target device the configlet should be applied. valid values are '%v", utils.ValidSectionsMap()),
			Required:            true,
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

var _ validator.Object = ConfigletGeneratorValidator{}

type ConfigletGeneratorValidator struct {
}

func (o ConfigletGeneratorValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that the section name matches the config style.")
}

func (o ConfigletGeneratorValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ConfigletGeneratorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var c ConfigletGenerator
	//resp.Diagnostics.Append(req.Config.Get(ctx, &c)...)
	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &c, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	cg := c.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	valid := false
	for _, i := range cg.ConfigStyle.ValidSections() {
		if i == cg.Section {
			valid = true
			goto done
		}
	}
done:
	if !valid {
		resp.Diagnostics.AddError("Invalid Section", fmt.Sprintf("Invalid Section %q used for Config Style %q", cg.Section.String(), cg.ConfigStyle.String()))
	}
	return
}

func ValidateConfigletGenerator() validator.Object {
	return ConfigletGeneratorValidator{}
}
