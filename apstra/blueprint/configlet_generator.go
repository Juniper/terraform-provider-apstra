package blueprint

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConfigletGenerator struct {
	ConfigStyle          types.String `tfsdk:"config_style"`
	Section              types.String `tfsdk:"section"`
	SectionCondition     types.String `tfsdk:"section_condition"`
	TemplateText         types.String `tfsdk:"template_text"`
	NegationTemplateText types.String `tfsdk:"negation_template_text"`
	FileName             types.String `tfsdk:"filename"`
}

func (o ConfigletGenerator) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"config_style": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Indicates Platform Specific Configuration Style",
			Computed:            true,
		},
		"section": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Config Section",
			Computed:            true,
		},
		"section_condition": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Section Condition",
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

func (o ConfigletGenerator) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"config_style": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Specifies the switch platform, must be one of '%s'.",
				strings.Join(utils.AllPlatformOSNames(), "', '")),
			Required:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllPlatformOSNames()...)},
		},
		"section": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Specifies where in the target device the configlet should be "+
				"applied. Varies by network OS:\n\n%s", utils.ValidSectionsAsTable()),
			Required:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllConfigletSectionNames()...)},
		},
		"section_condition": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Used to select interfaces for configlets used in sections "+
				"`%s`, `%s` and `%s`. See references to *Advanced Condition Editor* in the [Apstra User Guide]"+
				"(https://www.juniper.net/documentation/us/en/software/apstra5.0/apstra-user-guide/topics/task/configlet-import-blueprint.html). "+
				"e.g. `role in [\"spine_leaf\"]`",
				utils.StringersToFriendlyString(enum.ConfigletSectionInterface),
				utils.StringersToFriendlyString(enum.ConfigletSectionSetBasedInterface),
				utils.StringersToFriendlyString(enum.ConfigletSectionDeleteBasedInterface),
			),
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionFile)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionFrr)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionOspf)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionSetBasedSystem)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionSystem)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionSystemTop)),
				),
			},
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
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.RegexMatches(regexp.MustCompile("^/etc/"), "Only files in /etc/ are supported for configlets"),

				// required by section file
				apstravalidator.RequiredWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionFile)),
				),

				// incompatible with sections other than file
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionDeleteBasedInterface)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionInterface)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionFrr)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionOspf)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionSetBasedInterface)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionSetBasedSystem)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionSystem)),
				),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("section"),
					types.StringValue(utils.StringersToFriendlyString(enum.ConfigletSectionSystemTop)),
				),
			},
		},
	}
}

func (o ConfigletGenerator) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"config_style":           types.StringType,
		"section":                types.StringType,
		"section_condition":      types.StringType,
		"template_text":          types.StringType,
		"negation_template_text": types.StringType,
		"filename":               types.StringType,
	}
}

func (o *ConfigletGenerator) LoadApiData(ctx context.Context, in *apstra.ConfigletGenerator, diags *diag.Diagnostics) {
	o.ConfigStyle = types.StringValue(utils.StringersToFriendlyString(in.ConfigStyle))
	o.Section = types.StringValue(utils.StringersToFriendlyString(in.Section, in.ConfigStyle))
	o.SectionCondition = utils.StringValueOrNull(ctx, in.SectionCondition, diags)
	o.TemplateText = types.StringValue(in.TemplateText)
	o.NegationTemplateText = utils.StringValueOrNull(ctx, in.NegationTemplateText, diags)
	o.FileName = utils.StringValueOrNull(ctx, in.Filename, diags)
}

func (o *ConfigletGenerator) Request(_ context.Context, diags *diag.Diagnostics) *apstra.ConfigletGenerator {
	var err error

	var configStyle enum.ConfigletStyle
	err = configStyle.FromString(o.ConfigStyle.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing configlet config_style %q", o.ConfigStyle.ValueString()), err.Error())
	}

	var section enum.ConfigletSection
	err = utils.ApiStringerFromFriendlyString(&section, o.Section.ValueString(), o.ConfigStyle.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing configlet section %q", o.Section.ValueString()), err.Error())
	}

	return &apstra.ConfigletGenerator{
		ConfigStyle:          configStyle,
		Section:              section,
		SectionCondition:     o.SectionCondition.ValueString(),
		TemplateText:         o.TemplateText.ValueString(),
		NegationTemplateText: o.NegationTemplateText.ValueString(),
		Filename:             o.FileName.ValueString(),
	}
}
