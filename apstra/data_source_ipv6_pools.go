package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceIpv6Pools{}
var _ datasourceWithSetClient = &dataSourceIpv6Pools{}

type dataSourceIpv6Pools struct {
	client *apstra.Client
}

func (o *dataSourceIpv6Pools) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv6_pools"
}

func (o *dataSourceIpv6Pools) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceIpv6Pools) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryResources + "This data source returns the ID numbers of all IPv6 Pools",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceIpv6Pools) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	ids, err := o.client.ListIp6PoolIds(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving IPv6 Pool IDs", err.Error())
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

func (o *dataSourceIpv6Pools) setClient(client *apstra.Client) {
	o.client = client
}
