package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

var _ resource.ResourceWithConfigure = &resourceManagedDevice{}
var _ resource.ResourceWithValidateConfig = &resourceManagedDevice{}

type resourceManagedDevice struct {
	client *goapstra.Client
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

func (o *resourceManagedDevice) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource creates/installs an Agent for an Apstra Managed Device." +
			"Optionally, it will 'Acknolwedge' the discovered system if the `device key` (serial number)" +
			"reported by the agent matches the optional `device_key` field.",
		Attributes: map[string]tfsdk.Attribute{
			"agent_id": {
				MarkdownDescription: "Apstra ID for the Managed Device Agent.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"system_id": {
				MarkdownDescription: "Apstra ID for the System onboarded by the Managed Device Agent.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"management_ip": {
				MarkdownDescription: "Management IP address of the system.",
				Type:                types.StringType,
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"device_key": {
				MarkdownDescription: "Key which uniquely identifies a System asset. Possibly a MAC address or serial number.",
				Type:                types.StringType,
				Optional:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"agent_profile_id": {
				MarkdownDescription: "ID of the Agent Profile used when instantiating the Agent. An Agent Profile is" +
					"required to specify the login credentials and platform type.",
				Type:     types.StringType,
				Required: true,
			},
			"off_box": {
				MarkdownDescription: "Indicates that an 'Offbox' agent should be created (required for Junos devices)",
				Type:                types.BoolType,
				Computed:            true,
				Optional:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace(), resource.UseStateForUnknown()},
			},
		},
	}, nil
}

