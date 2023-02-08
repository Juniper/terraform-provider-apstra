package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Agent Profile. Required when `id` is omitted.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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
	var config agentProfile
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
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

	var config agentProfile
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var ap *goapstra.AgentProfile
	switch {
	case !config.Name.IsNull():
		ap, err = o.client.GetAgentProfileByLabel(ctx, config.Name.ValueString())
	case !config.Id.IsNull():
		ap, err = o.client.GetAgentProfile(ctx, goapstra.ObjectId(config.Id.ValueString()))
	default:
		resp.Diagnostics.AddError(errDataSourceReadFail, errInsufficientConfigElements)

	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving Agent Profile",
			fmt.Sprintf("error retrieving agent profile - %s", err),
		)
		return
	}

	var state agentProfile
	state.loadApiResponse(ctx, ap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
