package apstra

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceIp4PoolsType struct{}

func (r dataSourceIp4Pools) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "apstra_ip4_pool_ids"
}

func (r dataSourceIp4Pools) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source returns the pool IDs all IPv4 resource pools",
		Attributes: map[string]tfsdk.Attribute{
			"ids": {
				MarkdownDescription: "All pool IDs of all IPv4 resource pools.",
				Computed:            true,
				Type:                types.SetType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (r dataSourceIp4PoolsType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceIp4Pools{
		p: *(p.(*Provider)),
	}, nil
}

type dataSourceIp4Pools struct {
	p Provider
}

func (r dataSourceIp4Pools) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	poolIds, err := r.p.client.ListIp4PoolIds(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving IPv4 pool IDs",
			fmt.Sprintf("error retrieving ASN pool IDs - %s", err),
		)
		return
	}

	// map response body to resource schema
	var state DataIp4PoolIds
	for _, Id := range poolIds {
		state.Ids = append(state.Ids, types.String{Value: string(Id)})
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
