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

	// Ensure the following are populated:o//   - SystemNodeId (from node_name)
	//   - plan.SystemNodeId
	//   - plan.InterfaceMapCatalogId
	//   - plan.DeviceProfileNodeId
	plan.PopulateDataFromGraphDb(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.BlueprintId.IsNull() {
		resp.Diagnostics.AddError("blueprint does not exist", "blueprint vanished while we were working on it")
	}

	plan.SetInterfaceMap(ctx, o.client, &resp.Diagnostics)
	if plan.BlueprintId.IsNull() {
		resp.Diagnostics.AddError("blueprint does not exist", "blueprint vanished while we were working on it")
	}
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

	state.ReadSystemNode(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	state.GetCurrentInterfaceMapId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	state.GetCurrentDeviceProfileId(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.BlueprintId.IsNull() || state.NodeId.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDeviceAllocation) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Update not needed because any change triggers replacement
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

	state.InterfaceMapCatalogId = types.StringNull()
	state.SetInterfaceMap(ctx, o.client, &resp.Diagnostics)

	state.DeviceKey = types.StringNull()
	state.SetNodeSystemId(ctx, o.client, &resp.Diagnostics)
}
