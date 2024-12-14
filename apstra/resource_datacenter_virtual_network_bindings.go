package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resource.ResourceWithConfigure  = (*resourceDatacenterVirtualNetworkBindings)(nil)
	_ resource.ResourceWithModifyPlan = (*resourceDatacenterVirtualNetworkBindings)(nil)
	_ resourceWithSetDcBpClientFunc   = (*resourceDatacenterVirtualNetworkBindings)(nil)
	_ resourceWithSetBpLockFunc       = (*resourceDatacenterVirtualNetworkBindings)(nil)
	_ resourceWithSetClient           = (*resourceDatacenterVirtualNetworkBindings)(nil)
)

type resourceDatacenterVirtualNetworkBindings struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterVirtualNetworkBindings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_network_bindings"
}

func (o *resourceDatacenterVirtualNetworkBindings) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterVirtualNetworkBindings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource assigns Leaf Switches to a Virtual Network and sets " +
			"their interface addresses. It is intended for use with Apstra 5.0 and later configurations where Virtual Networks " +
			"are not bound to switches as a part of VN creation.",
		Attributes: blueprint.VirtualNetworkBinding{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterVirtualNetworkBindings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan blueprint.DatacenterVirtualNetwork
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

	err = bp.SetVirtualNetworkLeafBindings(ctx)
}

func (o *resourceDatacenterVirtualNetworkBindings) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterVirtualNetworkBindings) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
