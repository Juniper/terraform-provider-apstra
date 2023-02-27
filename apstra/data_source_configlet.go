package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceConfiglet{}

type dataSourceConfiglet struct {
	client *goapstra.Client
}

func (o *dataSourceConfiglet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configlet"
}

func (o *dataSourceConfiglet) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = dataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceConfiglet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific Configlet.\n\n" +
			"At least one optional attribute is required. ",
		Attributes: configlet{}.dataSourceAttributes(),
	}
}

func (o *dataSourceConfiglet) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config configlet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *goapstra.Configlet
	var ace goapstra.ApstraClientErr

	switch {
	case !config.Name.IsNull():
		api, err = o.client.GetConfigletByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Configlet not found",
				fmt.Sprintf("Configlet with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		api, err = o.client.GetConfiglet(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Configlet not found",
				fmt.Sprintf("Configlet with ID %q not found", config.Id.ValueString()))
			return
		}
	default:
		resp.Diagnostics.AddError(errInsufficientConfigElements, "neither 'name' nor 'id' set")
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Configlet", err.Error())
		return
	}

	// create new state object
	var state configlet
	state.Id = types.StringValue(string(api.Id))
	state.loadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
