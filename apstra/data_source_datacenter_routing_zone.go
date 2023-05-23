package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ datasource.DataSourceWithConfigure = &dataSourceRoutingZone{}

type dataSourceRoutingZone struct {
	client *apstra.Client
}

func (o *dataSourceRoutingZone) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zone"
}

func (o *dataSourceRoutingZone) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceRoutingZone) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns details of a specific *security " +
			"zone* (graph datastore term for *routing zone*) node within a Datacenter Blueprint's graph DB.",
		Attributes: blueprint.NodeTypeSecurityZone{}.DataSourceAttributes(),
	}
}

func (o *dataSourceRoutingZone) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config blueprint.NodeTypeSecurityZone
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
