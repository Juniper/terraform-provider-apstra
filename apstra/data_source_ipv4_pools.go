package tfapstra

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceIpv4Pools{}

type dataSourceIpv4Pools struct {
	client *apstra.Client
}

func (o *dataSourceIpv4Pools) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv4_pools"
}

func (o *dataSourceIpv4Pools) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceIpv4Pools) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the ID numbers of all IPv4 Pools",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceIpv4Pools) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	ids, err := o.client.ListIp4PoolIds(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving IPv4 Pool IDs", err.Error())
		return
	}

	idSet, diags := types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create new state object
	var state struct {
		Ids types.Set `tfsdk:"ids"`
	}
	state.Ids = idSet

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
