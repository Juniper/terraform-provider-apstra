package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	systemAgentStateConnected = "connected"
	sleepMilliSeconds         = 100
)

type resourceManagedDeviceType struct{}

func (r resourceManagedDevice) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "apstra_managed_device"
}

func (r resourceManagedDevice) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"agent_id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"system_id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"management_ip": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"device_key": {
				Type:          types.StringType,
				Optional:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"agent_profile_id": {
				Type:     types.StringType,
				Required: true,
			},
			"agent_label": {
				Type:     types.StringType,
				Optional: true,
			},
			"off_box": {
				Type:          types.BoolType,
				Computed:      true,
				Optional:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace(), resource.UseStateForUnknown()},
			},
			"location": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (r resourceManagedDeviceType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceManagedDevice{
		p: *(p.(*Provider)),
	}, nil
}

type resourceManagedDevice struct {
	p Provider
}

func (r resourceManagedDevice) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg ResourceManagedDevice
	req.Config.Get(ctx, &cfg)
	if cfg.DeviceKey.IsNull() && !cfg.Location.IsNull() {
		resp.Diagnostics.AddError(
			"invalid configuration",
			"element 'location' requires setting element 'device_key' - there's API reasons for this, but aside from that... Do you really know where something is, without knowing *what* it is?")
	}
}

