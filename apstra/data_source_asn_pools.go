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
			"apstra_asn_pools": {
				// When Computed is true, the provider will set value --
				// the user cannot define the value
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Type:     types.StringType,
						Computed: true,
					},
					"display_name": {
						Type:     types.StringType,
						Computed: true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
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
	// Declare struct that this function will set to this data source's state
	var resourceState struct {
		AsnPools []AsnPool `tfsdk:"apstra_asn_pools"`
	}

	asnPools, err := r.p.client.GetAsnPools(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving ASN pools",
			err.Error(),
		)
		return
	}

	// Map response body to resource schema
	for _, asnPool := range asnPools {
		resourceState.AsnPools = append(resourceState.AsnPools, AsnPool{
			Id:          types.String{Value: string(asnPool.Id)},
			DisplayName: types.String{Value: asnPool.DisplayName},
		})
	}

	// Sample debug message
	// To view this message, set the TF_LOG environment variable to DEBUG
	// 		`export TF_LOG=DEBUG`
	// To hide debug message, unset the environment variable
	// 		`unset TF_LOG`
	fmt.Fprintf(stderr, "[DEBUG]-Resource State:%+v", resourceState)

	// Set state
	diags := resp.State.Set(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
