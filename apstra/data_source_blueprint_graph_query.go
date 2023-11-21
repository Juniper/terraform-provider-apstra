package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceDatacenterGraphQuery{}

type dataSourceDatacenterGraphQuery struct {
	client *apstra.Client
}

func (o *dataSourceDatacenterGraphQuery) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_query"
}

func (o *dataSourceDatacenterGraphQuery) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceDatacenterGraphQuery) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This data source returns the result of a specific Graph DB query within a Blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "The Blueprint ID you want to run the query against",
				Required:            true,
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "The query string in graph format.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"result": schema.StringAttribute{
				MarkdownDescription: "The result of the query",
				Computed:            true,
			},
		},
	}
}

func (o *dataSourceDatacenterGraphQuery) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Query       types.String `tfsdk:"query"`
		Result      types.String `tfsdk:"result"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// prepare and execute query
	query := new(apstra.RawQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(apstra.ObjectId(config.BlueprintId.ValueString())).
		SetClient(o.client).
		SetQuery(config.Query.ValueString())
	err := query.Do(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("failed while running graph query", err.Error())
		return
	}

	// set the result into config.Result
	config.Result = types.StringValue(string(query.RawResult()))

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
