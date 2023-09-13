package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	systemAgents "github.com/Juniper/terraform-provider-apstra/apstra/system_agents"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceManagedDeviceAck{}

type resourceManagedDeviceAck struct {
	client *apstra.Client
}

func (o *resourceManagedDeviceAck) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_device_ack"
}

func (o *resourceManagedDeviceAck) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceManagedDeviceAck) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource *acknowledges* the System " +
			"(probably a switch) discovered by a running System Agent. The " +
			"acknowledgement of a System cannot be modified nor deleted. " +
			"Any modification to the inputs of this resource will cause it " +
			"to be removed from the Terraform state and recreated. Modifying " +
			"or deleting this resource has no effect on Apstra.",
		Attributes: systemAgents.SystemAck{}.ResourceAttributes(),
	}
}

func (o *resourceManagedDeviceAck) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan systemAgents.SystemAck
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// figure out the system ID
	agentInfo, err := o.client.GetSystemAgent(ctx, apstra.ObjectId(plan.AgentId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"error fetching Agent info",
			err.Error())
		return
	}

	// figure out the actual serial number (device_key)
	systemInfo, err := o.client.GetSystemInfo(ctx, agentInfo.Status.SystemId)
	if err != nil {
		resp.Diagnostics.AddError(
			"error fetching system info",
			err.Error())
	}

	// validate discovered device_key (serial number)
	if plan.DeviceKey.ValueString() != systemInfo.DeviceKey {
		resp.Diagnostics.AddAttributeError(
			path.Root("device_key"),
			"system device_key mismatch",
			fmt.Sprintf("config expects switch device_key %q, device reports %q",
				plan.DeviceKey.ValueString(), systemInfo.DeviceKey),
		)
		return
	}

	plan.SystemId = types.StringValue(string(agentInfo.Status.SystemId))
	plan.Acknowledge(ctx, systemInfo, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceManagedDeviceAck) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state systemAgents.SystemAck
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Set current state as final state - changes will never be detected.
	resp.Diagnostics.Append(req.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceManagedDeviceAck) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Because Read() is a no-op, Update() should never be called.
	resp.Diagnostics.Append(validatordiag.BugInProviderDiagnostic(
		"resourceManagedDeviceAck.Update() should never be called",
	))
}

// Delete resource
func (o *resourceManagedDeviceAck) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nothing to do. The Terraform Plugin Framework will remove the resource
	// from the terraform state for us. "ACK" is a one-way street.
}
