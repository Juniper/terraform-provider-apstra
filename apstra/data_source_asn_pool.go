package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceAsnPoolType struct{}

func (r dataSourceAsnPoolType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides details of a single ASN Resource Pool by its ID.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of the desired ASN Resource Pool.",
				Required:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Display name of the ASN Resource Pool.",
				Computed:            true,
				Type:                types.StringType,
			},
			"status": {
				MarkdownDescription: "Status of the ASN Resource Pool, as reported by Apstra.",
				Computed:            true,
				Type:                types.StringType,
			},
			"tags": {
				MarkdownDescription: "Tags applied to the ASN Resource Pool.",
				Computed:            true,
				Type:                types.ListType{ElemType: types.StringType},
			},
			"total": {
				MarkdownDescription: "Total number of ASNs in the ASN Resource Pool.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"used": {
				MarkdownDescription: "Count of used ASNs in the ASN Resource Pool.",
				Computed:            true,
				Type:                types.Int64Type,
			},
			"used_percentage": {
				MarkdownDescription: "Percent of used ASNs in the ASN Resource Pool.",
				Computed:            true,
				Type:                types.Float64Type,
			},
			"created_at": {
				MarkdownDescription: "Creation time of the ASN Resource Pool.",
				Computed:            true,
				Type:                types.StringType,
			},
			"last_modified_at": {
				MarkdownDescription: "Modification time of the ASN Resource Pool.",
				Computed:            true,
				Type:                types.StringType,
			},
			"ranges": {
				MarkdownDescription: "Detailed info about individual ASN Pool Ranges within the ASN Resource Pool.",
				Computed:            true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"status": {
						MarkdownDescription: "Status of the ASN Pool Range, as reported by Apstra.",
						Type:                types.StringType,
						Computed:            true,
					},
					"first": {
						MarkdownDescription: "Lowest numbered AS in this ASN Pool Range.",
						Type:                types.Int64Type,
						Computed:            true,
					},
					"last": {
						MarkdownDescription: "Highest numbered AS in this ASN Pool Range.",
						Type:                types.Int64Type,
						Computed:            true,
					},
					"total": {
						MarkdownDescription: "Total number of ASNs in the ASN Pool Range.",
						Type:                types.Int64Type,
						Computed:            true,
					},
					"used": {
						MarkdownDescription: "Count of used ASNs in the ASN Pool Range.",
						Type:                types.Int64Type,
						Computed:            true,
					},
					"used_percentage": {
						MarkdownDescription: "Percent of used ASNs in the ASN Pool Range.",
						Type:                types.Float64Type,
						Computed:            true,
					},
				}),
			},
		},
	}, nil
}

func (r dataSourceAsnPoolType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceAsnPool{
		p: *(p.(*apstraProvider)),
	}, nil
}

type dataSourceAsnPool struct {
	p apstraProvider
}

func (r dataSourceAsnPool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
