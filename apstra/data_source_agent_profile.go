package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAgentProfile{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceAgentProfile{}

type dataSourceAgentProfile struct {
	client *goapstra.Client
}

func (o *dataSourceAgentProfile) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_profile"
}

func (o *dataSourceAgentProfile) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceAgentProfile) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource looks up details of an Agent Profile using either its name (Apstra ensures these are unique), or its ID (but not both).",
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
	}, nil
}

func (o *dataSourceAgentProfile) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dAgentProfile
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
}

func (o *dataSourceAgentProfile) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dAgentProfile
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var agentProfile *goapstra.AgentProfile
	switch {
	case !config.Name.Null:
		agentProfile, err = o.client.GetAgentProfileByLabel(ctx, config.Name.Value)
	case !config.Id.Null:
		agentProfile, err = o.client.GetAgentProfile(ctx, goapstra.ObjectId(config.Id.Value))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)

	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Agent Profile",
			fmt.Sprintf("error retrieving agent profile - %s", err),
		)
		return
	}

	// reset every element in config
	config.Id = types.String{Value: string(agentProfile.Id)}
	config.Name = types.String{Value: agentProfile.Label}
	config.Platform = types.String{Value: agentProfile.Platform}
	config.HasUsername = types.Bool{Value: agentProfile.HasUsername}
	config.HasPassword = types.Bool{Value: agentProfile.HasPassword}
	config.OpenOptions = types.Map{
		Null:     len(agentProfile.OpenOptions) == 0,
		Elems:    make(map[string]attr.Value, len(agentProfile.OpenOptions)),
		ElemType: types.StringType,
	}
	config.Packages = types.Map{
		Null:     len(agentProfile.Packages) == 0,
		Elems:    make(map[string]attr.Value, len(agentProfile.Packages)),
		ElemType: types.StringType,
	}

	// populate OpenOptions map
	for k, v := range agentProfile.OpenOptions {
		config.OpenOptions.Elems[k] = types.String{Value: v}
	}

	// populate Packages map
	for k, v := range agentProfile.Packages {
		config.Packages.Elems[k] = types.String{Value: v}
	}

	// Set state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

type dAgentProfile struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Platform    types.String `tfsdk:"platform"`
	HasUsername types.Bool   `tfsdk:"has_username"`
	HasPassword types.Bool   `tfsdk:"has_password"`
	Packages    types.Map    `tfsdk:"packages"`
	OpenOptions types.Map    `tfsdk:"open_options"`
}
