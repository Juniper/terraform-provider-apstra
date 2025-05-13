package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprintAnomalies{}
var _ datasourceWithSetClient = &dataSourceBlueprintAnomalies{}

type dataSourceBlueprintAnomalies struct {
	client *apstra.Client
}

func (o *dataSourceBlueprintAnomalies) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_anomalies"
}

func (o *dataSourceBlueprintAnomalies) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceBlueprintAnomalies) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This data source provides per-node summary, " +
			"per-service summary and full details of anomalies in the specified Blueprint.",
		Attributes: blueprint.Anomalies{}.DataSourceAttributes(),
	}
}

func (o *dataSourceBlueprintAnomalies) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.Anomalies
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ReadFromApi(ctx, o.client, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceBlueprintAnomalies) setClient(client *apstra.Client) {
	o.client = client
}
