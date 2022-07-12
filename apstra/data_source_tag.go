package apstra

import (
	"context"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceTagType struct{}

func (r dataSourceTagType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"key": {
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"value": {
				Computed: true,
				Type:     types.StringType,
			},
		},
	}, nil
}

func (r dataSourceTagType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceTag{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceTag struct {
	p provider
}

func (r dataSourceTag) ValidateConfig(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
	var config DataTag
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Key.Null && config.Id.Null) || (!config.Key.Null && !config.Id.Null) { // XOR
		resp.Diagnostics.AddError("configuration error", "exactly one of 'id' and 'key' must be specified")
		return
	}
}

func (r dataSourceTag) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var config DataTag
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var tag *goapstra.DesignTag
	if config.Key.Null == false {
		tag, err = r.p.client.GetTagByLabel(ctx, config.Key.Value)
	}
	if config.Id.Null == false {
		tag, err = r.p.client.GetTag(ctx, goapstra.ObjectId(config.Id.Value))
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Tag", err.Error())
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &DataTag{
		Id:    types.String{Value: string(tag.Id)},
		Key:   types.String{Value: tag.Label},
		Value: types.String{Value: tag.Description},
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
