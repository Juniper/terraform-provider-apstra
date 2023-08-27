package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	systemAgents "github.com/Juniper/terraform-provider-apstra/apstra/system_agents"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceManagedDevice{}

type resourceManagedDevice struct {
	client *apstra.Client
}

func (o *resourceManagedDevice) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_device"
}

func (o *resourceManagedDevice) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errResourceConfigureProviderDataDetail,
			fmt.Sprintf(errResourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *resourceManagedDevice) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates/installs an Agent for an Apstra Managed Device." +
			"Optionally, it will 'Acknowledge' the discovered system if the `device key` (serial number)" +
			"reported by the agent matches the optional `device_key` field.",
		Attributes: systemAgents.ManagedDevice{}.ResourceAttributes(),
	}
}

func (o *resourceManagedDevice) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan systemAgents.ManagedDevice
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ValidateAgentProfile(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Agent for this Managed Device
	agentId, err := o.client.CreateSystemAgent(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent",
			err.Error())
		return
	}
	plan.AgentId = types.StringValue(string(agentId))

	// Install the new agent
	_, err = o.client.SystemAgentRunJob(ctx, agentId, apstra.AgentJobTypeInstall)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Could not run %q job on new agent %q", apstra.AgentJobTypeInstall.String(), agentId),
			err.Error())
		return
	}

	// figure out the new switch system ID
	agentInfo, err := o.client.GetSystemAgent(ctx, agentId)
	if err != nil {
		resp.Diagnostics.AddError(
			"error fetching Agent info",
			err.Error())
		return
	}
	plan.SystemId = types.StringValue(string(agentInfo.Status.SystemId))

	if !plan.DeviceKey.IsNull() {
		// figure out the actual serial number (device_key)
		systemInfo, err := o.client.GetSystemInfo(ctx, agentInfo.Status.SystemId)
		if err != nil {
			resp.Diagnostics.AddError(
				"error fetching system info",
				err.Error())
		}

		// validate discovered device_key (serial number)
		if plan.DeviceKey.ValueString() == systemInfo.DeviceKey {
			// "acknowledge" the managed device:q
			plan.Acknowledge(ctx, systemInfo, o.client, &resp.Diagnostics)
		} else {
			// device_key supplied by config does not match discovered asset
			resp.Diagnostics.AddAttributeError(
				path.Root("device_key"),
				"error system device_key mismatch",
				fmt.Sprintf("config expects switch device_key %q, device reports %q",
					plan.DeviceKey.ValueString(), systemInfo.DeviceKey),
			)
		}
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceManagedDevice) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state systemAgents.ManagedDevice
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AgentInfo from API
	agentInfo, err := o.client.GetSystemAgent(ctx, apstra.ObjectId(state.AgentId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"error reading managed device agent info",
			fmt.Sprintf("Could not Read %q (%s)- %s", state.AgentId.ValueString(), state.ManagementIp.ValueString(), err),
		)
		return
	}

	var newState systemAgents.ManagedDevice
	newState.LoadApiData(ctx, agentInfo, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Device_key has 'requiresReplace()', so if it's not set in the state,
	// then it's also not set in the config. Only fetch the serial number if
	// the config is expecting a serial number.
	if !state.DeviceKey.IsNull() {
		newState.GetDeviceKey(ctx, o.client, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceManagedDevice) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan systemAgents.ManagedDevice
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check Agent Profile for credentials, etc...
	plan.ValidateAgentProfile(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// agent profile ID is the only value permitted to change (others trigger replacement)
	err := o.client.AssignAgentProfile(ctx, &apstra.AssignAgentProfileRequest{
		SystemAgents: []apstra.ObjectId{apstra.ObjectId(plan.AgentId.ValueString())},
		ProfileId:    apstra.ObjectId(plan.AgentProfileId.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating managed device agent",
			fmt.Sprintf("error while updating managed device agent %q (%s) - %s",
				plan.AgentId.ValueString(), plan.ManagementIp.ValueString(), err.Error()),
		)
		return
	}

	// set state to match plan
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (o *resourceManagedDevice) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state systemAgents.ManagedDevice
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var agentDoesNotExist, systemDoesNotExist bool

	_, err := o.client.GetSystemAgent(ctx, apstra.ObjectId(state.AgentId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			agentDoesNotExist = true
		} else {
			resp.Diagnostics.AddError(
				"error pulling agent info",
				fmt.Sprintf("could not get info about agent %q - %s", state.AgentId.ValueString(), err.Error()),
			)
		}
	}

	_, err = o.client.GetSystemInfo(ctx, apstra.SystemId(state.SystemId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			systemDoesNotExist = true
		} else {
			resp.Diagnostics.AddError(
				"error pulling system info",
				fmt.Sprintf("could not get info about system %q - %s", state.SystemId.ValueString(), err.Error()),
			)
		}
	}

	if !agentDoesNotExist {
		err = o.client.DeleteSystemAgent(ctx, apstra.ObjectId(state.AgentId.ValueString()))
		if err != nil {
			if !utils.IsApstra404(err) {
				resp.Diagnostics.AddError(
					"error deleting device agent",
					fmt.Sprintf("device agent %q (%q %q) delete error - %s",
						state.AgentId.ValueString(), state.SystemId.ValueString(), state.ManagementIp.ValueString(), err.Error()))
			}
			return
		}
	}

	if !systemDoesNotExist {
		err = o.client.DeleteSystem(ctx, apstra.SystemId(state.SystemId.ValueString()))
		if err != nil {
			if !utils.IsApstra404(err) {
				resp.Diagnostics.AddError(
					"error deleting managed device",
					fmt.Sprintf("managed device %q (%s) delete error - %s",
						state.SystemId.ValueString(), state.ManagementIp.ValueString(), err.Error()))
			}
			return
		}
	}
}
