package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAnomalies{}

type dataSourceAnomalies struct {
	client *apstra.Client
}

func (o *dataSourceAnomalies) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_anomalies"
}

func (o *dataSourceAnomalies) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceAnomalies) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides per-node summary, " +
			"per-service summary and full details of anomalies in the specified Blueprint.",
		Attributes: blueprint.Anomalies{}.DataSourceAttributes(),
	}
}

func (o *dataSourceAnomalies) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config blueprint.Anomalies
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ReadFromApi(ctx, o.client, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
