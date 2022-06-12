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
						Computed: true,
						Type:     types.StringType,
					},
					"display_name": {
						Computed: true,
						Type:     types.StringType,
					},
					"status": {
						Computed: true,
						Type:     types.StringType,
					},
					"tags": {
						Computed: true,
						Type:     types.ListType{ElemType: types.StringType},
					},
					"used": {
						Computed: true,
						Type:     types.Int64Type,
					},
					"used_percentage": {
						Computed: true,
						Type:     types.Float64Type,
					},
					"created": {
						Computed: true,
						Type:     types.StringType,
					},
					"modified": {
						Computed: true,
						Type:     types.StringType,
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
			fmt.Sprintf("error retrieving ASN pools - %s", err),
		)
		return
	}

	// map response body to resource schema
	for _, asnPool := range asnPools {
		// convert tags from []string to []types.String
		var tags []types.String
		for _, t := range asnPool.Tags {
			tags = append(tags, types.String{Value: t})
		}

		resourceState.AsnPools = append(resourceState.AsnPools, AsnPool{
			Id:          types.String{Value: string(asnPool.Id)},
			DisplayName: types.String{Value: asnPool.DisplayName},
			Status:      types.String{Value: asnPool.Status},
			Tags:        tags,
			Used:        types.Int64{Value: int64(asnPool.Used)},
			UsedPercent: types.Float64{Value: float64(asnPool.UsedPercentage)},
			Created:     types.String{Value: asnPool.CreatedAt.String()},
			Modified:    types.String{Value: asnPool.LastModifiedAt.String()},
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
