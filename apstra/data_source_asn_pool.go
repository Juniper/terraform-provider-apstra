package apstra

import (
	"context"
	"fmt"
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceAsnPoolType struct{}

func (r dataSourceAsnPoolType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Required: true,
				Type:     types.StringType,
			},
			"name": {
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
			"created_at": {
				Computed: true,
				Type:     types.StringType,
			},
			"last_modified_at": {
				Computed: true,
				Type:     types.StringType,
			},
			"total": {
				Computed: true,
				Type:     types.Int64Type,
			},
			"ranges": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"status": {
						Type:     types.StringType,
						Computed: true,
					},
					"first": {
						Type:     types.Int64Type,
						Computed: true,
					},
					"last": {
						Type:     types.Int64Type,
						Computed: true,
					},
					"total": {
						Type:     types.Int64Type,
						Computed: true,
					},
					"used": {
						Type:     types.Int64Type,
						Computed: true,
					},
					"used_percentage": {
						Type:     types.Float64Type,
						Computed: true,
					},
				}),
			},
		},
	}, nil
}

func (r dataSourceAsnPoolType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceAsnPool{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceAsnPool struct {
	p provider
}

func (r dataSourceAsnPool) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var config DataAsnPool
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Id.IsNull() || config.Id.IsUnknown() {
		resp.Diagnostics.AddError(
			"pool id must be known and not null",
			fmt.Sprintf("pool id known: %t; pool id null: %t", config.Id.IsUnknown(), config.Id.IsNull()),
		)
	}

	asnPool, err := r.p.client.GetAsnPool(ctx, goapstra.ObjectId(config.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving ASN pool",
			fmt.Sprintf("error retrieving ASN pool '%s' - %s", config.Id.Value, err),
		)
		return
	}

	// convert pool tags from []string to []types.String
	tags := sliceStringToSliceTfString(asnPool.Tags)

	// convert pool ranges goapstra.AsnRange to AsnRange
	var asnRanges []AsnRange
	for _, r := range asnPool.Ranges {
		asnRanges = append(asnRanges, AsnRange{
			Status:         types.String{Value: r.Status},
			First:          types.Int64{Value: int64(r.First)},
			Last:           types.Int64{Value: int64(r.Last)},
			Total:          types.Int64{Value: int64(r.Total)},
			Used:           types.Int64{Value: int64(r.Used)},
			UsedPercentage: types.Float64{Value: float64(r.UsedPercentage)},
		})
	}

	// Set state
	diags = resp.State.Set(ctx, &DataAsnPool{
		Id:             types.String{Value: string(asnPool.Id)},
		Name:           types.String{Value: asnPool.DisplayName},
		Status:         types.String{Value: asnPool.Status},
		Tags:           tags,
		Used:           types.Int64{Value: int64(asnPool.Used)},
		UsedPercent:    types.Float64{Value: float64(asnPool.UsedPercentage)},
		CreatedAt:      types.String{Value: asnPool.CreatedAt.String()},
		LastModifiedAt: types.String{Value: asnPool.LastModifiedAt.String()},
		Total:          types.Int64{Value: int64(asnPool.Total)},
		Ranges:         asnRanges,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
