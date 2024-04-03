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

var _ datasource.DataSourceWithConfigure = &dataSourceTemplateRackBased{}
var _ datasourceWithSetClient = &dataSourceTemplateRackBased{}

type dataSourceTemplateRackBased struct {
	client *apstra.Client
}

func (o *dataSourceTemplateRackBased) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_rack_based"
}

func (o *dataSourceTemplateRackBased) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceTemplateRackBased) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This data source provides details of a specific Rack Based (3 stage) Template.\n\n" +
			"At least one optional attribute is required.",
		Attributes: design.TemplateRackBased{}.DataSourceAttributes(),
	}
}

func (o *dataSourceTemplateRackBased) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config design.TemplateRackBased
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.TemplateRackBased
	var ace apstra.ClientErr

	switch { // attribute validation ensures that one of Name and ID must be set.
	case !config.Name.IsNull():
		api, err = o.client.GetRackBasedTemplateByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) {
			switch ace.Type() {
			case apstra.ErrNotfound:
				resp.Diagnostics.AddError(
					"Rack Based Template not found",
					fmt.Sprintf("Rack Based Template with name %q does not exist", config.Name.ValueString()))
				return
			case apstra.ErrWrongType:
				resp.Diagnostics.AddError(fmt.Sprintf("Specified Template has wrong type: %s", api.Type()), err.Error())
				return
			}
		}
	case !config.Id.IsNull():
		api, err = o.client.GetRackBasedTemplate(ctx, apstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) {
			switch ace.Type() {
			case apstra.ErrNotfound:
				resp.Diagnostics.AddError(
					"Rack Based Template not found",
					fmt.Sprintf("Rack Based Template with ID %q does not exist", config.Id.ValueString()))
				return
			case apstra.ErrWrongType:
				resp.Diagnostics.AddError("Specified Template has wrong type", err.Error())
				return
			}
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Rack Based Template query error", err.Error())
		return
	}

	// create state object
	var state design.TemplateRackBased
	state.Id = types.StringValue(string(api.Id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceTemplateRackBased) setClient(client *apstra.Client) {
	o.client = client
}
