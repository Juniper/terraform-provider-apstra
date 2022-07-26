package apstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	systemAgentStateConnected = "connected"
	sleepMilliSeconds         = 100
)

type resourceManagedDeviceType struct{}

func (r resourceManagedDeviceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"agent_id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			},
			"system_id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			},
			"management_ip": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
			"device_key": {
				Type:          types.StringType,
				Optional:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
			"agent_profile_id": {
				Type:     types.StringType,
				Required: true,
			},
			"agent_label": {
				Type:     types.StringType,
				Optional: true,
			},
			"on_box": {
				Type:          types.BoolType,
				Computed:      true,
				Optional:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace(), tfsdk.UseStateForUnknown()},
			},
			"location": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (r resourceManagedDeviceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceManagedDevice{
		p: *(p.(*provider)),
	}, nil
}

type resourceManagedDevice struct {
	p provider
}

func (r resourceManagedDevice) ValidateConfig(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	var cfg ResourceManagedDevice
	req.Config.Get(ctx, &cfg)
	if cfg.DeviceKey.IsNull() && !cfg.Location.IsNull() {
		resp.Diagnostics.AddError(
			"invalid configuration",
			"element 'location' requires setting element 'device_key' - there's API reasons for this, but aside from that... Do you really know where something is, without knowing *what* it is?")
	}
}

func (r resourceManagedDevice) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
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

	var agentType goapstra.AgentType
	if plan.OnBox.Value {
		agentType = goapstra.AgentTypeOnbox
	} else {
		plan.OnBox = types.Bool{Value: false}
		agentType = goapstra.AgentTypeOffbox
	}

	// Create new Agent for this Managed Device
	agentId, err := r.p.client.CreateAgent(ctx, &goapstra.AgentCfg{
		AgentType:     agentType,
		ManagementIp:  plan.ManagementIp.Value,
		Profile:       goapstra.ObjectId(plan.AgentProfileId.Value),
		OperationMode: goapstra.AgentModeFull,
		Label:         plan.AgentLabel.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent",
			"Could not create, unexpected error: "+err.Error(),
		)
		return
	}

	// Install the new agent
	_, err = r.p.client.AgentRunJob(ctx, agentId, goapstra.AgentJobTypeInstall)
	if err != nil {
		resp.Diagnostics.AddError(
			"error running Install job",
			fmt.Sprintf("Could not run 'install' job on new agent '%s', unexpected error: %s",
				agentId, err.Error()),
		)
		return
	}

	// figure out the new switch system Id
	agentInfo, err := r.p.client.GetAgentInfo(ctx, agentId)
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
				fmt.Sprintf("Could config expects switch device_key '%s', device reports %s",
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

func (r resourceManagedDevice) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ResourceManagedDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AgentInfo from API
	agentInfo, err := r.p.client.GetAgentInfo(ctx, goapstra.ObjectId(state.AgentId.Value))
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
	state.OnBox = types.Bool{}

	if agentInfo.RunningConfig.AgentType == goapstra.AgentTypeOnbox {
		state.OnBox.Value = true
	}

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
func (r resourceManagedDevice) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
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
		err := r.p.client.UpdateAgent(ctx, goapstra.ObjectId(state.AgentId.Value), &goapstra.AgentCfg{
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
func (r resourceManagedDevice) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ResourceManagedDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var agentDoesNotExist, systemDoesNotExist bool

	_, err := r.p.client.GetAgentInfo(ctx, goapstra.ObjectId(state.AgentId.Value))
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
		err = r.p.client.DeleteAgent(ctx, goapstra.ObjectId(state.AgentId.Value))
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
