package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceAsnPool{}

type resourceAsnPool struct {
	client *goapstra.Client
}

func (o *resourceAsnPool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_pool"
}

func (o *resourceAsnPool) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (o *resourceAsnPool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an ASN resource pool",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Apstra ID number of the resource pool",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Pool name displayed in the Apstra web UI",
				Required:            true,
			},
		},
	}
}

func (o *resourceAsnPool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rAsnPool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new ASN Pool
	id, err := o.client.CreateAsnPool(ctx, &goapstra.AsnPoolRequest{
		DisplayName: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("error creating new ASN Pool", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &rAsnPool{
		Id:   types.StringValue(string(id)),
		Name: plan.Name,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceAsnPool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rAsnPool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ASN pool from API and then update what is in state from what the API returns
	pool, err := o.client.GetAsnPool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading ASN pool", err.Error())
			return
		}
	}

	dPool := parseAsnPool(ctx, pool, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &rAsnPool{
		Id:   dPool.Id,
		Name: dPool.Name,
	})
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (o *resourceAsnPool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan rAsnPool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state rAsnPool
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentPool, err := o.client.GetAsnPool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning("API error",
				fmt.Sprintf("error fetching existing ASN pool - pool '%s' not found", state.Id.ValueString()),
			)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError("API error",
			fmt.Sprintf("error fetching ASN pool '%s' - %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	// Generate API request body from plan (only the DisplayName can be changed here)
	send := &goapstra.AsnPoolRequest{
		DisplayName: plan.Name.ValueString(),
		Ranges:      make([]goapstra.IntfIntRange, len(currentPool.Ranges)),
	}

	// ranges are independent resources, so whatever was found via GET must be re-applied here.
	for i, r := range currentPool.Ranges {
		send.Ranges[i] = r
	}

	// Create/Update ASN pool
	err = o.client.UpdateAsnPool(ctx, goapstra.ObjectId(state.Id.ValueString()), send)
	if err != nil {
		resp.Diagnostics.AddError("error updating ASN pool", err.Error())
		return
	}

	// Set new state
	diags = resp.State.Set(ctx, &rAsnPool{
		Id:   state.Id,
		Name: plan.Name,
	})
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (o *resourceAsnPool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rAsnPool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete ASN pool by calling API
	err := o.client.DeleteAsnPool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound {
			resp.Diagnostics.AddError(
				"error deleting ASN pool", err.Error())
		}
		return
	}
}

type rAsnPool struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
