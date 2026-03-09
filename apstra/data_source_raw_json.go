package tfapstra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceRawJSON{}
var _ datasourceWithSetClient = &dataSourceRawJSON{}

type dataSourceRawJSON struct {
	client *apstra.Client
}

func (o *dataSourceRawJSON) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_raw_json"
}

func (o *dataSourceRawJSON) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceRawJSON) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFootGun + "**!!! Warning !!!**\n" +
			"This is resource is intended only to solve problems not addressed by the normal data sources. " +
			"Its use is discouraged and not supported. You're on your own with this thing.\n" +
			"**!!! Warning !!!**\n\n" +
			"This data source retrieves data as raw JSON via `GET` request",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The API URL associated with the raw JSON object.",
				Required:            true,
			},
			"result": schema.StringAttribute{
				CustomType:          jsontypes.ExactType{},
				MarkdownDescription: "The raw JSON returned by the Apstra API.",
				Computed:            true,
			},
		},
	}
}

func (o *dataSourceRawJSON) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		URL    types.String    `tfsdk:"url"`
		Result jsontypes.Exact `tfsdk:"result"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := url.Parse(config.URL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse URL", fmt.Sprintf("While parsing URL %q is not valid: %v", config.URL.ValueString(), err.Error()))
		return
	}

	var rawJSONResponse json.RawMessage
	if err = o.client.DoRawJsonTransaction(ctx, apstra.RawJsonRequest{
		Method: http.MethodGet,
		Url:    u,
	}, &rawJSONResponse); err != nil {
		resp.Diagnostics.AddError("Error retrieving raw JSON data", err.Error())
		return
	}

	if rawJSONResponse == nil {
		config.Result = jsontypes.NewExactNull()
	} else {
		config.Result = jsontypes.NewExactValue(string(rawJSONResponse))
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceRawJSON) setClient(client *apstra.Client) {
	o.client = client
}
