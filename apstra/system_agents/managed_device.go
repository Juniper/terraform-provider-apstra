package systemAgents

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/utils"
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

func (o ManagedDevice) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"agent_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the Managed Device Agent.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the System managed by the Agent.",
			Computed:            true,
		},
		"management_ip": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Management IP address of the system managed by the Agent.",
			Computed:            true,
		},
		"device_key": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Key which uniquely identifies a System asset probably the serial number.",
			Computed:            true,
		},
		"agent_profile_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Agent Profile ID associated with the Agent.",
			Computed:            true,
		},
		"off_box": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the agent runs on the switch (true) or on an Apstra node (false).",
			Computed:            true,
		},
	}
}

func (o ManagedDevice) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"agent_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the Managed Device Agent.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the System onboarded by the Managed Device Agent.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"management_ip": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Management IP address of the System.",
			Optional:            true,
			Validators:          []validator.String{apstravalidator.ParseIpOrCidr(false, false)},
		},
		"device_key": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Key which uniquely identifies a System asset, probably a serial number.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"agent_profile_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Agent Profile associated with the Agent.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"off_box": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the agent runs on the switch (true) or on an Apstra node (false).",
			Optional:            true,
			Validators: []validator.Bool{boolvalidator.AtLeastOneOf(
				path.MatchRoot("filter").AtName("agent_id"),
				path.MatchRoot("filter").AtName("system_id"),
				path.MatchRoot("filter").AtName("management_ip"),
				path.MatchRoot("filter").AtName("device_key"),
				path.MatchRoot("filter").AtName("agent_profile_id"),
				path.MatchRoot("filter").AtName("off_box"),
			)},
		},
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
		if utils.IsApstra404(err) {
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
		if utils.IsApstra404(err) {
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

// IpNetFromManagementIp is generally called when a ManagedDevice object is used
// as a filter for matching other ManagedDevice objects. In that case, the
// ManagementIp element might contain a CIDR block rather than an individual
// host address. In either case, returns a *net.IPNet representing the
// ManagementIp element. nil is returned if the ManagementIp element is null.
func (o *ManagedDevice) IpNetFromManagementIp(_ context.Context, diags *diag.Diagnostics) *net.IPNet {
	var err error
	var ip net.IP
	var mgmtPrefix *net.IPNet

	if o.ManagementIp.IsNull() {
		return nil
	}

	// we don't know if we got a single host address (192.168.10.10) or a cidr
	// block (192.168.10.0/24). Try parsing as CIDR first. Not generating diags
	// on err because we don't necessarily expect this to work.
	_, mgmtPrefix, err = net.ParseCIDR(o.ManagementIp.ValueString())
	if err == nil && mgmtPrefix != nil {
		return mgmtPrefix
	}

	// parsing as CIDR failed. Parse as IP.
	ip = net.ParseIP(o.ManagementIp.ValueString())
	if ip == nil {
		diags.AddError(
			"failed to parse management IP",
			fmt.Sprintf("couldn't parse: %q", o.ManagementIp.ValueString()),
		)
		return nil
	}

	// construct a mask of the appropriate size
	var mask net.IPMask
	switch {
	case len(ip.To4()) == net.IPv4len:
		mask = net.CIDRMask(32, 32)
	case len(ip) == net.IPv6len:
		mask = net.CIDRMask(128, 128)
	default:
		diags.AddError(
			"unable to determine appropriate mask size for ip address",
			fmt.Sprintf("%q didn't match IPv4 nor IPv6 detection strategy", ip.String()))
	}

	return &net.IPNet{
		IP:   ip,
		Mask: mask,
	}
}
