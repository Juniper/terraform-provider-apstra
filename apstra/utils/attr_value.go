package utils

import "github.com/hashicorp/terraform-plugin-framework/attr"

func Known(v attr.Value) bool {
	return !v.IsUnknown() && !v.IsNull()
}

func NullableValueRequiresUpdate(plan, state attr.Value) bool {
	if Known(plan) && !plan.Equal(state) {
		return true
	}

	if plan.IsNull() && !state.IsNull() {
		return true
	}

	return false
}
