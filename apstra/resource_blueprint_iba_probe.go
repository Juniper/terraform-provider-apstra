package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/iba"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resource.ResourceWithConfigure = &resourceBlueprintIbaProbe{}
var _ resourceWithSetDcBpClientFunc = &resourceBlueprintIbaProbe{}

type resourceBlueprintIbaProbe struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *resourceBlueprintIbaProbe) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_iba_probe"
}

func (o *resourceBlueprintIbaProbe) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceBlueprintIbaProbe) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This resource creates an IBA Probe within a Blueprint.",
		Attributes:          iba.Probe{}.ResourceAttributes(),
	}
}

func (o *resourceBlueprintIbaProbe) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan iba.Probe
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bpClient, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}
	var id apstra.ObjectId
	// Convert the plan into an API Request
	if plan.PredefinedProbeId.IsUnknown() || plan.PredefinedProbeId.IsNull() {
		// Create Probe from Json
		id, err = bpClient.CreateIbaProbeFromJson(ctx, []byte(plan.ProbeJson.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("failed to create Iba Probe", err.Error())
			return
		}
	} else {
		probeReq := &apstra.IbaPredefinedProbeRequest{
			Name: plan.PredefinedProbeId.ValueString(),
			Data: []byte(plan.ProbeConfig.ValueString()),
		}
		if resp.Diagnostics.HasError() {
			return
		}

		// Instantiate the probe
		id, err = bpClient.InstantiateIbaPredefinedProbe(ctx, probeReq)
		if err != nil {
			resp.Diagnostics.AddError("failed to create Iba Probe", err.Error())
			return
		}
	}
	// Fetch the probe details
	api, err := bpClient.GetIbaProbe(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read IBA Probe", err.Error())
		return
	}

	// Populate plan object with new probe details
	plan.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceBlueprintIbaProbe) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iba.Probe
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bpClient, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// read probe details from API
	api, err := bpClient.GetIbaProbe(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to Read IBA Probe", err.Error())
		return
	}

	// load probe details into state object
	state.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceBlueprintIbaProbe) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	resp.Diagnostics.AddError("not implemented", "Probe update not implemented via terraform. Please file a bug.")
}

// Delete resource
func (o *resourceBlueprintIbaProbe) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iba.Probe
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bpClient, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Delete IBA Probe by calling API
	err = bpClient.DeleteIbaProbe(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting IBA Probe", err.Error())
		return
	}
}

func (o *resourceBlueprintIbaProbe) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}
