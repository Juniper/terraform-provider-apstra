package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterSystemNodes{}
var _ datasourceWithSetClient= &dataSourceDatacenterSystemNodes{}

type dataSourceDatacenterSystemNodes struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterSystemNodes) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_systems"
}

func (o *dataSourceDatacenterSystemNodes) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterSystemNodes) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns Graph DB node IDs of *system* nodes within a Blueprint.\n\n" +
			"Optional `filters` can be used to select only interesting nodes.",
		Attributes: blueprint.NodesTypeSystem{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterSystemNodes) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.NodesTypeSystem
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ReadFromApi(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterSystemNodes) setClient(client *apstra.Client) {
	o.client = client
}
