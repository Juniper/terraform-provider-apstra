package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAsnPools{}
var _ datasourceWithSetClient = &dataSourceAsnPools{}

type dataSourceAsnPools struct {
	client *apstra.Client
}

func (o *dataSourceAsnPools) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_pools"
}

func (o *dataSourceAsnPools) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceAsnPools) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryResources + "This data source returns the ID numbers of all ASN Pools.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceAsnPools) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	ids, err := o.client.ListAsnPoolIds(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving ASN Pool IDs", err.Error())
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

func (o *dataSourceAsnPools) setClient(client *apstra.Client) {
	o.client = client
}
