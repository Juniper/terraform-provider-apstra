package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceAgentProfile{}

type resourceAgentProfile struct {
	client *goapstra.Client
}

func (o *resourceAgentProfile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_profile"
}

func (o *resourceAgentProfile) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (o *resourceAgentProfile) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an Agent Profile. Note that credentials (username/password) " +
			"be set using this resource because (a) Apstra doesn't allow them to be retrieved, so it's impossible " +
			"for terraform to detect drift and because (b) leaving credentials in the configuration/state isn't a" +
			"safe practice.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Apstra ID of the Agent Profile.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Apstra name of the Agent Profile.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"has_username": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether a username has been set.",
				Computed:            true,
			},
			"has_password": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether a password has been set.",
				Computed:            true,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "Device platform.",
				Optional:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					goapstra.AgentPlatformNXOS.String(),
					goapstra.AgentPlatformJunos.String(),
					goapstra.AgentPlatformEOS.String(),
				)},
			},
			"packages": schema.MapAttribute{
				MarkdownDescription: "List of [packages](https://www.juniper.net/documentation/us/en/software/apstra4.1/apstra-user-guide/topics/topic-map/packages.html) " +
					"to be included with agents deployed using this profile.",
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.Map{mapvalidator.SizeAtLeast(1)},
			},
			"open_options": schema.MapAttribute{
				MarkdownDescription: "Passes configured parameters to offbox agents. For example, to use HTTPS as the " +
					"API connection from offbox agents to devices, use the key-value pair: proto-https - port-443.",
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.Map{mapvalidator.SizeAtLeast(1)},
			},
		},
	}
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
	id, err := o.client.CreateAgentProfile(ctx, plan.AgentProfileConfig(ctx, &resp.Diagnostics))
	if err != nil {
		resp.Diagnostics.AddError(
			"error creating new Agent Profile",
			"Could not create, unexpected error: "+err.Error(),
		)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	diags = resp.State.Set(ctx, &dAgentProfile{
		Id:          types.StringValue(string(id)),
		Name:        plan.Name,
		Platform:    plan.Platform,
		HasUsername: types.BoolValue(false),
		HasPassword: types.BoolValue(false),
		Packages:    plan.Packages,
		OpenOptions: plan.OpenOptions,
	})
	resp.Diagnostics.Append(diags...)
}

func (o *resourceAgentProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state dAgentProfile
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Agent Profile from API and then update what is in state from what the API returns
	agentProfile, err := o.client.GetAgentProfile(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading Agent Profile",
				fmt.Sprintf("Could not Read '%s' - %s", state.Id.ValueString(), err),
			)
			return
		}
	}

	// Create new state object
	newState := parseAgentProfile(ctx, agentProfile, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	diags = resp.State.Set(ctx, newState)
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
	err := o.client.UpdateAgentProfile(ctx, goapstra.ObjectId(state.Id.ValueString()), plan.AgentProfileConfig(ctx, &resp.Diagnostics))
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Agent Profile",
			fmt.Sprintf("Could not Update '%s' - %s", state.Id.ValueString(), err),
		)
		return
	}

	agentProfile, err := o.client.GetAgentProfile(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"error updating Agent Profile",
			fmt.Sprintf("Could not Update '%s' - %s", state.Id.ValueString(), err),
		)
		return
	}

	// Create new state object
	newState := parseAgentProfile(ctx, agentProfile, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	diags = resp.State.Set(ctx, newState)
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
	err := o.client.DeleteAgentProfile(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError(
				"error deleting Agent Profile",
				fmt.Sprintf("could not delete Agent Profile '%s' - %s", state.Id.ValueString(), err),
			)
			return
		}
	}
}
