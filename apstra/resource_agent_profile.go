package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceAgentProfile{}

type resourceAgentProfile struct {
	client *goapstra.Client
}

func (o *resourceAgentProfile) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "apstra_agent_profile"
}

func (o *resourceAgentProfile) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errResourceConfigureProviderDataDetail,
			fmt.Sprintf(errResourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *resourceAgentProfile) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource creates an Agent Profile. Note that credentials (username/password) " +
			"be set using this resource because (a) Apstra doesn't allow them to be retrieved, so it's impossible " +
			"for terraform to detect drift and because (b) leaving credentials in the configuration/state isn't a" +
			"safe practice.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Apstra ID of the Agent Profile.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				MarkdownDescription: "Apstra name of the Agent Profile.",
				Type:                types.StringType,
				Required:            true,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"has_username": {
				MarkdownDescription: "Indicates whether a username has been set.",
				Type:                types.BoolType,
				Computed:            true,
			},
			"has_password": {
				MarkdownDescription: "Indicates whether a password has been set.",
				Type:                types.BoolType,
				Computed:            true,
			},
			"platform": {
				MarkdownDescription: "Device platform.",
				Type:                types.StringType,
				Optional:            true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(
					goapstra.AgentPlatformNXOS.String(),
					goapstra.AgentPlatformJunos.String(),
					goapstra.AgentPlatformEOS.String(),
				)},
			},
			"packages": {
				MarkdownDescription: "List of [packages](https://www.juniper.net/documentation/us/en/software/apstra4.1/apstra-user-guide/topics/topic-map/packages.html) " +
					"to be included with agents deployed using this profile.",
				Optional: true,
				Type:     types.MapType{ElemType: types.StringType},
			},
			"open_options": {
				MarkdownDescription: "Passes configured parameters to offbox agents. For example, to use HTTPS as the " +
					"API connection from offbox agents to devices, use the key-value pair: proto-https - port-443.",
				Type:     types.MapType{ElemType: types.StringType},
				Optional: true,
			},
		},
	}, nil
}

func (o *resourceAgentProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan dAgentProfile
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Agent Profile
	id, err := o.client.CreateAgentProfile(ctx, plan.AgentProfileConfig())
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent Profile",
			"Could not create, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &dAgentProfile{
		Id:          types.String{Value: string(id)},
		Name:        plan.Name,
		Platform:    plan.Platform,
		HasUsername: types.Bool{Value: false},
		HasPassword: types.Bool{Value: false},
		Packages:    plan.Packages,
		OpenOptions: plan.OpenOptions,
	})
	resp.Diagnostics.Append(diags...)
}

func (o *resourceAgentProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
	}

	// Get current state
	var state dAgentProfile
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Agent Profile from API and then update what is in state from what the API returns
	agentProfile, err := o.client.GetAgentProfile(ctx, goapstra.ObjectId(state.Id.Value))
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

	// Set state
	diags = resp.State.Set(ctx, parseAgentProfile(agentProfile))
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (o *resourceAgentProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get current state
	var state dAgentProfile
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan dAgentProfile
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update new Agent Profile
	err := o.client.UpdateAgentProfile(ctx, goapstra.ObjectId(state.Id.Value), plan.AgentProfileConfig())
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Agent Profile",
			fmt.Sprintf("Could not Update '%s' - %s", state.Id.Value, err),
		)
		return
	}

	agentProfile, err := o.client.GetAgentProfile(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Agent Profile",
			fmt.Sprintf("Could not Update '%s' - %s", state.Id.Value, err),
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, parseAgentProfile(agentProfile))
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (o *resourceAgentProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state dAgentProfile
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Agent Profile by calling API
	err := o.client.DeleteAgentProfile(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError(
				"error deleting Agent Profile",
				fmt.Sprintf("could not delete Agent Profile '%s' - %s", state.Id.Value, err),
			)
			return
		}
	}
}

func parseAgentProfile(in *goapstra.AgentProfile) *dAgentProfile {
	return &dAgentProfile{
		Id:          types.String{Value: string(in.Id)},
		Name:        types.String{Value: in.Label},
		Platform:    platformToTFString(in.Platform),
		HasUsername: types.Bool{Value: in.HasUsername},
		HasPassword: types.Bool{Value: in.HasPassword},
		Packages:    mapStringStringToTypesMap(in.Packages),
		OpenOptions: mapStringStringToTypesMap(in.OpenOptions),
	}
}
