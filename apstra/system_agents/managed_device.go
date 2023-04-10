package systemAgents

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
)

type ManagedDevice struct {
	AgentId        types.String `tfsdk:"agent_id"`
	SystemId       types.String `tfsdk:"system_id"`
	ManagementIp   types.String `tfsdk:"management_ip"`
	DeviceKey      types.String `tfsdk:"device_key"`
	AgentProfileId types.String `tfsdk:"agent_profile_id"`
	OffBox         types.Bool   `tfsdk:"off_box"`
}

func (o *ManagedDevice) Request(_ context.Context, _ *diag.Diagnostics) *apstra.SystemAgentRequest {
	return &apstra.SystemAgentRequest{
		AgentTypeOffbox: apstra.AgentTypeOffbox(o.OffBox.ValueBool()),
		ManagementIp:    o.ManagementIp.ValueString(),
		Profile:         apstra.ObjectId(o.AgentProfileId.ValueString()),
		OperationMode:   apstra.AgentModeFull,
	}
}

func (o ManagedDevice) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"agent_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the Managed Device Agent.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"system_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the System onboarded by the Managed Device Agent.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"management_ip": resourceSchema.StringAttribute{
			MarkdownDescription: "Management IP address of the system.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{apstravalidator.ParseIp(false, false)},
		},
		"device_key": resourceSchema.StringAttribute{
			MarkdownDescription: "Key which uniquely identifies a System asset. Possibly a MAC address or serial number.",
			Optional:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"agent_profile_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Agent Profile used when instantiating the Agent. An Agent Profile is" +
				"required to specify the login credentials and platform type.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"off_box": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates that an *offbox* agent should be created (required for Junos devices, default: `true`)",
			Required:            true,
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
	}
}

func (o *ManagedDevice) LoadApiData(_ context.Context, in *apstra.SystemAgent, _ *diag.Diagnostics) {
	o.SystemId = types.StringValue(string(in.Status.SystemId))
	o.ManagementIp = types.StringValue(in.Config.ManagementIp)
	o.AgentProfileId = types.StringValue(string(in.Config.Profile))
	o.OffBox = types.BoolValue(bool(in.Config.AgentTypeOffBox))
	o.AgentId = types.StringValue(string(in.Id))
}

func (o *ManagedDevice) ValidateAgentProfile(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	agentProfile, err := client.GetAgentProfile(ctx, apstra.ObjectId(o.AgentProfileId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			diags.AddAttributeError(
				path.Root("agent_profile_id"),
				"agent profile not found",
				fmt.Sprintf("agent profile %q does not exist", o.AgentProfileId.ValueString()))
		}
		diags.AddError("error validating agent profile", err.Error())
		return
	}

	// require credentials (we can't automate login otherwise)
	if !agentProfile.HasCredentials() {
		diags.AddAttributeError(
			path.Root("agent_profile_id"),
			"Agent Profile needs credentials",
			fmt.Sprintf("selected agent_profile_id %q (%s) must have credentials - please fix via Web UI",
				agentProfile.Label, agentProfile.Id))
	}

	// require platform (assignment will fail without platform)
	if agentProfile.Platform == "" {
		diags.AddAttributeError(
			path.Root("agent_profile_id"),
			"Agent Profile needs platform",
			fmt.Sprintf("selected agent_profile_id %q (%s) must specify the platform type",
				agentProfile.Label, agentProfile.Id))
	}
}

func (o *ManagedDevice) Acknowledge(ctx context.Context, si *apstra.ManagedSystemInfo, client *apstra.Client, diags *diag.Diagnostics) {
	// update with new SystemUserConfig
	err := client.UpdateSystem(ctx, apstra.SystemId(o.SystemId.ValueString()), &apstra.SystemUserConfig{
		AosHclModel: si.Facts.AosHclModel,
		AdminState:  apstra.SystemAdminStateNormal,
	})
	if err != nil {
		diags.AddError(
			"error updating managed device",
			fmt.Sprintf("unexpected error while updating user config: %s", err.Error()),
		)
		return
	}

}

func (o *ManagedDevice) GetDeviceKey(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	// Get SystemInfo from API
	systemInfo, err := client.GetSystemInfo(ctx, apstra.SystemId(o.SystemId.ValueString()))
	if err != nil {
		var ace apstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
		} else {
			diags.AddError(
				"error reading managed device system info",
				fmt.Sprintf("Could not Read %q (%s) - %s", o.SystemId.ValueString(), o.ManagementIp.ValueString(), err),
			)
			return
		}
	}

	// record device key and location if possible
	if systemInfo == nil {
		o.DeviceKey = types.StringNull()
	} else {
		o.DeviceKey = types.StringValue(systemInfo.DeviceKey)
	}
}