func (r resourceManagedDevice) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceManagedDevice
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// look up agent profile info
	agentProfile, err := r.p.client.GetAgentProfile(ctx, goapstra.ObjectId(plan.AgentProfileId.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent",
			"Could not create, unexpected error: "+err.Error(),
		)
		return
	}

	// require credentials (we can't automate login otherwise)
	if !agentProfile.HasUsername || !agentProfile.HasPassword {
		resp.Diagnostics.AddWarning(
			"Managed Device Agent Profile Credentials",
			fmt.Sprintf("selected agent_profile_id is '%s' (%s) missing credentials - please fix via Web UI",
				agentProfile.Label, plan.AgentProfileId.Value),
		)
	}

	// Create new Agent for this Managed Device
	agentId, err := r.p.client.CreateAgent(ctx, &goapstra.SystemAgentRequest{
		AgentTypeOffbox: goapstra.AgentTypeOffbox(plan.OffBox.Value),
		ManagementIp:    plan.ManagementIp.Value,
		Profile:         goapstra.ObjectId(plan.AgentProfileId.Value),
		OperationMode:   goapstra.AgentModeFull,
		Label:           plan.AgentLabel.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent",
			"Could not create, unexpected error: "+err.Error(),
		)
		return
	}

	// Install the new agent
	_, err = r.p.client.SystemAgentRunJob(ctx, agentId, goapstra.AgentJobTypeInstall)
	if err != nil {
		resp.Diagnostics.AddError(
			"error running Install job",
			fmt.Sprintf("Could not run 'install' job on new agent '%s', unexpected error: %s",
				agentId, err.Error()),
		)
		return
	}

	// figure out the new switch system Id
	agentInfo, err := r.p.client.GetSystemAgent(ctx, agentId)
	if err != nil {
		resp.Diagnostics.AddError(
			"error fetching agent info",
			fmt.Sprintf("Could not fetch info from new agent '%s', unexpected error: %s",
				agentId, err.Error()),
		)
		return
	}

	// figure out the new switch serial number (device_key)
	systemInfo, err := r.p.client.GetSystemInfo(ctx, agentInfo.Status.SystemId)
	if err != nil {
		resp.Diagnostics.AddError(
			"error fetching system info",
			fmt.Sprintf("Could not fetch info from new system '%s', unexpected error: %s",
				agentInfo.Status.SystemId, err.Error()),
		)
		return
	}

	// submit a SystemUserConfig only if device_key was supplied
	if !plan.DeviceKey.IsNull() && !plan.DeviceKey.IsUnknown() {
		// mismatched device key is fatal
		if plan.DeviceKey.Value != systemInfo.DeviceKey {
			resp.Diagnostics.AddError(
				"error system mismatch",
				fmt.Sprintf("config expects switch device_key '%s', device reports '%s'",
					plan.DeviceKey.Value, systemInfo.DeviceKey),
			)
			return
		}

		// update with new SystemUserConfig
		err = r.p.client.UpdateSystem(ctx, agentInfo.Status.SystemId, &goapstra.SystemUserConfig{
			AosHclModel: systemInfo.Facts.AosHclModel,
			AdminState:  goapstra.SystemAdminStateNormal,
			Location:    plan.Location.Value,
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

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceManagedDevice) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ResourceManagedDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AgentInfo from API
	agentInfo, err := r.p.client.GetSystemAgent(ctx, goapstra.ObjectId(state.AgentId.Value))
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
	systemInfo, err := r.p.client.GetSystemInfo(ctx, goapstra.SystemId(state.SystemId.Value))
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

	var agentLabel types.String
	if agentInfo.RunningConfig.Label == "" {
		agentLabel = types.String{Null: true}
	} else {
		agentLabel = types.String{Value: agentInfo.RunningConfig.Label}
	}

	state.SystemId = types.String{Value: string(systemInfo.Id)}
	state.ManagementIp = types.String{Value: agentInfo.RunningConfig.ManagementIp}
	state.AgentProfileId = types.String{Value: string(agentInfo.Config.Profile)}
	state.AgentLabel = agentLabel
	state.OffBox = types.Bool{Value: bool(agentInfo.Config.AgentTypeOffBox)}

	// record device key and location if possible
	if systemInfo != nil {
		if !state.DeviceKey.IsNull() {
			state.DeviceKey = types.String{Value: systemInfo.DeviceKey}
		}
		if !state.Location.IsNull() {
			state.Location = types.String{Value: systemInfo.UserConfig.Location}
		}
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceManagedDevice) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get current state
	var state ResourceManagedDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan ResourceManagedDevice
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update agent as needed
	if state.AgentProfileId.Value != plan.AgentProfileId.Value || state.AgentLabel.Value != plan.AgentLabel.Value {
		err := r.p.client.UpdateSystemAgent(ctx, goapstra.ObjectId(state.AgentId.Value), &goapstra.SystemAgentRequest{
			Profile: goapstra.ObjectId(plan.AgentProfileId.Value),
			Label:   plan.AgentLabel.Value,
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

	// update system as needed
	if state.Location.Value != plan.Location.Value {
		// 'location' is an element of user config, which (swagger says) doesn't support PATCH.
		// fetch the whole system info, which contains the user config we need
		systemInfo, err := r.p.client.GetSystemInfo(ctx, goapstra.SystemId(state.SystemId.Value))
		if err != nil {
			resp.Diagnostics.AddError(
				"error fetching managed device info",
				fmt.Sprintf("error while reading managed device system info '%s' (%s) - %s",
					state.SystemId.Value, state.ManagementIp.Value, err.Error()),
			)
			return
		}

		// update the user config structure
		systemInfo.UserConfig.Location = plan.Location.Value

		err = r.p.client.UpdateSystem(ctx, goapstra.SystemId(state.SystemId.Value), &systemInfo.UserConfig)
		if err != nil {
			resp.Diagnostics.AddError(
				"error updating managed device user config info",
				fmt.Sprintf("error while updating managed device user config info '%s' (%s) - %s",
					state.SystemId.Value, state.ManagementIp.Value, err.Error()),
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
func (r resourceManagedDevice) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceManagedDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var agentDoesNotExist, systemDoesNotExist bool

	_, err := r.p.client.GetSystemAgent(ctx, goapstra.ObjectId(state.AgentId.Value))
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

	_, err = r.p.client.GetSystemInfo(ctx, goapstra.SystemId(state.SystemId.Value))
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
		err = r.p.client.DeleteSystemAgent(ctx, goapstra.ObjectId(state.AgentId.Value))
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
		err = r.p.client.DeleteSystem(ctx, goapstra.SystemId(state.SystemId.Value))
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
