package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceTagType struct{}

func (r dataSourceTagType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source provides details of a specific tag.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one tag. " +
			"Matching zero tags or more than one tag will produce an error.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Tag id.  Required when the tag name is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Tag name.  Required when tag id is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"description": {
				MarkdownDescription: "The description of the returned tag.",
				Computed:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func (r dataSourceTagType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceTag{
		p: *(p.(*apstraProvider)),
	}, nil
}

type dataSourceTag struct {
	p apstraProvider
}

func (r dataSourceTag) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config DataTag
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.Null && config.Id.Null) || (!config.Name.Null && !config.Id.Null) { // XOR
		resp.Diagnostics.AddError("configuration error", "exactly one of 'id' and 'key' must be specified")
		return
	}
}

func (r dataSourceTag) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataTag
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var tag *goapstra.DesignTag
	if config.Name.Null == false {
		tag, err = r.p.client.GetTagByLabel(ctx, goapstra.TagLabel(config.Name.Value))
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
		Id:          types.String{Value: string(tag.Id)},
		Name:        types.String{Value: string(tag.Label)},
		Description: types.String{Value: tag.Description},
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
