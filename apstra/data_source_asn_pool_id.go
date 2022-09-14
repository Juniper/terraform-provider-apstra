package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceAsnPoolIdType struct{}

func (r dataSourceAsnPoolIdType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source returns the pool ID of the ASN resource pool matching the supplied criteria. " +
			"At least one optional attribute is required.  " +
			"It is incumbent on the user to ensure the criteria matches exactly one ASN pool. " +
			"Matching zero pools or more than one pool will produce an error.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "The pool ID of the single matching ASN resource pool.",
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Used to filter when searching for a single match.  The name of the single matching ASN resource pool.",
				Optional:            true,
				Type:                types.StringType,
			},
			"tags": {
				MarkdownDescription: "Used to filter when searching for a single match.  Required tags of the single matching ASN resource pool.  The pool may have other tags which do not appear on this list",
				Optional:            true,
				Type:                types.ListType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (r dataSourceAsnPoolIdType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceAsnPoolId{
		p: *(p.(*apstraProvider)),
	}, nil
}

type dataSourceAsnPoolId struct {
	p apstraProvider
}

func (r dataSourceAsnPoolId) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// get all ASN Pools from Apstra
	pools, err := r.p.client.GetAsnPools(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving ASN pools",
			fmt.Sprintf("error retrieving ASN pools - %s", err),
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
		if asnPoolFilterMatch(&p, &config) {
			if !found {
				found = true
				asnPoolId = string(p.Id)
			} else {
				resp.Diagnostics.AddError(
					"multiple matches found",
					"consider updating 'name' or 'tags' to ensure exactly one ASN Pool is matched",
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

func asnPoolFilterMatch(p *goapstra.AsnPool, cfg *DataAsnPoolId) bool {
	if !cfg.Name.IsNull() && !asnPoolMatchName(p, cfg) {
		return false
	}

	if len(cfg.Tags) > 0 && !asnPoolMatchTags(p, cfg) {
		return false
	}

	return true
}

func asnPoolMatchName(p *goapstra.AsnPool, cfg *DataAsnPoolId) bool {
	if p.DisplayName == cfg.Name.Value {
		return true
	}
	return false
}

func asnPoolMatchTags(p *goapstra.AsnPool, cfg *DataAsnPoolId) bool {
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
