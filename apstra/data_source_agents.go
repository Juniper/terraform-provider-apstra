package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"net"
	systemAgents "terraform-provider-apstra/apstra/system_agents"
	"terraform-provider-apstra/apstra/utils"
)

var _ datasource.DataSourceWithConfigure = &dataSourceAgents{}

type dataSourceAgents struct {
	client *apstra.Client
}

func (o *dataSourceAgents) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agents"
}

func (o *dataSourceAgents) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	o.client = DataSourceGetClient(ctx, req, resp)
}

func (o *dataSourceAgents) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source returns the IDs of Managed Device Agents. " +
			"All of the `filters` attributes are optional.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of Routing Zone IDs",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"filters": schema.SingleNestedAttribute{
				MarkdownDescription: "Agent attributes used as filters",
				Optional:            true,
				Attributes:          systemAgents.ManagedDevice{}.DataSourceFilterAttributes(),
			},
		},
	}
}

func (o *dataSourceAgents) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	type systems struct {
		IDs     types.Set    `tfsdk:"ids"`
		Filters types.Object `tfsdk:"filters"`
	}

	var config systems
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Filters.IsNull() {
		// no filters specified, use the HTTP OPTIONS method as a shortcut
		ids, err := o.client.ListSystemAgents(ctx)
		if err != nil {
			resp.Diagnostics.AddError("while listing system agents", err.Error())
			return
		}

		// store the result before committing it to the state
		config.IDs = utils.SetValueOrNull(ctx, types.StringType, ids, &resp.Diagnostics)

		// set state
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	filters := &systemAgents.ManagedDevice{}
	d := config.Filters.As(ctx, &filters, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	agents, err := o.client.GetAllSystemAgents(ctx)
	if err != nil {
		resp.Diagnostics.AddError("while getting all system agents", err.Error())
		return
	}

	var agentIdVals []attr.Value
	for _, agent := range agents {
		if !filters.AgentId.IsNull() && filters.AgentId.ValueString() != agent.Id.String() {
			continue
		}

		if !filters.SystemId.IsNull() && filters.SystemId.ValueString() != string(agent.Status.SystemId) {
			continue
		}

		if !filters.ManagementIp.IsNull() {
			agentIp := net.ParseIP(agent.Config.ManagementIp)
			if agentIp == nil {
				continue // no address or bogus address == no match
			}

			mgmtNet := filters.IpNetFromManagementIp(ctx, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			if !mgmtNet.Contains(agentIp) {
				continue
			}
		}

		if !filters.DeviceKey.IsNull() && filters.DeviceKey.ValueString() != string(agent.Status.SystemId) {
			continue
		}

		if !filters.AgentProfileId.IsNull() && filters.AgentProfileId.ValueString() != agent.Config.Profile.String() {
			continue
		}

		if !filters.OffBox.IsNull() && filters.OffBox.ValueBool() != bool(agent.Config.AgentTypeOffBox) {
			continue
		}

		agentIdVals = append(agentIdVals, types.StringValue(agent.Id.String()))
	}

	// store the result before committing it to the state
	config.IDs = types.SetValueMust(types.StringType, agentIdVals)
	if len(config.IDs.Elements()) == 0 {
		// return null rather than empty list
		config.IDs = types.SetNull(types.StringType)
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
