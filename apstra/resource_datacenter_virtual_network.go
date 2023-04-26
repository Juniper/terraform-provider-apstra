package tfapstra

//
//import (
//	"context"
//	"fmt"
//	"github.com/Juniper/apstra-go-sdk/apstra"
//	"github.com/hashicorp/terraform-plugin-framework/resource"
//	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
//	"terraform-provider-apstra/apstra/blueprint"
//)
//
//var _ resource.ResourceWithConfigure = &resourceDatacenterVirtualNetwork{}
//
//type resourceDatacenterVirtualNetwork struct {
//	client   *apstra.Client
//	lockFunc func(context.Context, string) error
//}
//
//func (o *resourceDatacenterVirtualNetwork) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
//	resp.TypeName = req.ProviderTypeName + "_datacenter_virtual_network"
//}
//
//func (o *resourceDatacenterVirtualNetwork) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
//	o.client = ResourceGetClient(ctx, req, resp)
//	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
//}
//
//func (o *resourceDatacenterVirtualNetwork) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
//	resp.Schema = schema.Schema{
//		MarkdownDescription: "This resource creates a Virtual Network within a Blueprint.",
//		Attributes:          blueprint.DatacenterVirtualNetwork{}.ResourceAttributes(),
//	}
//}
//
//func (o *resourceDatacenterVirtualNetwork) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
//	if o.client == nil {
//		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
//		return
//	}
//
//	// Retrieve values from plan.
//	var plan blueprint.DatacenterVirtualNetwork
//	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	// Lock the blueprint mutex.
//	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
//	if err != nil {
//		resp.Diagnostics.AddError(
//			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
//			err.Error())
//		return
//	}
//
//	bp, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
//	if err != nil {
//		resp.Diagnostics.AddError("error creating blueprint client", err.Error())
//		return
//	}
//
//}
