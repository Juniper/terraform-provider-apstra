package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ datasource.DataSourceWithConfigure = &dataSourceBlueprintSystemNodes{}

type dataSourceBlueprintSystemNodes struct {
	client *apstra.Client
}

func (o *dataSourceBlueprintSystemNodes) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint_system_nodes"
}

func (o *dataSourceBlueprintSystemNodes) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceBlueprintSystemNodes) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns Graph DB node IDs of *system* nodes within a Blueprint.\n\n" +
			"Optional attributes filter the result list so that it only contains IDs of nodes which match the filters.",
		Attributes: blueprint.Systems{}.DataSourceAttributes(),
	}
}

func (o *dataSourceBlueprintSystemNodes) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config blueprint.Systems
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var queryResponse struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	query := config.Query(ctx, &resp.Diagnostics).
		SetClient(o.client).
		SetBlueprintId(apstra.ObjectId(config.BlueprintId.ValueString())).
		SetBlueprintType(apstra.BlueprintTypeStaging)
	if resp.Diagnostics.HasError() { // catch errors fro
		return
	}

	err := query.Do(ctx, &queryResponse)
	if err != nil {
		resp.Diagnostics.AddError("Error executing Blueprint query", err.Error())
		return
	}

	ids := make([]attr.Value, len(queryResponse.Items))
	for i := range queryResponse.Items {
		ids[i] = types.StringValue(queryResponse.Items[i].System.Id)
	}
	config.Ids = types.SetValueMust(types.StringType, ids)
	config.QueryString = types.StringValue(query.String())

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
