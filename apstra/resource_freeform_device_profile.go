package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/freeform"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure = &resourceFreeformDeviceProfile{}
	_ resourceWithSetFfBpClientFunc  = &resourceFreeformDeviceProfile{}
	_ resourceWithSetBpLockFunc      = &resourceFreeformDeviceProfile{}
)

type resourceFreeformDeviceProfile struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceFreeformDeviceProfile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_device_profile"
}

func (o *resourceFreeformDeviceProfile) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceFreeformDeviceProfile) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This resource imports a Device Profile in a Freeform Blueprint from the Global Catalog.",
		Attributes:          freeform.FreeformDeviceProfile{}.ResourceAttributes(),
	}
}

func (o *resourceFreeformDeviceProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan freeform.FreeformDeviceProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	id, err := bp.ImportDeviceProfile(ctx, apstra.ObjectId(plan.DeviceProfileId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error importing DeviceProfile", err.Error())
		return
	}

	plan.Id = types.StringValue(id.String())

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformDeviceProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state freeform.FreeformDeviceProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	err = bp.Client().GetNode(ctx, bp.Id(), apstra.ObjectId(state.Id.ValueString()), nil)
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving DeviceProfile", err.Error())
		return
	}
}

func (o *resourceFreeformDeviceProfile) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (o *resourceFreeformDeviceProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state freeform.FreeformDeviceProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// Delete Config Template by calling API
	err = bp.DeleteDeviceProfile(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Config Template", err.Error())
		return
	}
}

func (o *resourceFreeformDeviceProfile) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceFreeformDeviceProfile) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
