package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAsnPool{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceAsnPool{}

type dataSourceAsnPool struct {
	client *goapstra.Client
}

func (o *dataSourceAsnPool) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_pool"
}

func (o *dataSourceAsnPool) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errDataSourceConfigureProviderDataDetail,
			fmt.Sprintf(errDataSourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *dataSourceAsnPool) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides details of a single ASN Resource Pool. It is incumbent upon " +
			"the user to set enough optional criteria to match exactly one ASN Resource Pool. Matching zero or more " +
			"pools will produce an error.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of the desired ASN Resource Pool.",
				Computed:            true,
				Optional:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Display name of the ASN Resource Pool.",
				Computed:            true,
				Optional:            true,
				Type:                types.StringType,
			},
			"status": {
				MarkdownDescription: "Status of the ASN Resource Pool.",
				Computed:            true,
				Type:                types.StringType,
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
						Computed:            true,
						Type:                types.StringType,
					},
					"first": {
						MarkdownDescription: "Lowest numbered AS in this ASN Pool Range.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"last": {
						MarkdownDescription: "Highest numbered AS in this ASN Pool Range.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"total": {
						MarkdownDescription: "Total number of ASNs in the ASN Pool Range.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"used": {
						MarkdownDescription: "Count of used ASNs in the ASN Pool Range.",
						Computed:            true,
						Type:                types.Int64Type,
					},
					"used_percentage": {
						MarkdownDescription: "Percent of used ASNs in the ASN Pool Range.",
						Computed:            true,
						Type:                types.Float64Type,
					},
				}),
			},
		},
	}, nil
}

func (o *dataSourceAsnPool) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dAsnPool
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.Null && config.Id.Null) || (!config.Name.Null && !config.Id.Null) { // XOR
		resp.Diagnostics.AddError(
			"cannot search for ASN Pool",
			"exactly one of 'name' or 'id' must be specified",
		)
	}
}

func (o *dataSourceAsnPool) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dAsnPool
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var asnPool *goapstra.AsnPool
	switch {
	case !config.Name.Null:
		asnPool, err = o.client.GetAsnPoolByName(ctx, config.Name.Value)
	case !config.Id.Null:
		asnPool, err = o.client.GetAsnPool(ctx, goapstra.ObjectId(config.Id.Value))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving ASN pool",
			fmt.Sprintf("cannot retrieve ASN pool - %s", err),
		)
		return
	}

	config.Id = types.String{Value: string(asnPool.Id)}
	config.Name = types.String{Value: asnPool.DisplayName}
	config.Status = types.String{Value: asnPool.Status}
	config.Used = types.Int64{Value: int64(asnPool.Used)}
	config.UsedPercent = types.Float64{Value: float64(asnPool.UsedPercentage)}
	config.CreatedAt = types.String{Value: asnPool.CreatedAt.String()}
	config.LastModifiedAt = types.String{Value: asnPool.LastModifiedAt.String()}
	config.Total = types.Int64{Value: int64(asnPool.Total)}
	config.Ranges = make([]dAsnRange, len(asnPool.Ranges))

	for i, r := range asnPool.Ranges {
		config.Ranges[i] = dAsnRange{
			Status:         types.String{Value: r.Status},
			First:          types.Int64{Value: int64(r.First)},
			Last:           types.Int64{Value: int64(r.Last)},
			Total:          types.Int64{Value: int64(r.Total)},
			Used:           types.Int64{Value: int64(r.Used)},
			UsedPercentage: types.Float64{Value: float64(r.UsedPercentage)},
		}
	}

	// Set state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

type dAsnPool struct {
	Id             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	Status         types.String  `tfsdk:"status"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercent    types.Float64 `tfsdk:"used_percentage"`
	CreatedAt      types.String  `tfsdk:"created_at"`
	LastModifiedAt types.String  `tfsdk:"last_modified_at"`
	Total          types.Int64   `tfsdk:"total"`
	Ranges         []dAsnRange   `tfsdk:"ranges"`
}

type dAsnRange struct {
	Status         types.String  `tfsdk:"status"`
	First          types.Int64   `tfsdk:"first"`
	Last           types.Int64   `tfsdk:"last"`
	Total          types.Int64   `tfsdk:"total"`
	Used           types.Int64   `tfsdk:"used"`
	UsedPercentage types.Float64 `tfsdk:"used_percentage"`
}
