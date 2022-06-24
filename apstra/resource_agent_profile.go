package apstra

import (
	"context"
	"errors"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceAgentProfileType struct{}

func (r resourceAgentProfileType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			// todo: validate non-empty
			"name": {
				Type:       types.StringType,
				Required:   true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"packages": {
				Optional:   true,
				Type:       types.SetType{ElemType: types.StringType},
				Validators: []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			// todo: validate non-empty
			"platform": {
				Type:       types.StringType,
				Optional:   true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
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

	// Create new Agent Profile
	id, err := r.p.client.CreateSystemAgentProfile(ctx, &goapstra.SystemAgentProfileConfig{
		Label:    plan.Name.Value,
		Platform: plan.Platform.Value,
		Packages: agentProfilePackagesFromPlan(plan.Packages),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent Profile",
			"Could not create, unexpected error: "+err.Error(),
		)
		return
	}

	// Generate resource state struct
	var result = ResourceAgentProfile{
		Name:     types.String{Value: plan.Name.Value},
		Id:       types.String{Value: string(id)},
		Packages: plan.Packages,
		Platform: plan.Platform,
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceAgentProfile) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
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
				fmt.Sprintf("Could not Read '%s' - %s", state.Id.Value, err),
			)
			return
		}
	}

	// Map response body to resource schema attribute
	state.Id = types.String{Value: string(agentProfile.Id)}
	state.Name = types.String{Value: agentProfile.Label}
	state.Packages = agentProfilePackagesFromApi(agentProfile.Packages)

	if agentProfile.Platform == "" {
		state.Platform = types.String{Null: true}
	} else {
		state.Platform = types.String{Value: agentProfile.Platform}
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceAgentProfile) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get current state
	var state ResourceAgentProfile
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan ResourceAgentProfile
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update new Agent Profile
	err := r.p.client.UpdateSystemAgentProfile(ctx, goapstra.ObjectId(state.Id.Value), &goapstra.SystemAgentProfileConfig{
		Label:    plan.Name.Value,
		Platform: plan.Platform.Value,
		Packages: agentProfilePackagesFromPlan(plan.Packages),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Agent Profile",
			fmt.Sprintf("Could not Update '%s' - %s", state.Id.Value, err),
		)
		return
	}

	// Update state
	state.Name = plan.Name
	state.Platform = plan.Platform
	state.Packages = plan.Packages

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceAgentProfile) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ResourceAgentProfile
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete System Agent Profile by calling API
	err := r.p.client.DeleteSystemAgentProfile(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"error deleting Agent Profile",
			fmt.Sprintf("could not delete Agent Profile '%s' - %s", state.Id.Value, err),
		)
		return
	}
}

func agentProfilePackagesFromPlan(in []types.String) []string {
	var out []string
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
