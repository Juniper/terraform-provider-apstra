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

var _ datasource.DataSourceWithConfigure = &dataSourceAgentProfile{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceAgentProfile{}

type dataSourceAgentProfile struct {
	client *goapstra.Client
}

func (o *dataSourceAgentProfile) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_profile"
}

func (o *dataSourceAgentProfile) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
		MarkdownDescription: "This data source looks up details of an Agent Profile using either its name (Apstra ensures these are unique), or its ID (but not both).",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of the Agent Profile. Required when `name` is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Name of the Agent Profile. Required when `id` is omitted.",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"platform": {
				MarkdownDescription: "Indicates the platform supported by the Agent Profile.",
				Computed:            true,
				Type:                types.StringType,
			},
			"has_username": {
				MarkdownDescription: "Indicates whether a username has been configured.",
				Computed:            true,
				Type:                types.BoolType,
			},
			"has_password": {
				MarkdownDescription: "Indicates whether a password has been configured.",
				Computed:            true,
				Type:                types.BoolType,
			},
			"packages": {
				MarkdownDescription: "Admin-provided software packages stored on the Apstra server applied to devices using the profile.",
				Computed:            true,
				Type:                types.MapType{ElemType: types.StringType},
			},
			"open_options": {
				MarkdownDescription: "Configured parameters for offbox agents",
				Computed:            true,
				Type:                types.MapType{ElemType: types.StringType},
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
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

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

	// Set state
	diags = resp.State.Set(ctx, &dAgentProfile{
		Id:          types.String{Value: string(agentProfile.Id)},
		Name:        types.String{Value: agentProfile.Label},
		Platform:    platformToTFString(agentProfile.Platform),
		HasUsername: types.Bool{Value: agentProfile.HasUsername},
		HasPassword: types.Bool{Value: agentProfile.HasPassword},
		Packages:    mapStringStringToTypesMap(agentProfile.Packages),
		OpenOptions: mapStringStringToTypesMap(agentProfile.OpenOptions),
	})
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

func (o *dAgentProfile) AgentProfileConfig() *goapstra.AgentProfileConfig {
	var platform string
	if o.Platform.IsNull() || o.Platform.IsUnknown() {
		platform = ""
	} else {
		platform = o.Platform.Value
	}
	return &goapstra.AgentProfileConfig{
		Label:       o.Name.Value,
		Platform:    platform,
		Packages:    typesMapToMapStringString(o.Packages),
		OpenOptions: typesMapToMapStringString(o.OpenOptions),
	}
}

func platformToTFString(platform string) types.String {
	var result types.String
	if platform == "" {
		result = types.String{Null: true}
	} else {
		result = types.String{Value: platform}
	}
	return result
}
