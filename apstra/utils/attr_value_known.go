package utils

import "github.com/hashicorp/terraform-plugin-framework/attr"

func Known(v attr.Value) bool {
	return !v.IsUnknown() && !v.IsNull()
}
