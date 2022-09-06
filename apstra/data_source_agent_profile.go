package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceAgentProfileType struct{}

func (r dataSourceAgentProfileType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
				MarkdownDescription: "ID of the agent profile. Required when name is omitted.",
			},
			"name": {
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
				MarkdownDescription: "Name of the agent profile. Required when id is omitted.",
			},
			// todo: make platform a search criteria
			"platform": {
				Computed:            true,
				Type:                types.StringType,
				MarkdownDescription: "Indicates the platform supported by the agent profile.",
			},
			"has_username": {
				Computed:            true,
				Type:                types.BoolType,
				MarkdownDescription: "Indicates whether a username has been configured.",
			},
			"has_password": {
				Computed:            true,
				Type:                types.BoolType,
				MarkdownDescription: "Indicates whether a password has been configured.",
			},
			"packages": {
				Computed:            true,
				Type:                types.MapType{ElemType: types.StringType},
				MarkdownDescription: "Admin-provided software packages stored on the Apstra server applied to devices using the profile.",
			},
			"open_options": {
				Computed:            true,
				Type:                types.MapType{ElemType: types.StringType},
				MarkdownDescription: "Configured parameters for offbox agents",
			},
		},
		MarkdownDescription: "This resource looks up details of an Agent Profile using either its name (Apstra ensures these are unique), or its ID (but not both).",
	}, nil
}

func (r dataSourceAgentProfileType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceAgentProfile{
		p: *(p.(*provider)),
	}, nil
}

type dataSourceAgentProfile struct {
	p provider
}

func (r dataSourceAgentProfile) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var config DataAgentProfile
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.Null && config.Id.Null) || (!config.Name.Null && !config.Id.Null) { // XOR
		resp.Diagnostics.AddError(
			"cannot search for Agent Profile",
			"exactly one of 'name' or 'id' must be specified",
		)
	}

	var err error
	var agentProfile *goapstra.AgentProfile
	if config.Name.Null == false {
		agentProfile, err = r.p.client.GetAgentProfileByLabel(ctx, config.Name.Value)
	}
	if config.Id.Null == false {
		agentProfile, err = r.p.client.GetAgentProfile(ctx, goapstra.ObjectId(config.Id.Value))
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving Agent Profile",
			fmt.Sprintf("error retrieving agent profile '%s%s' - %s", config.Id.Value, config.Name.Value, err),
		)
		return
	}

	// convert open options from map[string]string to []types.Map
	openOptions := types.Map{
		Null:     len(agentProfile.OpenOptions) == 0,
		Elems:    make(map[string]attr.Value),
		ElemType: types.StringType,
	}
	for k, v := range agentProfile.OpenOptions {
		openOptions.Elems[k] = types.String{Value: v}
	}

	// convert packages from map[string]string to []types.Map
	packages := types.Map{
		Null:     len(agentProfile.Packages) == 0,
		Elems:    make(map[string]attr.Value),
		ElemType: types.StringType,
	}
	for k, v := range agentProfile.Packages {
		packages.Elems[k] = types.String{Value: v}
	}

	// Set state
	diags = resp.State.Set(ctx, &DataAgentProfile{
		Id:          types.String{Value: string(agentProfile.Id)},
		Name:        types.String{Value: agentProfile.Label},
		Platform:    types.String{Value: agentProfile.Platform},
		HasUsername: types.Bool{Value: agentProfile.HasUsername},
		HasPassword: types.Bool{Value: agentProfile.HasPassword},
		Packages:    packages,
		OpenOptions: openOptions,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
