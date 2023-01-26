package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func (o *dataSourceAgentProfile) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source looks up details of an Agent Profile using either its name (Apstra ensures these are unique), or its ID (but not both).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Agent Profile. Required when `name` is omitted.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Agent Profile. Required when `id` is omitted.",
				Optional:            true,
				Computed:            true,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "Indicates the platform supported by the Agent Profile.",
				Computed:            true,
			},
			"has_username": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether a username has been configured.",
				Computed:            true,
			},
			"has_password": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether a password has been configured.",
				Computed:            true,
			},
			"packages": schema.MapAttribute{
				MarkdownDescription: "Admin-provided software packages stored on the Apstra server applied to devices using the profile.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"open_options": schema.MapAttribute{
				MarkdownDescription: "Configured parameters for offbox agents",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceAgentProfile) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dAgentProfile
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
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
	case !config.Name.IsNull():
		agentProfile, err = o.client.GetAgentProfileByLabel(ctx, config.Name.ValueString())
	case !config.Id.IsNull():
		agentProfile, err = o.client.GetAgentProfile(ctx, goapstra.ObjectId(config.Id.ValueString()))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)

	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Agent Profile",
			fmt.Sprintf("error retrieving agent profile - %s", err),
		)
		return
	}

	state := parseAgentProfile(ctx, agentProfile, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, state)
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

func (o *dAgentProfile) AgentProfileConfig(ctx context.Context, diags *diag.Diagnostics) *goapstra.AgentProfileConfig {
	var platform string
	if o.Platform.IsNull() || o.Platform.IsUnknown() {
		platform = ""
	} else {
		platform = o.Platform.ValueString()
	}

	packages := make(goapstra.AgentPackages)
	diags.Append(o.Packages.ElementsAs(ctx, &packages, false)...)
	if diags.HasError() {
		return nil
	}

	options := make(map[string]string)
	diags.Append(o.Packages.ElementsAs(ctx, &options, false)...)
	if diags.HasError() {
		return nil
	}

	return &goapstra.AgentProfileConfig{
		Label:       o.Name.ValueString(),
		Platform:    platform,
		Packages:    packages,
		OpenOptions: options,
	}
}

func platformToTFString(platform string) types.String {
	var result types.String
	if platform == "" {
		result = types.StringNull()
	} else {
		result = types.StringValue(platform)
	}
	return result
}

func parseAgentProfile(ctx context.Context, in *goapstra.AgentProfile, diags *diag.Diagnostics) *dAgentProfile {
	var d diag.Diagnostics
	var openOptions, packages types.Map

	if len(in.OpenOptions) == 0 {
		openOptions = types.MapNull(types.StringType)
	} else {
		openOptions, d = types.MapValueFrom(ctx, types.StringType, in.OpenOptions)
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}
	}

	if len(in.Packages) == 0 {
		packages = types.MapNull(types.StringType)
	} else {
		packages, d = types.MapValueFrom(ctx, types.StringType, in.Packages)
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}
	}

	return &dAgentProfile{
		Id:          types.StringValue(string(in.Id)),
		Name:        types.StringValue(in.Label),
		Platform:    platformToTFString(in.Platform),
		HasUsername: types.BoolValue(in.HasUsername),
		HasPassword: types.BoolValue(in.HasPassword),
		Packages:    packages,
		OpenOptions: openOptions,
	}
}
