package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

type Configlet struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RefArchs   types.Set    `tfsdk:"ref_archs"`
	Generators types.List   `tfsdk:"generators"`
}

func (o Configlet) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
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

func (o Configlet) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Configlet by ID. Required when `name`is omitted.",
			Computed:            true,
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Configlet by name. Required when `id`is omitted.",
			Required:            true,
		},
		"ref_archs": resourceSchema.SetAttribute{
			MarkdownDescription: "List of architectures",
			Required:            true,
			ElementType:         types.StringType,
		},
		"generators": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Generators organized by Network OS",
			Required:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: configletGenerator{}.resourceAttributes(),
			},
		},
	}
}

func (o *Configlet) Request(ctx context.Context, diags *diag.Diagnostics) *goapstra.ConfigletRequest {
	var tf_gen []configletGenerator
	var r *goapstra.ConfigletRequest = &goapstra.ConfigletRequest{}

	diags.Append(o.Generators.ElementsAs(ctx, &tf_gen, true)...)
	r.RefArchs = make([]goapstra.RefDesign, len(o.RefArchs.Elements()))
	refArches := make([]string, len(o.RefArchs.Elements()))
	d := o.RefArchs.ElementsAs(ctx, &refArches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	for i, j := range refArches {
		e := r.RefArchs[i].FromString(j)
		if e != nil {
			diags.AddError(fmt.Sprintf("error parsing reference architecture : %q", j), e.Error())
		}
	}
	r.Generators = make([]goapstra.ConfigletGenerator, len(o.Generators.Elements()))
	dCG := make([]configletGenerator, len(o.Generators.Elements()))
	d = o.Generators.ElementsAs(ctx, &dCG, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	for i, j := range dCG {
		var a goapstra.ApstraPlatformOS
		e := a.FromString(j.ConfigStyle.ValueString())
		if e != nil {
			diags.AddError(fmt.Sprintf("error parsing configlet style : '%s'", j.ConfigStyle.ValueString()), e.Error())
		}
		var s goapstra.ApstraConfigletSection

		e = s.FromString(j.Section.ValueString())
		if e != nil {
			diags.AddError(fmt.Sprintf("error parsing configlet section : '%s'", j.Section.ValueString()), e.Error())
		}
		r.Generators[i] = goapstra.ConfigletGenerator{
			ConfigStyle:          a,
			Section:              s,
			TemplateText:         j.TemplateText.ValueString(),
			NegationTemplateText: j.NegationTemplateText.ValueString(),
			Filename:             j.FileName.ValueString(),
		}
	}

	var d diag.Diagnostics

	refArchStrings := make([]string, len(o.RefArchs.Elements()))
	d = o.RefArchs.ElementsAs(ctx, &refArchStrings, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	refArchs := make([]goapstra.RefDesign, len(refArchStrings))
	for i, s := range refArchStrings {
		err := refArchs[i].FromString(s)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing reference architecture %q", s), err.Error())
		}
	}

	generators := make([]goapstra.ConfigletGenerator, len(o.Generators.Elements()))
	d = o.RefArchs.ElementsAs(ctx, &generators, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	genRequests := make()

	return &goapstra.ConfigletRequest{
		DisplayName: o.Name.ValueString(),
		RefArchs:    refArchs,
		Generators:  nil, // todo
	}
}

func (o *Configlet) loadApiData(ctx context.Context, in *goapstra.ConfigletData, diags *diag.Diagnostics) {
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
