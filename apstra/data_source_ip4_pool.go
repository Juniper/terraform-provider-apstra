package apstra

import (
	"context"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceIp4PoolType struct{}

func (r dataSourceIp4PoolType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"subnets": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"status": {
						Type:     types.StringType,
						Computed: true,
					},
					"network": {
						Type:     types.StringType,
						Required: true,
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

func (r dataSourceIp4PoolType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceIp4Pool{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceIp4Pool struct {
	p provider
}

func (r dataSourceIp4Pool) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var config DataIp4Pool
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

	ip4Pool, err := r.p.client.GetIp4Pool(ctx, goapstra.ObjectId(config.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving IPv4 pool",
			fmt.Sprintf("error retrieving IPv4 pool '%s' - %s", config.Id.Value, err),
		)
		return
	}

	// convert pool tags from []string to []types.String
	tags := sliceStringToSliceTfString(ip4Pool.Tags)

	// convert pool subnets goapstra.AsnRange to AsnRange
	var subnets []Ip4Subnet
	for _, s := range ip4Pool.Subnets {
		subnets = append(subnets, Ip4Subnet{
			Status:         types.String{Value: s.Status},
			Network:        types.String{Value: s.Network.String()},
			Total:          types.Int64{Value: s.Total},
			Used:           types.Int64{Value: s.Used},
			UsedPercentage: types.Float64{Value: float64(s.UsedPercentage)},
		})
	}

	// Set state
	diags = resp.State.Set(ctx, &DataIp4Pool{
		Id:             types.String{Value: string(ip4Pool.Id)},
		Name:           types.String{Value: ip4Pool.DisplayName},
		Status:         types.String{Value: ip4Pool.Status},
		Tags:           tags,
		Used:           types.Int64{Value: int64(ip4Pool.Used)},
		UsedPercent:    types.Float64{Value: float64(ip4Pool.UsedPercentage)},
		CreatedAt:      types.String{Value: ip4Pool.CreatedAt.String()},
		LastModifiedAt: types.String{Value: ip4Pool.LastModifiedAt.String()},
		Total:          types.Int64{Value: int64(ip4Pool.Total)},
		Subnets:        subnets,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
