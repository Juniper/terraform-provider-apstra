package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = (*resourceDatacenterRoutingZoneLoopbackAddresses)(nil)
	_ resource.ResourceWithValidateConfig = (*resourceDatacenterRoutingZoneLoopbackAddresses)(nil)
	_ resourceWithSetClient               = (*resourceDatacenterRoutingZoneLoopbackAddresses)(nil)
	_ resourceWithSetDcBpClientFunc       = (*resourceDatacenterRoutingZoneLoopbackAddresses)(nil)
	_ resourceWithSetBpLockFunc           = (*resourceDatacenterRoutingZoneLoopbackAddresses)(nil)
)

type resourceDatacenterRoutingZoneLoopbackAddresses struct {
	client          *apstra.Client
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_routing_zone_loopback_addresses"
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + fmt.Sprintf("This resource configures loopback interface "+
			"addresses of *switch* nodes in a Datacenter Blueprint.\n\n"+
			"Note that the loopback interface addresses within the `default` routing zone can also be configured "+
			"using the `apstra_datacenter_device_allocation` resource. Configuring loopback interface addresses using "+
			"both resources can lead to configuration churn, and should be avoided.\n\n"+
			"Note that loopback interface addresses can only be configured on switches *actively participating* in "+
			"the given Routing Zone. For Leaf Switch loopback interfaces in non-default Routing Zones, participation "+
			"requires that a Virtual Network belonging to the Routing Zone be bound to the Switch. The Terraform "+
			"project must be structured to ensure that those bindings exist before this resource is created or updated.\n\n"+
			"Requires Apstra %s.", compatibility.SecurityZoneLoopbackApiSupported),
		Attributes: blueprint.RoutingZoneLoopbacks{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) ValidateConfig(_ context.Context, _ resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if o.client == nil {
		return
	}

	if !compatibility.SecurityZoneLoopbackApiSupported.Check(version.Must(version.NewVersion(o.client.ApiVersion()))) {
		resp.Diagnostics.AddError(
			constants.ErrInvalidConfig,
			"this resource requires Apstra "+compatibility.SecurityZoneLoopbackApiSupported.String(),
		)
	}
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blueprint.RoutingZoneLoopbacks
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

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// create an API request
	request, ps := plan.Request(ctx, bp, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// record assignments to private state
	ps.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set loopback addresses
	err = bp.SetSecurityZoneLoopbacks(ctx, apstra.ObjectId(plan.RoutingZoneId.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("failed to set loopback addresses", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blueprint.RoutingZoneLoopbacks
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", state.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	api, err := bp.GetSecurityZoneInfo(ctx, apstra.ObjectId(state.RoutingZoneId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to get security zone info", err.Error())
		return
	}

	state.LoadApiData(ctx, api, req.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan blueprint.RoutingZoneLoopbacks
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

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// extract private state (previously configured loopbacks)
	var previousLoopbackMap private.ResourceDatacenterRoutingZoneLoopbackAddresses
	previousLoopbackMap.LoadPrivateState(ctx, req.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create an API request
	request, ps := plan.Request(ctx, bp, previousLoopbackMap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// record assignments to private state
	ps.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set loopback addresses
	err = bp.SetSecurityZoneLoopbacks(ctx, apstra.ObjectId(plan.RoutingZoneId.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("failed to set loopback addresses", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blueprint.RoutingZoneLoopbacks
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a plan with empty loopback address map
	plan := blueprint.RoutingZoneLoopbacks{
		BlueprintId:   state.BlueprintId,
		RoutingZoneId: state.RoutingZoneId,
		Loopbacks:     types.MapNull(types.ObjectType{AttrTypes: blueprint.RoutingZoneLoopback{}.AttrTypes()}),
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
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

	// extract private state (previously configured loopbacks)
	var previousLoopbackMap private.ResourceDatacenterRoutingZoneLoopbackAddresses
	previousLoopbackMap.LoadPrivateState(ctx, req.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create an API request
	request, _ := plan.Request(ctx, bp, previousLoopbackMap, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// clear loopback addresses
	err = bp.SetSecurityZoneLoopbacks(ctx, apstra.ObjectId(plan.RoutingZoneId.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("failed to set loopback addresses", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

func (o *resourceDatacenterRoutingZoneLoopbackAddresses) setClient(client *apstra.Client) {
	o.client = client
}
