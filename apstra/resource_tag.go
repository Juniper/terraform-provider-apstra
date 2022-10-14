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

var _ resource.ResourceWithConfigure = &resourceTag{}

type resourceTag struct {
	client *goapstra.Client
}

func (o *resourceTag) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (o *resourceTag) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (o *resourceTag) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource creates a Tag in the Apstra Design tab.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Apstra ID of the Tag.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				MarkdownDescription: "Name of the Tag as seen in the web UI.",
				Type:                types.StringType,
				Required:            true,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"description": {
				MarkdownDescription: "Indicates whether a username has been set.",
				Type:                types.StringType,
				Optional:            true,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
		},
	}, nil
}

func (o *resourceTag) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan dTag
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var description string
	if plan.Description.IsNull() {
		description = ""
	} else {
		description = plan.Description.Value
	}

	// Create new Tag
	id, err := o.client.CreateTag(ctx, &goapstra.DesignTagRequest{
		Label:       plan.Name.Value,
		Description: description,
	})
	if err != nil {
		resp.Diagnostics.AddError("error creating new Tag", err.Error())
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &dTag{
		Id:          types.String{Value: string(id)},
		Name:        plan.Name,
		Description: plan.Description,
	})
	resp.Diagnostics.Append(diags...)
}

func (o *resourceTag) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state dTag
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := o.client.GetTag(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading Tag", err.Error())
			return
		}
	}

	var description types.String
	if tag.Data.Description == "" {
		description = types.String{Null: true}
	} else {
		description = types.String{Value: tag.Data.Description}
	}

	// Set state
	diags = resp.State.Set(ctx, &dTag{
		Id:          types.String{Value: string(tag.Id)},
		Name:        types.String{Value: tag.Data.Label},
		Description: description,
	})
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (o *resourceTag) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get current state
	var state dTag
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan dTag
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var description string
	if plan.Description.IsNull() {
		description = ""
	} else {
		description = plan.Description.Value
	}

	// Update new Agent Profile
	err := o.client.UpdateTag(ctx, goapstra.ObjectId(state.Id.Value), &goapstra.DesignTagRequest{
		Label:       plan.Name.Value,
		Description: description,
	})
	if err != nil {
		resp.Diagnostics.AddError("error updating Tag", err.Error())
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &dTag{
		Id:          plan.Id,
		Name:        plan.Name,
		Description: plan.Description,
	})
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (o *resourceTag) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state dTag
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Tag by calling API
	err := o.client.DeleteTag(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError("error deleting Tag", err.Error())
			return
		}
	}
}
