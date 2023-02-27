package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceLogicalDevice{}
var _ resource.ResourceWithValidateConfig = &resourceLogicalDevice{}

type resourceLogicalDevice struct {
	client *goapstra.Client
}

func (o *resourceLogicalDevice) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logical_device"
}

func (o *resourceLogicalDevice) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = resourceGetClient(ctx, req, resp)
}

func (o *resourceLogicalDevice) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an IPv4 resource pool",
		Attributes:          logicalDevice{}.resourceAttributes(),
	}
}

func (o *resourceLogicalDevice) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config logicalDevice
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// extract []logicalDevicePanel from the resourceLogicalDevice
	panels := config.panels(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// validate each panel
	for i, panel := range panels {
		panel.validate(ctx, i, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (o *resourceLogicalDevice) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan logicalDevice
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateLogicalDevice(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Logical Device", err.Error())
	}

	plan.Id = types.StringValue(string(id))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceLogicalDevice) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state logicalDevice
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Logical Device from API and then update what is in state from what the API returns
	ld, err := o.client.GetLogicalDevice(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"error reading Logical Device",
			fmt.Sprintf("Could not Read %q - %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	// Create new state object
	var newState logicalDevice
	newState.Id = types.StringValue(string(ld.Id))
	newState.loadApiData(ctx, ld.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceLogicalDevice) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan logicalDevice
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Logical Device
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var ace goapstra.ApstraClientErr
	err := o.client.UpdateLogicalDevice(ctx, goapstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
			resp.State.RemoveResource(ctx)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError("error updating Logical Device", err.Error())
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceLogicalDevice) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state logicalDevice
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Logical Device by calling API
	err := o.client.DeleteLogicalDevice(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError("error deleting Logical Device", err.Error())
			return
		}
	}
}