func (o *resourceManagedDevice) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if o.client == nil { // cannot proceed without a client
		return
	}

	var config rManagedDevice
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipStr := net.ParseIP(config.ManagementIp.Value)
	if ipStr == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("management_ip"),
			"cannot parse management_ip",
			fmt.Sprintf("is '%s' an IP address?", config.ManagementIp.Value))
	}

	config.validateAgentProfile(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceManagedDevice) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rManagedDevice
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.validateAgentProfile(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Agent for this Managed Device
	agentId, err := o.client.CreateAgent(ctx, &goapstra.SystemAgentRequest{
		AgentTypeOffbox: goapstra.AgentTypeOffbox(plan.OffBox.Value),
		ManagementIp:    plan.ManagementIp.Value,
		Profile:         goapstra.ObjectId(plan.AgentProfileId.Value),
		OperationMode:   goapstra.AgentModeFull,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent",
			err.Error())
		return
	}

	// Install the new agent
	_, err = o.client.SystemAgentRunJob(ctx, agentId, goapstra.AgentJobTypeInstall)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Could not run 'install' job on new agent '%s'", agentId),
			err.Error())
		return
	}

	// figure out the new switch system Id
	agentInfo, err := o.client.GetSystemAgent(ctx, agentId)
	if err != nil {
		resp.Diagnostics.AddError(
			"error fetching Agent info",
			err.Error())
		return
	}

	// figure out the new switch serial number (device_key)
	systemInfo, err := o.client.GetSystemInfo(ctx, agentInfo.Status.SystemId)
	if err != nil {
		resp.Diagnostics.AddError(
			"error fetching system info",
			err.Error())
		return
	}

	// submit a SystemUserConfig only if device_key was supplied
	if !plan.DeviceKey.IsNull() && !plan.DeviceKey.IsUnknown() {
		// mismatched device key is fatal
		if plan.DeviceKey.Value != systemInfo.DeviceKey {
			resp.Diagnostics.AddAttributeError(
				path.Root("device_key"),
				"error system device_key mismatch",
				fmt.Sprintf("config expects switch device_key '%s', device reports '%s'",
					plan.DeviceKey.Value, systemInfo.DeviceKey),
			)
			return
		}

		// update with new SystemUserConfig
		err = o.client.UpdateSystem(ctx, agentInfo.Status.SystemId, &goapstra.SystemUserConfig{
			AosHclModel: systemInfo.Facts.AosHclModel,
			AdminState:  goapstra.SystemAdminStateNormal,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"error updating managed device",
				fmt.Sprintf("unexpected error while updating user config: %s", err.Error()),
			)
			return
		}
	}

	plan.AgentId = types.String{Value: string(agentId)}
	plan.SystemId = types.String{Value: string(agentInfo.Status.SystemId)}

	diags = resp.State.Set(ctx, &rManagedDevice{
		AgentId:        types.String{Value: string(agentId)},
		SystemId:       types.String{Value: string(agentInfo.Status.SystemId)},
		ManagementIp:   types.String{Value: agentInfo.Config.ManagementIp},
		DeviceKey:      types.String{Value: systemInfo.DeviceKey},
		AgentProfileId: types.String{Value: string(agentInfo.Config.Profile)},
		OffBox:         types.Bool{Value: bool(agentInfo.Config.AgentTypeOffBox)},
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceManagedDevice) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rManagedDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AgentInfo from API
	agentInfo, err := o.client.GetSystemAgent(ctx, goapstra.ObjectId(state.AgentId.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading managed device agent info",
				fmt.Sprintf("Could not Read '%s' (%s)- %s", state.AgentId.Value, state.ManagementIp.Value, err),
			)
			return
		}
	}

	// Get SystemInfo from API
	systemInfo, err := o.client.GetSystemInfo(ctx, agentInfo.Status.SystemId)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
		} else {
			resp.Diagnostics.AddError(
				"error reading managed device system info",
				fmt.Sprintf("Could not Read '%s' (%s) - %s", state.SystemId.Value, state.ManagementIp.Value, err),
			)
			return
		}
	}

	// 	AgentId        types.String `tfsdk:"agent_id"`
	//	DeviceKey      types.String `tfsdk:"device_key"`
	//state.SystemId = types.String{Value: string(systemInfo.Id)}
	//state.ManagementIp = types.String{Value: agentInfo.RunningConfig.ManagementIp}
	//state.AgentProfileId = types.String{Value: string(agentInfo.Config.Profile)}
	//state.OffBox = types.Bool{Value: bool(agentInfo.Config.AgentTypeOffBox)}

	// record device key and location if possible
	var deviceKey types.String
	if systemInfo != nil {
		if !state.DeviceKey.IsNull() {
			deviceKey = types.String{Value: systemInfo.DeviceKey}
		} else {
			deviceKey = types.String{Null: true}
		}
	} else {
		deviceKey = types.String{Null: true}
	}

	// Set state
	diags = resp.State.Set(ctx, &rManagedDevice{
		SystemId:       types.String{Value: string(agentInfo.Status.SystemId)},
		ManagementIp:   types.String{Value: agentInfo.Config.ManagementIp},
		AgentProfileId: types.String{Value: string(agentInfo.Config.Profile)},
		OffBox:         types.Bool{Value: bool(agentInfo.Config.AgentTypeOffBox)},
		AgentId:        types.String{Value: string(agentInfo.Id)},
		DeviceKey:      deviceKey,
	})
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (o *resourceManagedDevice) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get current state
	var state rManagedDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan rManagedDevice
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update agent as needed
	if state.AgentProfileId.Value != plan.AgentProfileId.Value {
		err := o.client.AssignAgentProfile(ctx, &goapstra.AssignAgentProfileRequest{
			SystemAgents: []goapstra.ObjectId{goapstra.ObjectId(state.AgentId.Value)},
			ProfileId:    goapstra.ObjectId(plan.AgentProfileId.Value),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"error updating managed device agent",
				fmt.Sprintf("error while updating managed device agent '%s' (%s) - %s",
					state.AgentId.Value, state.ManagementIp.Value, err.Error()),
			)
			return
		}
	}

	// Set state to match plan
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (o *resourceManagedDevice) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rManagedDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var agentDoesNotExist, systemDoesNotExist bool

	_, err := o.client.GetSystemAgent(ctx, goapstra.ObjectId(state.AgentId.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound) {
			agentDoesNotExist = true
		} else {
			resp.Diagnostics.AddError(
				"error pulling agent info",
				fmt.Sprintf("could not get info about agent '%s' - %s", state.AgentId.Value, err.Error()),
			)
			return
		}
	}

	_, err = o.client.GetSystemInfo(ctx, goapstra.SystemId(state.SystemId.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound) {
			systemDoesNotExist = true
		} else {
			resp.Diagnostics.AddError(
				"error pulling system info",
				fmt.Sprintf("could not get info about system '%s' - %s", state.SystemId.Value, err.Error()),
			)
			return
		}
	}

	if !agentDoesNotExist {
		err = o.client.DeleteSystemAgent(ctx, goapstra.ObjectId(state.AgentId.Value))
		if err != nil {
			var ace goapstra.ApstraClientErr
			if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound) {
				resp.Diagnostics.AddError(
					"error deleting device agent",
					fmt.Sprintf("device agent '%s' ('%s' '%s') delete error - %s",
						state.AgentId.Value, state.SystemId.Value, state.ManagementIp.Value, err.Error()))
			}
			return
		}
	}

	if !systemDoesNotExist {
		err = o.client.DeleteSystem(ctx, goapstra.SystemId(state.SystemId.Value))
		if err != nil {
			var ace goapstra.ApstraClientErr
			if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound) {
				resp.Diagnostics.AddError(
					"error deleting managed device",
					fmt.Sprintf("managed device '%s' ('%s') delete error - %s",
						state.SystemId.Value, state.ManagementIp.Value, err.Error()))
			}
			return
		}
	}
}

type rManagedDevice struct {
	AgentId        types.String `tfsdk:"agent_id"`
	SystemId       types.String `tfsdk:"system_id"`
	ManagementIp   types.String `tfsdk:"management_ip"`
	DeviceKey      types.String `tfsdk:"device_key"`
	AgentProfileId types.String `tfsdk:"agent_profile_id"`
	OffBox         types.Bool   `tfsdk:"off_box"`
}

func (o *rManagedDevice) validateAgentProfile(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	agentProfile, err := client.GetAgentProfile(ctx, goapstra.ObjectId(o.AgentProfileId.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			diags.AddAttributeError(
				path.Root("agent_profile_id"),
				"agent profile not found",
				fmt.Sprintf("agent profile '%s' does not exist", o.AgentProfileId.Value))
		}
		diags.AddError("error validating agent profile", err.Error())
		return
	}

	// require credentials (we can't automate login otherwise)
	if !agentProfile.HasUsername || !agentProfile.HasPassword {
		diags.AddAttributeError(
			path.Root("agent_profile_id"),
			"Agent Profile needs credentials",
			fmt.Sprintf("selected agent_profile_id '%s' (%s) must have credentials - please fix via Web UI",
				agentProfile.Label, agentProfile.Id))
	}
}
