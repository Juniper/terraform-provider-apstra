package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceConfiglet{}
var _ datasourceWithSetClient = &dataSourceConfiglet{}

type dataSourceConfiglet struct {
	client *apstra.Client
}

func (o *dataSourceConfiglet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configlet"
}

func (o *dataSourceConfiglet) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceConfiglet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This data source provides details of a specific Configlet.\n\n" +
			"At least one optional attribute is required.",
		Attributes: design.Configlet{}.DataSourceAttributes(),
	}
}

func (o *dataSourceConfiglet) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config design.Configlet
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.Configlet

	switch {
	case !config.Name.IsNull():
		api, err = o.client.GetConfigletByName(ctx, config.Name.ValueString())
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Configlet not found",
				fmt.Sprintf("Configlet with name %q not found", config.Name.ValueString()))
			return
		}
	case !config.Id.IsNull():
		api, err = o.client.GetConfiglet(ctx, apstra.ObjectId(config.Id.ValueString()))
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Configlet not found",
				fmt.Sprintf("Configlet with ID %q not found", config.Id.ValueString()))
			return
		}
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Configlet", err.Error())
		return
	}

	// create new state object
	var state design.Configlet
	state.Id = types.StringValue(string(api.Id))
	state.LoadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *dataSourceConfiglet) setClient(client *apstra.Client) {
	o.client = client
}
