package apstra

import (
	"context"
	"fmt"
	"github.com/chrismarget-j/goapstra"
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
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"name": {
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"platform": {
				Computed: true,
				Type:     types.StringType,
			},
			"has_username": {
				Computed: true,
				Type:     types.BoolType,
			},
			"has_password": {
				Computed: true,
				Type:     types.BoolType,
			},
			"packages": {
				Computed: true,
				Type:     types.MapType{ElemType: types.StringType},
			},
			"open_options": {
				Computed: true,
				Type:     types.MapType{ElemType: types.StringType},
			},
		},
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
