package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
				Type:     types.StringType,
				Required: true,
				//Validators: []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"platform": {
				Type:     types.StringType,
				Optional: true,
			},
			"packages": {
				Optional: true,
				Type:     types.MapType{ElemType: types.StringType},
			},
			"open_options": {
				Type:     types.MapType{ElemType: types.StringType},
				Optional: true,
			},
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
	id, err := r.p.client.CreateAgentProfile(ctx, &goapstra.AgentProfileConfig{
		Label:       plan.Name.Value,
		Platform:    plan.Platform.Value,
		Packages:    typeMapStringToMapStringString(plan.Packages),
		OpenOptions: typeMapStringToMapStringString(plan.OpenOptions),
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
		Name:        types.String{Value: plan.Name.Value},
		Id:          types.String{Value: string(id)},
		Packages:    plan.Packages,
		Platform:    plan.Platform,
		OpenOptions: plan.OpenOptions,
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
	agentProfile, err := r.p.client.GetAgentProfile(ctx, goapstra.ObjectId(state.Id.Value))
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
	state.Packages = mapStringStringToTypeMapString(agentProfile.Packages)
	state.OpenOptions = mapStringStringToTypeMapString(agentProfile.OpenOptions)

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
	err := r.p.client.UpdateAgentProfile(ctx, goapstra.ObjectId(state.Id.Value), &goapstra.AgentProfileConfig{
		Label:       plan.Name.Value,
		Platform:    plan.Platform.Value,
		Packages:    typeMapStringToMapStringString(plan.Packages),
		OpenOptions: typeMapStringToMapStringString(plan.OpenOptions),
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
	state.OpenOptions = plan.OpenOptions

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

	// Delete Agent Profile by calling API
	err := r.p.client.DeleteAgentProfile(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"error deleting Agent Profile",
			fmt.Sprintf("could not delete Agent Profile '%s' - %s", state.Id.Value, err),
		)
		return
	}
}

func typeMapStringToMapStringString(in types.Map) map[string]string {
	var out map[string]string
	if len(in.Elems) > 0 {
		out = make(map[string]string)
	}
	for k, v := range in.Elems {
		out[k] = v.(types.String).Value
	}
	return out
}

func mapStringStringToTypeMapString(in map[string]string) types.Map {
	out := types.Map{
		Null:     len(in) == 0,
		ElemType: types.StringType,
		Elems:    make(map[string]attr.Value),
	}
	for k, v := range in {
		out.Elems[k] = types.String{Value: v}
	}
	return out

}
