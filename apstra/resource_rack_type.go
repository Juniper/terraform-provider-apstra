package apstra

import (
	"context"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"strings"
)

type resourceRackTypeType struct{}

func (r resourceRackTypeType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	fcpModes := []string{
		goapstra.FabricConnectivityDesignL3Clos.String(),
		goapstra.FabricConnectivityDesignL3Collapsed.String()}
	fcdRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
		strings.Join(fcpModes, "$|^")))
	if err != nil {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("error compiling fabric connectivity design regex", err.Error())
		return tfsdk.Schema{}, diagnostics
	}

	leafRedundancyProtocols := []string{
		goapstra.LeafRedundancyProtocolEsi.String(),
		goapstra.LeafRedundancyProtocolMlag.String(),
	}
	leafRedundancyRegexp, err := regexp.Compile(fmt.Sprintf("^%s$",
		strings.Join(leafRedundancyProtocols, "$|^")))
	if err != nil {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("error compiling leaf redundancy regex", err.Error())
		return tfsdk.Schema{}, diagnostics
	}

	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:       types.StringType,
				Required:   true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"description": {
				Type:     types.StringType,
				Optional: true,
			},
			"fabric_connectivity_design": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
					fcdRegexp,
					fmt.Sprintf("fabric_connectivity_design must be one of: '%s', '%s'",
						goapstra.FabricConnectivityDesignL3Clos.String(),
						goapstra.FabricConnectivityDesignL3Collapsed.String()))},
			},
			"leaf_switches": {
				Required: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
					},
					"spine_link_count": {
						Type:          types.Int64Type,
						Computed:      true,
						Optional:      true,
						PlanModifiers: tfsdk.AttributePlanModifiers{defaultInt64Modifier{Default: 1}},
					},
					"spine_link_speed": {
						Type:     types.StringType,
						Required: true,
					},
					"l3_peer_link_count": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"l3_peer_link_speed": {
						Type:     types.StringType,
						Optional: true,
					},
					"l3_peer_link_port_channel_id": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"peer_link_count": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"peer_link_speed": {
						Type:     types.StringType,
						Optional: true,
					},
					"peer_link_port_channel_id": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"mlag_vlan_id": {
						Type:     types.Int64Type,
						Optional: true,
					},
					"redundancy_protocol": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(
							leafRedundancyRegexp,
							fmt.Sprintf("redundancy_protocol must be one of: '%s', '%s'",
								goapstra.LeafRedundancyProtocolEsi.String(),
								goapstra.LeafRedundancyProtocolMlag.String()))},
					},
					"tags": {
						Type:     types.SetType{ElemType: types.StringType},
						Optional: true,
					},
					"logical_device_id": {
						Type:     types.StringType,
						Required: true,
					},
				})},
		},
	}, nil
}

func (r resourceRackTypeType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceRackType{
		p: *(p.(*provider)),
	}, nil
}

type resourceRackType struct {
	p provider
}

func (r resourceRackType) ValidateConfig(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	var cfg ResourceRackType
	req.Config.Get(ctx, &cfg)

	if len(cfg.LeafSwitches) == 0 {
		resp.Diagnostics.AddError(
			"missing required configuration element",
			"at least one 'leaf_switches' element is required")
	}
}

func (r resourceRackType) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceRackType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceRackType) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ResourceRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceRackType) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get current state
	var state ResourceRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan values
	var plan ResourceRackType
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceRackType) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ResourceRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete rack type by calling API
	err := r.p.client.DeleteRackType(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"error deleting Rack Type",
			fmt.Sprintf("could not delete Rack Type '%s' - %s", state.Id.Value, err),
		)
		return
	}
}

// defaultInt64Modifier is a plan modifier that sets a default value on
// types.Int64Type attributes when not configured. The attribute must be marked
// as Optional and Computed. When setting the state during the resource Create,
// Read, or Update methods, this default value must also be included or the
// Terraform CLI will generate an error.
type defaultInt64Modifier struct {
	Default int64
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m defaultInt64Modifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %d", m.Default)
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m defaultInt64Modifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%d`", m.Default)
}

// Modify runs the logic of the plan modifier.
// Access to the configuration, plan, and state is available in `req`, while
// `resp` contains fields for updating the planned value, triggering resource
// replacement, and returning diagnostics.
func (m defaultInt64Modifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	var i types.Int64
	diags := tfsdk.ValueAs(ctx, req.AttributePlan, &i)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if i.Value != 0 {
		return
	}

	resp.AttributePlan = types.Int64{Value: m.Default}
}
