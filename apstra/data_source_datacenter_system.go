package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterSystemNode{}
var _ datasourceWithSetClient= &dataSourceDatacenterSystemNode{}

type dataSourceDatacenterSystemNode struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterSystemNode) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_system"
}

func (o *dataSourceDatacenterSystemNode) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterSystemNode) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns details of a specific *system* Graph DB node within a Blueprint.\n\n" +
			"At least one optional attribute is required.",
		Attributes: blueprint.NodeTypeSystem{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterSystemNode) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.NodeTypeSystem
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// read "attributes" object element from the API
	config.AttributesFromApi(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Attributes.IsNull() {
		// set state and return
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// if the user supplied "id", then "name" will be null. Fill it in.
	if config.Name.IsNull() {
		config.Name = config.Attributes.Attributes()["label"].(basetypes.StringValue)
	}

	// if the user supplied "name", then "id" will be null. Fill it in.
	if config.Id.IsNull() {
		config.Id = config.Attributes.Attributes()["id"].(basetypes.StringValue)
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterSystemNode) setClient(client *apstra.Client) {
	o.client = client
}
