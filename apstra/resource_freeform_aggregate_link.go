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
	_ resource.ResourceWithConfigure = &resourceFreeformAggregateLink{}
	_ resourceWithSetFfBpClientFunc  = &resourceFreeformAggregateLink{}
	_ resourceWithSetBpLockFunc      = &resourceFreeformAggregateLink{}
)

type resourceFreeformAggregateLink struct {
	getBpClientFunc func(context.Context, string) (*apstra.FreeformClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceFreeformAggregateLink) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_freeform_aggregate_link"
}

func (o *resourceFreeformAggregateLink) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceFreeformAggregateLink) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFreeform + "This resource creates an Aggregate Link in a Freeform Blueprint.",
		Attributes:          (*freeform.AggregateLink)(nil).ResourceAttributes(),
	}
}

func (o *resourceFreeformAggregateLink) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan freeform.AggregateLink
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the Freeform reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintID), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintID.ValueString()),
			err.Error())
		return
	}

	// Convert the plan into an API Request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the link
	id, err := bp.CreateAggregateLink(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Aggregate Link", err.Error())
		return
	}
	plan.ID = types.StringValue(id)

	// Read back the new link
	link, err := bp.GetAggregateLink(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error reading new Aggregate Link", err.Error())
	}
	plan.LoadApiData(ctx, link, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceFreeformAggregateLink) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (o *resourceFreeformAggregateLink) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (o *resourceFreeformAggregateLink) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (o *resourceFreeformAggregateLink) setBpClientFunc(f func(context.Context, string) (*apstra.FreeformClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceFreeformAggregateLink) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
