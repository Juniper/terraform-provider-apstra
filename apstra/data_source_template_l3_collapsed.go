package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceTag{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceTag{}

type dataSourceTemplateL3Collapsed struct {
	client *goapstra.Client
}

func (o *dataSourceTemplateL3Collapsed) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_l3_collapsed"
}

func (o *dataSourceTemplateL3Collapsed) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceTemplateL3Collapsed) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides details of a specific L3" +
			"collapsed template from the Apstra design API.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one template. " +
			"Matching zero templates or more than one template will produce an error.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Template id. Required when the template `name` is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"name": {
				MarkdownDescription: "Template name. Required when template `id` is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"mesh_link_count": {
				MarkdownDescription: "Count of links between any two leaf switches in the template.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"mesh_link_speed": {
				MarkdownDescription: "Speed of links between leaf switches in the template",
				Computed:            true,
				Type:                types.StringType,
			},
			"rack_type": {
				MarkdownDescription: "Details Rack Types used in this template.",
				Computed:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Rack Type name displayed in the Apstra web UI.",
						Computed:            true,
						Type:                types.StringType,
					},
					"description": {
						MarkdownDescription: "Rack Type description displayed in the Apstra web UI.",
						Computed:            true,
						Type:                types.StringType,
					},
				}),
			},
		},
	}, nil
}

func (o *dataSourceTemplateL3Collapsed) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dTemplateL3Collapsed
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
		resp.Diagnostics.AddError("configuration error", "exactly one of 'id' and 'name' must be specified")
		return
	}
}

func (o *dataSourceTemplateL3Collapsed) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dTemplateL3Collapsed
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var ace goapstra.ApstraClientErr
	var templateId goapstra.ObjectId
	var templateType goapstra.TemplateType

	// maybe the config gave us the template name?
	if !config.Id.IsNull() { // fetch template by id
		templateId = goapstra.ObjectId(config.Id.ValueString())
		templateType, err = o.client.GetTemplateType(ctx, templateId)
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(path.Root("id"), "Not found", err.Error())
			return
		}
	}

	if !config.Name.IsNull() { // fetch template by name
		templateId, templateType, err = o.client.GetTemplateIdTypeByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(path.Root("name"), "Not found", err.Error())
			return
		}
	}

	if templateType != goapstra.TemplateTypeL3Collapsed {
		resp.Diagnostics.AddError("wrong template type",
			fmt.Sprintf("selected template has type '%s', not '%s'", templateType, goapstra.TemplateTypeL3Collapsed.String()))
		return
	}

	template, err := o.client.GetL3CollapsedTemplate(ctx, templateId)
	if err != nil {
		resp.Diagnostics.AddError("error fetching template", err.Error())
		return
	}

	var state dTemplateL3Collapsed
	state.parseApi(template, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

type dTemplateL3Collapsed struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	MeshLinkCount types.Int64  `tfsdk:"mesh_link_count"`
	MeshLinkSpeed types.String `tfsdk:"mesh_link_speed"`
	RackType      types.Object `tfsdk:"rack_type"`
}

func (o *dTemplateL3Collapsed) parseApi(in *goapstra.TemplateL3Collapsed, diags *diag.Diagnostics) {
	var d diag.Diagnostics
	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.MeshLinkCount = types.Int64Value(int64(in.Data.MeshLinkCount))
	o.MeshLinkSpeed = types.StringValue(string(in.Data.MeshLinkSpeed))
	o.RackType, d = types.ObjectValue(dTemplateLogicalDeviceAttrTypes(), map[string]attr.Value{
		"name":        types.StringValue(in.Data.RackTypes[0].Data.DisplayName),
		"description": types.StringValue(in.Data.RackTypes[0].Data.Description),
	})
	diags.Append(d...)
}

func dTemplateLogicalDeviceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
	}
}
