package apstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type resourceAgentProfileType struct{}

func (r resourceAgentProfileType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				//PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			//"packages": {
			//	Optional: true,
			//	Type:     types.SetType{ElemType: types.StringType},
			//},
			//"has_username": {
			//	Type:     types.BoolType,
			//	Computed: true,
			//	PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			//},
			//"has_password": {
			//	Type:     types.BoolType,
			//	Computed: true,
			//	PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			//},
			"username": {
				Type:     types.StringType,
				Optional: true,
				//PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			},
			"password": {
				Type:     types.StringType,
				Optional: true,
				//PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			},
			"platform": {
				Type:     types.StringType,
				Optional: true,
			},
			//"open_options": {
			//	Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{}),
			//	Optional:   true,
			//},
		},
	}, nil
}

func (r resourceAgentProfileType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceAgentProfile{
		p: *(p.(*provider)),
	}, nil
}

type resourceAgentProfile struct {
	p provider
}

func (r resourceAgentProfile) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	fmt.Fprintf(stderr, "[DEBUG]-xxxxxx this is Create\n")
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceAgentProfile
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// prep packages for new Agent Profile
	//packages := agentProfilePackagesFromPlan(plan.Packages)

	// Create new Agent Profile
	agentProfileConfig := goapstra.SystemAgentProfileConfig{
		Label:    plan.Name.Value,
		Username: plan.Username.Value,
		Password: plan.Password.Value,
		Platform: plan.Platform.Value,
	}
	id, err := r.p.client.CreateSystemAgentProfile(ctx, &agentProfileConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent Profile",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "created Agent Profile", map[string]interface{}{"id": string(id)})

	// Generate resource state struct
	var result = ResourceAgentProfile{
		Name: types.String{Value: plan.Name.Value},
		Id:   types.String{Value: string(id)},
		//Packages: plan.Packages,
		//HasUsername: types.Bool{Value: !plan.Username.IsNull()},
		//HasPassword: types.Bool{Value: !plan.Password.IsNull()},
		Username: types.String{Value: plan.Username.Value},
		Password: types.String{Value: plan.Password.Value},
		Platform: plan.Platform,
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceAgentProfile) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	fmt.Fprintf(stderr, "[DEBUG]-xxxxxx this is Read\n")
	// Get current state
	var state ResourceAgentProfile
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Agent Profile from API and then update what is in state from what the API returns
	agentProfile, err := r.p.client.GetSystemAgentProfile(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading Agent Profile",
				fmt.Sprintf("could not read Agent Profile '%s' - %s", state.Id.Value, err),
			)
			return
		}
	}

	// Map response body to resource schema attribute
	// todo: error check state.Id vs. agentProfile.Id
	state.Id = types.String{Value: string(agentProfile.Id)}
	state.Name = types.String{Value: agentProfile.Label}
	state.Username = types.String{Value: "bogus username"} // default?
	state.Password = types.String{Value: "bogus password"} // default?
	if agentProfile.Platform == "" {
		state.Platform = types.String{Null: true}
	} else {
		state.Platform = types.String{Value: agentProfile.Platform}
	}
	//state.Packages = agentProfilePackagesFromApi(agentProfile.Packages)
	//state.HasUsername = types.Bool{Value: agentProfile.HasUsername}
	//state.HasPassword = types.Bool{Value: agentProfile.HasPassword}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceAgentProfile) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	//// Get plan values
	//var plan ResourceAgentProfile
	//diags := req.Plan.Get(ctx, &plan)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// Get current state
	//var state ResourceAgentProfile
	//diags = req.State.Get(ctx, &state)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// Fetch existing []goapstra.AsnRange
	////goland:noinspection GoPreferNilSlice
	//agentProfileFromApi, err := r.p.client.GetSystemAgentProfile(ctx, goapstra.ObjectId(state.Id.Value))
	//if err != nil {
	//	var ace goapstra.ApstraClientErr
	//	if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
	//		resp.Diagnostics.AddError(
	//			fmt.Sprintf("cannot update %s", resourceAgentProfileName),
	//			fmt.Sprintf("error fetching existing ASN ranges - ASN pool '%s' not found", state.Id.Value),
	//		)
	//		return
	//	}
	//	// some other unknown error
	//	resp.Diagnostics.AddError(
	//		fmt.Sprintf("cannot update %s", resourceAgentProfileName),
	//		fmt.Sprintf("error fetching existing ASN ranges for ASN pool '%s' - %s", state.Id.Value, err.Error()),
	//	)
	//	return
	//}
	//
	//// Generate API request body from plan
	//// todo: flesh out
	//send := &goapstra.SystemAgentProfileConfig{
	//	Label:    plan.Label.Value,
	//	Packages: agentProfilePackagesFromPlan(plan.Packages),
	//}
	//
	//// Create/Update the Agent Profile
	//err = r.p.client.UpdateSystemAgentProfile(ctx, goapstra.ObjectId(state.Id.Value), send)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		fmt.Sprintf("cannot update %s", resourceAgentProfileName),
	//		fmt.Sprintf("cannot update %s '%s' - %s", resourceAgentProfileName, plan.Id.Value, err.Error()),
	//	)
	//	return
	//}
	//// todo: pretty bold saving plan data directly to state w/out checking whether the API call worked...
	//state.DisplayName = plan.DisplayName
	//state.Tags = plan.Tags
	//
	//// Set new state
	//diags = resp.State.Set(ctx, state)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
}

// Delete resource
func (r resourceAgentProfile) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ResourceAgentProfile
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get System Agent Profile ID from state
	id := state.Id.Value

	// Delete System Agent Profile by calling API
	err := r.p.client.DeleteSystemAgentProfile(ctx, goapstra.ObjectId(id))
	if err != nil {
		resp.Diagnostics.AddError(
			"error deleting Agent Profile",
			fmt.Sprintf("could not delete Agent Profile '%s' - %s", id, err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func agentProfilePackagesFromPlan(in []types.String) []string {
	//goland:noinspection GoPreferNilSlice
	out := []string{}
	for _, p := range in {
		out = append(out, p.Value)
	}
	return out
}

func agentProfilePackagesFromApi(in []string) []types.String {
	var out []types.String
	for _, p := range in {
		out = append(out, types.String{Value: p})
	}
	return out
}
