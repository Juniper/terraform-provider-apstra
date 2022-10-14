package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAsnPools{}

type dataSourceAsnPools struct {
	client *goapstra.Client
}

func (o *dataSourceAsnPools) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_pools"
}

func (o *dataSourceAsnPools) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceAsnPools) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source returns the IDs of all ASN resource pools",
		Attributes: map[string]tfsdk.Attribute{
			"ids": {
				MarkdownDescription: "Pool IDs of all ASN resource pools.",
				Computed:            true,
				Type:                types.SetType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (o *dataSourceAsnPools) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config struct {
		Ids []types.String `tfsdk:"ids"`
	}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolIds, err := o.client.ListAsnPoolIds(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving ASN pool IDs",
			fmt.Sprintf("error retrieving ASN pool IDs - %s", err),
		)
		return
	}

	// map response body to resource schema
	config.Ids = make([]types.String, len(poolIds))
	for i, id := range poolIds {
		config.Ids[i] = types.String{Value: string(id)}
	}

	// Set state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
