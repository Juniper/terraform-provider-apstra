package utils

import "github.com/hashicorp/terraform-plugin-framework/attr"

func HasValue(v attr.Value) bool {
	return !v.IsUnknown() && !v.IsNull()
}
