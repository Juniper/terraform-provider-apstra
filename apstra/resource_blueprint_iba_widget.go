package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/iba"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceBlueprintIbaWidget{}
	_ resource.ResourceWithValidateConfig = &resourceBlueprintIbaWidget{}
	_ resourceWithSetDcBpClientFunc       = &resourceBlueprintIbaWidget{}
	_ resourceWithSetClient               = &resourceBlueprintIbaWidget{}
)

type resourceBlueprintIbaWidget struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	client          *apstra.Client
}

func (o *resourceBlueprintIbaWidget) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_iba_widget"
}

func (o *resourceBlueprintIbaWidget) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceBlueprintIbaWidget) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryRefDesignAny + "This resource creates an IBA Widget.\n\n" +
			"*Note: Compatible only with Apstra " + compatibility.BpIbaWidgetOk.String() + "*",

		Attributes: iba.Widget{}.ResourceAttributes(),
	}
}

func (o *resourceBlueprintIbaWidget) ValidateConfig(_ context.Context, _ resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// cannot proceed to api version validation if the provider has not been configured
	if o.client == nil {
		return
	}

	// only supported with Apstra 4.x
	if !compatibility.BpIbaWidgetOk.Check(version.Must(version.NewVersion(o.client.ApiVersion()))) {
		resp.Diagnostics.AddError(
			"Incompatible API version",
			"This resource is compatible only with Apstra "+compatibility.BpIbaWidgetOk.String(),
		)
	}
}

func (o *resourceBlueprintIbaWidget) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan iba.Widget
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Convert the plan into an API Request
	probeReq := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := bp.CreateIbaWidget(ctx, probeReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create IBA Widget", err.Error())
		return
	}

	// Set state
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceBlueprintIbaWidget) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iba.Widget
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	api, err := bp.GetIbaWidget(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to Read IBA Widget", err.Error())
		return
	}

	// create new state object
	state.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceBlueprintIbaWidget) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan iba.Widget
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	widgetReq := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Update IBA Widget
	err = bp.UpdateIbaWidget(ctx, apstra.ObjectId(plan.Id.ValueString()), widgetReq)
	if err != nil {
		resp.Diagnostics.AddError("error updating IBA Widget", err.Error())
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceBlueprintIbaWidget) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iba.Widget
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Delete IBA Probe by calling API
	err = bp.DeleteIbaWidget(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting IBA Probe", err.Error())
		return
	}
}

func (o *resourceBlueprintIbaWidget) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

// setClient is used for API version compatibility check only
func (o *resourceBlueprintIbaWidget) setClient(client *apstra.Client) {
	o.client = client
}
