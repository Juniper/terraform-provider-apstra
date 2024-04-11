package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceTemplateCollapsed{}
var _ datasourceWithSetClient = &dataSourceTemplateCollapsed{}

type dataSourceTemplateCollapsed struct {
	client *apstra.Client
}

func (o *dataSourceTemplateCollapsed) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_collapsed"
}

func (o *dataSourceTemplateCollapsed) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceTemplateCollapsed) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This data source provides details of a specific Collapsed Fabric Template.\n\n" +
			"At least one optional attribute is required.",
		Attributes: design.TemplateCollapsed{}.DataSourceAttributes(),
	}
}

func (o *dataSourceTemplateCollapsed) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config design.TemplateCollapsed
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.TemplateL3Collapsed
	var ace apstra.ClientErr

	switch { // attribute validation ensures that one of Name and ID must be set.
	case !config.Name.IsNull():
		api, err = o.client.GetL3CollapsedTemplateByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) {
			switch ace.Type() {
			case apstra.ErrNotfound:
				resp.Diagnostics.AddError(
					"Collapsed Template not found",
					fmt.Sprintf("Collapsed Template with name %q does not exist", config.Name.ValueString()))
				return
			case apstra.ErrWrongType:
				resp.Diagnostics.AddError(fmt.Sprintf("Specified Template has wrong type: %s", api.Type()), err.Error())
				return
			}
		}
	case !config.Id.IsNull():
		api, err = o.client.GetL3CollapsedTemplate(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) {
			switch ace.Type() {
			case apstra.ErrNotfound:
				resp.Diagnostics.AddError(
					"Collapsed Template not found",
					fmt.Sprintf("Collapsed Template with ID %q does not exist", config.Id.ValueString()))
				return
			case apstra.ErrWrongType:
				resp.Diagnostics.AddError("Specified Template has wrong type", err.Error())
				return
			}
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Collapsed Template query error", err.Error())
		return
	}

	// create state object
	var state design.TemplateCollapsed
	state.Id = types.StringValue(string(api.Id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceTemplateCollapsed) setClient(client *apstra.Client) {
	o.client = client
}
