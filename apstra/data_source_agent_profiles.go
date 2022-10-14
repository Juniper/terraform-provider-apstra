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

var _ datasource.DataSourceWithConfigure = &dataSourceAgentProfiles{}

type dataSourceAgentProfiles struct {
	client *goapstra.Client
}

func (o *dataSourceAgentProfiles) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_profiles"
}

func (o *dataSourceAgentProfiles) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceAgentProfiles) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource returns the ID numbers of each Agent Profile.",
		Attributes: map[string]tfsdk.Attribute{
			"ids": {
				Computed:            true,
				Type:                types.SetType{ElemType: types.StringType},
				MarkdownDescription: "A set of Apstra ID numbers of each Agent Profile.",
			},
		},
	}, nil
}

func (o *dataSourceAgentProfiles) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	ids, err := o.client.ListAgentProfileIds(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving Agent Profile IDs",
			fmt.Sprintf("error retrieving Agent Profile IDs - %s", err),
		)
		return
	}

	// map response body to resource schema
	config.Ids = make([]types.String, len(ids))
	for i, Id := range ids {
		config.Ids[i] = types.String{Value: string(Id)}
	}

	// Set state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
