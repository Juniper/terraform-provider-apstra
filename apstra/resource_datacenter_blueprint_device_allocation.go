package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ resource.ResourceWithConfigure = &resourceDeviceAllocation{}

type resourceDeviceAllocation struct {
	client *goapstra.Client
}

func (o *resourceDeviceAllocation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_blueprint_device_allocation"
}

func (o *resourceDeviceAllocation) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceDeviceAllocation) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource allocates a Managed Device (probably a switch) to a node role" +
			" (spine1, etc...) within a Blueprint.",
		Attributes: blueprint.DeviceAllocation{}.ResourceAttributes(),
	}
}

func (o *resourceDeviceAllocation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.PopulateDataFromGraphDb(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SetInterfaceMap(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDeviceAllocation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve values from state
	var state blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.GetCurrentSystemId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.GetCurrentInterfaceMapId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.GetCurrentDeviceProfileId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SystemNodeId.IsNull() || state.InterfaceMapId.IsNull() || state.DeviceProfileId.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDeviceAllocation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.PopulateDataFromGraphDb(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SetInterfaceMap(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDeviceAllocation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	// Retrieve values from state
	var state blueprint.DeviceAllocation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.InterfaceMapId = types.StringNull()
	state.SetInterfaceMap(ctx, o.client, &resp.Diagnostics)

	state.DeviceKey = types.StringNull()
	state.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
}
