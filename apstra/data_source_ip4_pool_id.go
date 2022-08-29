package apstra

import (
	"context"
	"fmt"
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceIp4PoolIdType struct{}

func (r dataSourceIp4PoolIdType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"name": {
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

func (r dataSourceIp4PoolIdType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceIp4PoolId{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceIp4PoolId struct {
	p provider
}

func (r dataSourceIp4PoolId) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// get all ASN Pools from Apstra
	pools, err := r.p.client.GetIp4Pools(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving IPv4 pools",
			fmt.Sprintf("error retrieving IPv4 pools - %s", err),
		)
		return
	}

	// read the incoming config (filters are here)
	var config DataIp4PoolId
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagstr := ":"
	for _, t := range config.Tags {
		tagstr = tagstr + t.Value + ":"
	}

	// loop through IPv4 Pools, find a match
	var ip4PoolId string // for search results
	var found bool       // for search results
	for _, p := range pools {
		if ip4PoolFilterMatch(&p, &config) {
			if !found {
				found = true
				ip4PoolId = string(p.Id)
			} else {
				resp.Diagnostics.AddError(
					"multiple matches found",
					"consider updating 'name' or 'tags' to ensure exactly one IPv4 Pool is matched",
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
	diags = resp.State.Set(ctx, &DataIp4PoolId{Id: types.String{Value: ip4PoolId}})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func ip4PoolFilterMatch(p *goapstra.Ip4Pool, cfg *DataIp4PoolId) bool {
	if !cfg.Name.IsNull() && !ip4PoolMatchName(p, cfg) {
		return false
	}

	if len(cfg.Tags) > 0 && !ip4PoolMatchTags(p, cfg) {
		return false
	}

	return true
}

func ip4PoolMatchName(p *goapstra.Ip4Pool, cfg *DataIp4PoolId) bool {
	if p.DisplayName == cfg.Name.Value {
		return true
	}
	return false
}

func ip4PoolMatchTags(p *goapstra.Ip4Pool, cfg *DataIp4PoolId) bool {
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
