package apstraplanmodifier

// the contents of this file are based on:
// https://github.com/hashicorp/terraform-plugin-framework/blob/v1.9.0/resource/schema/stringplanmodifier/use_state_for_unknown.go
//
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseNullStateForUnknown returns a plan modifier that copies a known prior state
// value into the planned value. Use this when it is known that an unconfigured
// value will remain the same after a resource update.
//
// To prevent Terraform errors, the framework automatically sets unconfigured
// and Computed attributes to an unknown value "(known after apply)" on update.
// Using this plan modifier will instead display the prior state value in the
// plan, unless a prior plan modifier adjusts the value.
func UseNullStateForUnknown() useNullStateForUnknownModifier {
	return useNullStateForUnknownModifier{}
}

var (
	_ planmodifier.Bool    = (*useNullStateForUnknownModifier)(nil)
	_ planmodifier.Float64 = (*useNullStateForUnknownModifier)(nil)
	_ planmodifier.Int64   = (*useNullStateForUnknownModifier)(nil)
	_ planmodifier.List    = (*useNullStateForUnknownModifier)(nil)
	_ planmodifier.Map     = (*useNullStateForUnknownModifier)(nil)
	_ planmodifier.Number  = (*useNullStateForUnknownModifier)(nil)
	_ planmodifier.Object  = (*useNullStateForUnknownModifier)(nil)
	_ planmodifier.Set     = (*useNullStateForUnknownModifier)(nil)
	_ planmodifier.String  = (*useNullStateForUnknownModifier)(nil)
)

// useNullStateForUnknownModifier implements the plan modifier.
type useNullStateForUnknownModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m useNullStateForUnknownModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change, even if it's null."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useNullStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change, even if it's null."
}

// PlanModifyBool implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifyInt64 implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifyFloat64 implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifyFloat64(_ context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifyList implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifyMap implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifyMap(_ context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifyNumber implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifyNumber(_ context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifyObject implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifyObject(_ context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifySet implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

// PlanModifyString implements the plan modification logic.
func (m useNullStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}
