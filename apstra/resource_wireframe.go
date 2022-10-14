package apstra

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceWireframeType struct{}

func (r resourceWireframe) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "apstra_wireframe"
}

func (r resourceWireframe) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"tags": {
				Optional: true,
				Type:     types.SetType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (r resourceWireframeType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceWireframe{
		p: *(p.(*Provider)),
	}, nil
}

type resourceWireframe struct {
	p Provider
}

func (r resourceWireframe) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	plan := &ResourceWireframe{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//// Create new wireframe object
	//id, err := r.p.client.CreateWireframe(ctx, &goapstra.Wireframe{
	//	DisplayName: plan.Name.Value,
	//	Tags:        tags,
	//})
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		"error creating new wireframe",
	//		"Could not create wireframe, unexpected error: "+err.Error(),
	//	)
	//	return
	//}

	// Generate resource state struct
	var result = ResourceWireframe{
		//Id:   types.String{Value: string(id)},
		Name: plan.Name,
		Tags: plan.Tags,
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceWireframe) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ResourceWireframe
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//// Get object from API and then update what is in state from what the API returns
	//wireframe, err := r.p.client.GetWireframe(ctx, goapstra.ObjectId(state.Id.Value))
	//if err != nil {
	//	var ace goapstra.ApstraClientErr
	//	if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
	//		// resource deleted outside of terraform
	//		resp.State.RemoveResource(ctx)
	//		return
	//	} else {
	//		resp.Diagnostics.AddError(
	//			"error reading wireframe",
	//			fmt.Sprintf("could not read wireframe '%s' - %s", state.Id.Value, err),
	//		)
	//		return
	//	}
	//}
	//
	//// Map response body to resource schema attribute
	//state.Id = types.String{Value: string(wireframe.Id)}
	//state.Name = types.String{Value: wireframe.DisplayName}
	//state.Tags = wireframeTagsFromApi(wireframe.Tags)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceWireframe) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan ResourceWireframe
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state ResourceWireframe
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//// Fetch existing
	//wireframe, err := r.p.client.GetWireframe(ctx, goapstra.ObjectId(state.Id.Value))
	//if err != nil {
	//	var ace goapstra.ApstraClientErr
	//	if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
	//		resp.Diagnostics.AddError(
	//			fmt.Sprintf("cannot update %s", resourceWireframeName),
	//			fmt.Sprintf("error fetching existing wireframe '%s' not found", state.Id.Value),
	//		)
	//		return
	//	}
	//	// some other unknown error
	//	resp.Diagnostics.AddError(
	//		fmt.Sprintf("cannot update %s", resourceWireframeName),
	//		fmt.Sprintf("error fetching existing wireframe '%s' - %s", state.Id.Value, err.Error()),
	//	)
	//	return
	//}

	//// Create/Update object @ API
	//err = r.p.client.UpdateWireframe(ctx, goapstra.ObjectId(state.Id.Value), &goapstra.Wireframe{
	//	DisplayName: plan.Name.Value,
	//	Tags:        wireframeTagsFromPlan(plan.Tags),
	//})
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		fmt.Sprintf("cannot update %s", resourceWireframeName),
	//		fmt.Sprintf("cannot update %s '%s' - %s", resourceWireframeName, plan.Id.Value, err.Error()),
	//	)
	//	return
	//}
	state.Name = plan.Name
	state.Tags = plan.Tags

	// Set new state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceWireframe) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceWireframe
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//// Delete wireframe by calling API
	//err := r.p.client.DeleteWireframe(ctx, goapstra.ObjectId(state.Id.Value))
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		"error deleting wireframe",
	//		fmt.Sprintf("could not delete wireframe '%s' - %s", id, err),
	//	)
	//	return
	//}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}
