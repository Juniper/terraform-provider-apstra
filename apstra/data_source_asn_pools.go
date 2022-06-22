package apstra

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceAsnPoolsType struct{}

func (r dataSourceAsnPoolsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"ids": {
				Computed: true,
				Type:     types.SetType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (r dataSourceAsnPoolsType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceAsnPools{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceAsnPools struct {
	p provider
}

func (r dataSourceAsnPools) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
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
