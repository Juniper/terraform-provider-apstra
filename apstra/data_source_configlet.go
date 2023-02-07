package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceConfiglet{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceConfiglet{}

type dataSourceConfiglet struct {
	client *goapstra.Client
}

func (o *dataSourceConfiglet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configlet"
}

func (o *dataSourceConfiglet) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errDataSourceConfigureProviderDataDetail,
			fmt.Sprintf(errDataSourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *dataSourceConfiglet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific configlet.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one configlet. " +
			"Matching zero configlet or more than one configlet will produce an error.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Configlet id. Required when the configlet name is omitted.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Configlet display name. Required when configlet id is omitted.",
				Optional:            true,
				Computed:            true,
			},
			"ref_archs": schema.SetAttribute{
				MarkdownDescription: "List of architectures",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"generators": schema.ListNestedAttribute{
				MarkdownDescription: "Generators organized by Network OS",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dConfigletGenerator{}.schema(),
				},
			},
		},
	}
}

func (o *dataSourceConfiglet) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dConfiglet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
		resp.Diagnostics.AddError("configuration error", "exactly one of 'id' and 'name' must be specified")
		return
	}
}

func (o *dataSourceConfiglet) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dConfiglet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var cl *goapstra.Configlet
	var ace goapstra.ApstraClientErr
	if !config.Name.IsNull() {
		cl, err = o.client.GetConfigletByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Configlet not found",
				fmt.Sprintf("Configlet with name '%s' not found", config.Name.ValueString()))
			return
		}
	}
	if !config.Id.IsNull() {
		cl, err = o.client.GetConfiglet(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Configlet not found",
				fmt.Sprintf("Configlet with id '%s' not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Configlet", err.Error())
		return
	}

	// create new state object
	d := dConfiglet{}
	d.loadApiResponse(ctx, cl, &resp.Diagnostics)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}

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

	o.Generators, d = types.ListValueFrom(ctx, dConfigletGenerator{}.attrType(), configletGenerators)
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

func (o dConfigletGenerator) schema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"config_style": schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(""),
			Computed:            true,
		},
		"section": schema.StringAttribute{
			MarkdownDescription: "Config Section",
			Computed:            true,
		},
		"template_text": schema.StringAttribute{
			MarkdownDescription: "Template Text",
			Computed:            true,
		},
		"negation_template_text": schema.StringAttribute{
			MarkdownDescription: "Negation Template Text",
			Computed:            true,
		},
		"filename": schema.StringAttribute{
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

func (o dConfigletGenerator) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
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
