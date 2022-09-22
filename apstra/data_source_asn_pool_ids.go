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

type dataSourceAsnPoolsType struct{}

func (r dataSourceAsnPools) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "apstra_asn_pool_ids"
}

func (r dataSourceAsnPools) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source returns the IDs of all ASN resource pools",
		Attributes: map[string]tfsdk.Attribute{
			"ids": {
				MarkdownDescription: "The IDs of all ASN resource pools.",
				Computed:            true,
				Type:                types.SetType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (r dataSourceAsnPoolsType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceAsnPools{
		p: *(p.(*Provider)),
	}, nil
}

type dataSourceAsnPools struct {
	p Provider
}

func (r dataSourceAsnPools) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	poolIds, err := r.p.client.ListAsnPoolIds(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving ASN pool IDs",
			fmt.Sprintf("error retrieving ASN pool IDs - %s", err),
		)
		return
	}

	// map response body to resource schema
	var state DataAsnPoolIds
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
