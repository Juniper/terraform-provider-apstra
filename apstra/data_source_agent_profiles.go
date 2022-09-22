package apstra

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceAgentProfilesType struct{}

func (r dataSourceAgentProfiles) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "apstra_agent_profiles"
}

func (r dataSourceAgentProfiles) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"ids": {
				Computed:            true,
				Type:                types.SetType{ElemType: types.StringType},
				MarkdownDescription: "A set of Apstra ID numbers of each Agent Profile.",
			},
		},
		MarkdownDescription: "This resource returns the ID numbers of each Agent Profile.",
	}, nil
}

func (r dataSourceAgentProfilesType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return dataSourceAgentProfiles{
		p: *(p.(*Provider)),
	}, nil
}

type dataSourceAgentProfiles struct {
	p Provider
}

func (r dataSourceAgentProfiles) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ids, err := r.p.client.ListAgentProfileIds(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving ASN pool IDs",
			fmt.Sprintf("error retrieving ASN pool IDs - %s", err),
		)
		return
	}

	// map response body to resource schema
	var state DataAgentProfileIds
	for _, Id := range ids {
		state.Ids = append(state.Ids, types.String{Value: string(Id)})
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
