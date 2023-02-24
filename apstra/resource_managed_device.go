package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (o *resourceManagedDevice) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates/installs an Agent for an Apstra Managed Device." +
			"Optionally, it will 'Acknolwedge' the discovered system if the `device key` (serial number)" +
			"reported by the agent matches the optional `device_key` field.",
		Attributes: map[string]schema.Attribute{
			"agent_id": schema.StringAttribute{
				MarkdownDescription: "Apstra ID for the Managed Device Agent.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"system_id": schema.StringAttribute{
				MarkdownDescription: "Apstra ID for the System onboarded by the Managed Device Agent.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"management_ip": schema.StringAttribute{
				MarkdownDescription: "Management IP address of the system.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:          []validator.String{parseIp(false, false)},
			},
			"device_key": schema.StringAttribute{
				MarkdownDescription: "Key which uniquely identifies a System asset. Possibly a MAC address or serial number.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"agent_profile_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Agent Profile used when instantiating the Agent. An Agent Profile is" +
					"required to specify the login credentials and platform type.",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"off_box": schema.BoolAttribute{
				MarkdownDescription: "Indicates that an 'Offbox' agent should be created (required for Junos devices)",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.validateAgentProfile(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Agent for this Managed Device
	agentId, err := o.client.CreateAgent(ctx, &goapstra.SystemAgentRequest{
		AgentTypeOffbox: goapstra.AgentTypeOffbox(plan.OffBox.ValueBool()),
		ManagementIp:    plan.ManagementIp.ValueString(),
		Profile:         goapstra.ObjectId(plan.AgentProfileId.ValueString()),
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
			fmt.Sprintf("Could not run 'install' job on new agent %q", agentId),
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
		if plan.DeviceKey.ValueString() != systemInfo.DeviceKey {
			resp.Diagnostics.AddAttributeError(
				path.Root("device_key"),
				"error system device_key mismatch",
				fmt.Sprintf("config expects switch device_key %q, device reports %q",
					plan.DeviceKey.ValueString(), systemInfo.DeviceKey),
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

	// create new state object
	state := rManagedDevice{
		AgentId:        types.StringValue(string(agentId)),
		SystemId:       types.StringValue(string(agentInfo.Status.SystemId)),
		ManagementIp:   types.StringValue(agentInfo.Config.ManagementIp),
		DeviceKey:      types.StringValue(systemInfo.DeviceKey),
		AgentProfileId: types.StringValue(string(agentInfo.Config.Profile)),
		OffBox:         types.BoolValue(bool(agentInfo.Config.AgentTypeOffBox)),
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AgentInfo from API
	agentInfo, err := o.client.GetSystemAgent(ctx, goapstra.ObjectId(state.AgentId.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading managed device agent info",
				fmt.Sprintf("Could not Read %q (%s)- %s", state.AgentId.ValueString(), state.ManagementIp.ValueString(), err),
			)
			return
		}
	}

	var newState rManagedDevice
	newState.loadApiResponse(ctx, agentInfo, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newState.getDeviceKey(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// not currently clear why we need to copy this from old state
	if newState.DeviceKey.IsNull() && !state.DeviceKey.IsNull() {
		newState.DeviceKey = types.StringValue(state.DeviceKey.ValueString())
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
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
	if state.AgentProfileId.ValueString() != plan.AgentProfileId.ValueString() {
		err := o.client.AssignAgentProfile(ctx, &goapstra.AssignAgentProfileRequest{
			SystemAgents: []goapstra.ObjectId{goapstra.ObjectId(state.AgentId.ValueString())},
			ProfileId:    goapstra.ObjectId(plan.AgentProfileId.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"error updating managed device agent",
				fmt.Sprintf("error while updating managed device agent %q (%s) - %s",
					state.AgentId.ValueString(), state.ManagementIp.ValueString(), err.Error()),
			)
			return
		}
	}

	// set state to match plan
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

	_, err := o.client.GetSystemAgent(ctx, goapstra.ObjectId(state.AgentId.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound) {
			agentDoesNotExist = true
		} else {
			resp.Diagnostics.AddError(
				"error pulling agent info",
				fmt.Sprintf("could not get info about agent %q - %s", state.AgentId.ValueString(), err.Error()),
			)
			return
		}
	}

	_, err = o.client.GetSystemInfo(ctx, goapstra.SystemId(state.SystemId.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound) {
			systemDoesNotExist = true
		} else {
			resp.Diagnostics.AddError(
				"error pulling system info",
				fmt.Sprintf("could not get info about system %q - %s", state.SystemId.ValueString(), err.Error()),
			)
			return
		}
	}

	if !agentDoesNotExist {
		err = o.client.DeleteSystemAgent(ctx, goapstra.ObjectId(state.AgentId.ValueString()))
		if err != nil {
			var ace goapstra.ApstraClientErr
			if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound) {
				resp.Diagnostics.AddError(
					"error deleting device agent",
					fmt.Sprintf("device agent %q (%q %q) delete error - %s",
						state.AgentId.ValueString(), state.SystemId.ValueString(), state.ManagementIp.ValueString(), err.Error()))
			}
			return
		}
	}

	if !systemDoesNotExist {
		err = o.client.DeleteSystem(ctx, goapstra.SystemId(state.SystemId.ValueString()))
		if err != nil {
			var ace goapstra.ApstraClientErr
			if !(errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound) {
				resp.Diagnostics.AddError(
					"error deleting managed device",
					fmt.Sprintf("managed device %q (%s) delete error - %s",
						state.SystemId.ValueString(), state.ManagementIp.ValueString(), err.Error()))
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

func (o *rManagedDevice) loadApiResponse(ctx context.Context, in *goapstra.SystemAgent, diags *diag.Diagnostics) {
	o.SystemId = types.StringValue(string(in.Status.SystemId))
	o.ManagementIp = types.StringValue(in.Config.ManagementIp)
	o.AgentProfileId = types.StringValue(string(in.Config.Profile))
	o.OffBox = types.BoolValue(bool(in.Config.AgentTypeOffBox))
	o.AgentId = types.StringValue(string(in.Id))
}

func (o *rManagedDevice) getDeviceKey(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	// Get SystemInfo from API
	systemInfo, err := client.GetSystemInfo(ctx, goapstra.SystemId(o.SystemId.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
		} else {
			diags.AddError(
				"error reading managed device system info",
				fmt.Sprintf("Could not Read %q (%s) - %s", o.SystemId.ValueString(), o.ManagementIp.ValueString(), err),
			)
			return
		}
	}

	// record device key and location if possible
	if systemInfo != nil {
		o.DeviceKey = types.StringNull()
	} else {
		o.DeviceKey = types.StringValue(systemInfo.DeviceKey)
	}
}

func (o *rManagedDevice) validateAgentProfile(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	agentProfile, err := client.GetAgentProfile(ctx, goapstra.ObjectId(o.AgentProfileId.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			diags.AddAttributeError(
				path.Root("agent_profile_id"),
				"agent profile not found",
				fmt.Sprintf("agent profile %q does not exist", o.AgentProfileId.ValueString()))
		}
		diags.AddError("error validating agent profile", err.Error())
		return
	}

	// require credentials (we can't automate login otherwise)
	if !agentProfile.HasUsername || !agentProfile.HasPassword {
		diags.AddAttributeError(
			path.Root("agent_profile_id"),
			"Agent Profile needs credentials",
			fmt.Sprintf("selected agent_profile_id %q (%s) must have credentials - please fix via Web UI",
				agentProfile.Label, agentProfile.Id))
	}
}
