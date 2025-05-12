package tfapstra

import (
	"context"
	"fmt"
	"net"

	"github.com/Juniper/apstra-go-sdk/apstra"
	systemAgents "github.com/Juniper/terraform-provider-apstra/apstra/system_agents"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceAgents{}
	_ datasourceWithSetClient            = &dataSourceAgents{}
)

type dataSourceAgents struct {
	client *apstra.Client
}

func (o *dataSourceAgents) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agents"
}

func (o *dataSourceAgents) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceAgents) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDevices + "This data source returns the IDs of Managed Device Agents. " +
			"All of the `filter` attributes are optional.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of Agent IDs",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Agent attributes used as a filter",
				DeprecationMessage: "The `filter` attribute is deprecated and will be removed in a future " +
					"release. Please migrate your configuration to use `filters` instead.",
				Optional:   true,
				Attributes: systemAgents.ManagedDevice{}.DataSourceFilterAttributes(),
				Validators: []validator.Object{objectvalidator.ConflictsWith(path.MatchRoot("filters"))},
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "Agent attributes used as a filters",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: systemAgents.ManagedDevice{}.DataSourceFilterAttributes(),
				},
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
			},
		},
	}
}

func (o *dataSourceAgents) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		IDs     types.Set    `tfsdk:"ids"`
		Filter  types.Object `tfsdk:"filter"`
		Filters types.List   `tfsdk:"filters"`
	}

	// Retrieve values from config.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Filter.IsNull() && config.Filters.IsNull() {
		// no filter specified, use the HTTP OPTIONS method as a shortcut
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

	var filters []systemAgents.ManagedDevice
	resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(filters) == 0 {
		filter := systemAgents.ManagedDevice{}
		resp.Diagnostics.Append(config.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		filters = append(filters, filter)
	}

	agents, err := o.client.GetAllSystemAgents(ctx)
	if err != nil {
		resp.Diagnostics.AddError("while getting all system agents", err.Error())
		return
	}

	agentIds := make(map[string]struct{})
	for _, filter := range filters {
		for _, agent := range agents {
			if !filter.AgentId.IsNull() && filter.AgentId.ValueString() != agent.Id.String() {
				continue
			}

			if !filter.SystemId.IsNull() && filter.SystemId.ValueString() != string(agent.Status.SystemId) {
				continue
			}

			if !filter.ManagementIp.IsNull() {
				agentIp := net.ParseIP(agent.Config.ManagementIp)
				if agentIp == nil {
					continue // no address or bogus address == no match
				}

				mgmtNet := filter.IpNetFromManagementIp(ctx, &resp.Diagnostics)
				if resp.Diagnostics.HasError() {
					return
				}

				if !mgmtNet.Contains(agentIp) {
					continue
				}
			}

			if !filter.DeviceKey.IsNull() && filter.DeviceKey.ValueString() != string(agent.Status.SystemId) {
				continue
			}

			if !filter.AgentProfileId.IsNull() && filter.AgentProfileId.ValueString() != agent.Config.Profile.String() {
				continue
			}

			if !filter.OffBox.IsNull() && filter.OffBox.ValueBool() != bool(agent.Config.AgentTypeOffBox) {
				continue
			}

			if !filter.Location.IsNull() {
				systemInfo, err := o.client.GetSystemInfo(ctx, agent.Status.SystemId)
				if err != nil {
					resp.Diagnostics.AddError(fmt.Sprintf("While getting info for system %q", agent.Status.SystemId), err.Error())
					return
				}

				if filter.Location.ValueString() != systemInfo.UserConfig.Location {
					continue
				}
			}

			agentIds[agent.Id.String()] = struct{}{}
		}
	}

	agentIdVals := make([]attr.Value, len(agentIds))
	var i int
	for agentId := range agentIds {
		agentIdVals[i] = types.StringValue(agentId)
		i++
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

func (o *dataSourceAgents) setClient(client *apstra.Client) {
	o.client = client
}
