package apstra

import (
	"context"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceAsnPoolIdType struct{}

func (r dataSourceAsnPoolIdType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"display_name": {
				Optional: true,
				Type:     types.StringType,
			},
			"tags": {
				Optional: true,
				Type:     types.ListType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (r dataSourceAsnPoolIdType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceAsnPoolId{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceAsnPoolId struct {
	p provider
}

func (r dataSourceAsnPoolId) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// get all ASN Pools from Apstra
	pools, err := r.p.client.GetAsnPools(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving ASN pool IDs",
			fmt.Sprintf("error retrieving ASN pool IDs - %s", err),
		)
		return
	}

	// read the incoming config (filters are here)
	var config DataAsnPoolId
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// loop through ASN Pools, find a match
	var asnPoolId string // for search results
	var found bool       // for search results
	for _, p := range pools {
		if asnPoolFilterMatch(ctx, &p, &config) {
			if !found {
				found = true
				asnPoolId = string(p.Id)
			} else {
				resp.Diagnostics.AddError(
					"multiple matches found",
					"consider updating 'display_name' or 'tags' to ensure exactly one ASN Pool is matched",
				)
				return
			}
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"no matches found",
			"no ASN pools matched filter criteria",
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &DataAsnPoolId{Id: types.String{Value: asnPoolId}})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func asnPoolFilterMatch(ctx context.Context, p *goapstra.AsnPool, cfg *DataAsnPoolId) bool {
	if !cfg.DisplayName.IsNull() && !asnPoolMatchDisplayName(ctx, p, cfg) {
		return false
	}

	if len(cfg.Tags) > 0 && !asnPoolMatchTags(ctx, p, cfg) {
		return false
	}

	return true
}

func asnPoolMatchDisplayName(ctx context.Context, p *goapstra.AsnPool, cfg *DataAsnPoolId) bool {
	if p.DisplayName == cfg.DisplayName.Value {
		return true
	}
	return false
}

func asnPoolMatchTags(ctx context.Context, p *goapstra.AsnPool, cfg *DataAsnPoolId) bool {
	// map-ify pool tags from API
	pTags := make(map[string]struct{})
	for _, t := range p.Tags {
		pTags[t] = struct{}{}
	}

	// every tag specified in the data source config must also be found in pTags
	for _, cfgTag := range cfg.Tags {
		if _, found := pTags[cfgTag.Value]; !found {
			return false
		}
	}
	return true
}
