package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterSystemNode{}

type dataSourceDatacenterSystemNode struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterSystemNode) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_system"
}

func (o *dataSourceDatacenterSystemNode) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterSystemNode) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns details of a specific " +
			"*system* Graph DB node within a Blueprint.",
		Attributes: blueprint.NodeTypeSystem{}.DataSourceAttributes(),
	}
}

func (o *dataSourceDatacenterSystemNode) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.NodeTypeSystem
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
