package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAgentProfiles{}

type dataSourceAgentProfiles struct {
	client *apstra.Client
}

func (o *dataSourceAgentProfiles) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_profiles"
}

func (o *dataSourceAgentProfiles) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceAgentProfiles) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the ID numbers of all Agent Profiles.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "Filter to select only Agent Profiles for the specified platform.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf(utils.AgentProfilePlatforms()...)},
			},
			"has_credentials": schema.BoolAttribute{
				MarkdownDescription: "Filter to select only Agent Profiles configured with (or without) Username and Password.",
				Optional:            true,
			},
		},
	}
}

func (o *dataSourceAgentProfiles) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := struct {
		Ids            types.Set    `tfsdk:"ids"`
		Platform       types.String `tfsdk:"platform"`
		HasCredentials types.Bool   `tfsdk:"has_credentials"`
	}{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error

	// both of these slices start as nil...
	var agentProfileIds []apstra.ObjectId
	var agentProfiles []apstra.AgentProfile

	// one of the slices from above will get populated, depending on whether the user included filter attributes...
	if config.Ids.IsNull() && config.Platform.IsNull() && config.HasCredentials.IsNull() {
		agentProfileIds, err = o.client.ListAgentProfileIds(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving Agent Profile IDs", err.Error())
			return
		}
	} else {
		agentProfiles, err = o.client.GetAllAgentProfiles(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving Agent Profile IDs", err.Error())
			return
		}
	}

	// loop over agentProfiles unconditionally. If anything's in here, it's interesting.
	for _, profile := range agentProfiles {
		if !config.HasCredentials.IsNull() && config.HasCredentials.ValueBool() != profile.HasCredentials() {
			continue
		}

		if !config.Platform.IsNull() && config.Platform.ValueString() != profile.Platform {
			continue
		}

		agentProfileIds = append(agentProfileIds, profile.Id)
	}

	// convert agentProfileIds ([]apstra.ObjectId) to a types.Set of types.String
	idSet, diags := types.SetValueFrom(ctx, types.StringType, agentProfileIds)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Ids = idSet

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
