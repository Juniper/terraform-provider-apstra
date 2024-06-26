package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprintDeploy{}
var _ datasourceWithSetClient = &dataSourceBlueprintDeploy{}

type dataSourceBlueprintDeploy struct {
	client *apstra.Client
}

func (o *dataSourceBlueprintDeploy) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_deployment"
}

func (o *dataSourceBlueprintDeploy) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceBlueprintDeploy) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This data source returns the deployment status of a Blueprint.",
		Attributes:          blueprint.Deploy{}.DataSourceAttributes(),
	}
}

func (o *dataSourceBlueprintDeploy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config blueprint.Deploy
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//config.Comment = types.StringUnknown()
	config.Read(ctx, o.client, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceBlueprintDeploy) setClient(client *apstra.Client) {
	o.client = client
}
