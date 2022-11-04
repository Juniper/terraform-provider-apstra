package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceLogicalDevice{}

type resourceLogicalDevice struct {
	client *goapstra.Client
}

func (o *resourceLogicalDevice) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logical_device"
}

func (o *resourceLogicalDevice) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (o *resourceLogicalDevice) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource creates an IPv4 resource pool",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Apstra ID number of the resource pool",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				MarkdownDescription: "Pool name displayed in the Apstra web UI",
				Type:                types.StringType,
				Required:            true,
			},
			"panels": {
				MarkdownDescription: "Details physical layout of interfaces on the device.",
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"rows": {
						MarkdownDescription: "Physical vertical dimension of the panel.",
						Required:            true,
						Type:                types.Int64Type,
						Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
					},
					"columns": {
						MarkdownDescription: "Physical horizontal dimension of the panel.",
						Required:            true,
						Type:                types.Int64Type,
						Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
					},
					"port_groups": {
						MarkdownDescription: "Ordered logical groupings of interfaces by speed or purpose within a panel",
						Required:            true,
						Validators:          []tfsdk.AttributeValidator{listvalidator.SizeAtLeast(1)},
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"port_count": {
								MarkdownDescription: "Number of ports in the group.",
								Required:            true,
								Type:                types.Int64Type,
								Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
							},
							"port_speed": {
								MarkdownDescription: "Port speed.",
								Required:            true,
								Type:                types.StringType,
								Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(2)}, // todo: regex validator?
							},
							"port_roles": {
								MarkdownDescription: "One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.",
								Required:            true,
								Type:                types.SetType{ElemType: types.StringType},
								Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)}, // todo: validate the strings as well?
							},
						}),
					},
				}),
			},
		},
	}, nil
}

func (o *resourceLogicalDevice) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan logicalDevice
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logicalDeviceRequest := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateLogicalDevice(ctx, logicalDeviceRequest)
	if err != nil {
		resp.Diagnostics.AddError("error creating logical device", err.Error())
	}

	plan.Id = types.StringValue(string(id))
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceLogicalDevice) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state logicalDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ld, err := o.client.GetLogicalDevice(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"error reading Logical Device",
				fmt.Sprintf("Could not Read '%s' - %s", state.Id.ValueString(), err),
			)
			return
		}
	}

	var apiState logicalDevice
	apiState.parseApi(ctx, ld, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &apiState)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceLogicalDevice) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get current state
	var state logicalDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan logicalDevice
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.UpdateLogicalDevice(ctx, goapstra.ObjectId(id), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Logical Device", err.Error())
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceLogicalDevice) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state logicalDevice
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Agent Profile by calling API
	err := o.client.DeleteLogicalDevice(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError(
				"error deleting Logical Device",
				fmt.Sprintf("could not delete Logical Device '%s' - %s", state.Id.Value, err),
			)
			return
		}
	}
}
