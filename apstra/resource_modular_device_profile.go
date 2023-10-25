package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/device"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

var _ resource.ResourceWithConfigure = &resourceModularDeviceProfile{}
var _ resource.ResourceWithValidateConfig = &resourceModularDeviceProfile{}

type resourceModularDeviceProfile struct {
	client *apstra.Client
}

func (o *resourceModularDeviceProfile) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config
	var config device.ModularDeviceProfile
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for key := range config.LineCardProfileIds.Elements() {
		_, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("line_card_profile_ids").AtMapKey(key),
				"Slot Numbers must be positive numbers",
				fmt.Sprintf("expected positive integer values for chassis slot numbers, got %q", key),
			)
		}
	}
}

func (o *resourceModularDeviceProfile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_modular_device_profile"
}

func (o *resourceModularDeviceProfile) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceModularDeviceProfile) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates a Device Profile for a chassis-based device",
		Attributes:          device.ModularDeviceProfile{}.ResourceAttributes(),
	}
}

func (o *resourceModularDeviceProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan device.ModularDeviceProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateModularDeviceProfile(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating modular device profile", err.Error())
		return
	}

	// id is not in the apstra.InterfaceMapData object we're using, so set it directly
	plan.Id = types.StringValue(id.String())

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceModularDeviceProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state device.ModularDeviceProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get device profile from api
	apiData, err := o.client.GetModularDeviceProfile(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading Modular Device Profile", err.Error())
			return
		}
	}

	// load api data into state object
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceModularDeviceProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan device.ModularDeviceProfile
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.UpdateModularDeviceProfile(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Modular Device Profile", err.Error())
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceModularDeviceProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state device.ModularDeviceProfile
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Interface Map by calling API
	err := o.client.DeleteModularDeviceProfile(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			"error deleting Modular Device Profile", err.Error())
		return
	}
}
