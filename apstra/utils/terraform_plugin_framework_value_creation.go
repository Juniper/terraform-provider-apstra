package utils

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringValueOrNull returns a types.String based on the supplied string. If the
// supplied string is empty, the returned types.String will be flagged as null.
func StringValueOrNull(_ context.Context, in string, _ *diag.Diagnostics) types.String {
	if in == "" {
		return types.StringNull()
	}

	return types.StringValue(in)
}

// StringValueWithNull returns a types.String based on the supplied inStr. If
// inStr matches nullStr or is empty, the returned types.String will be flagged
// as null.
func StringValueWithNull(ctx context.Context, inStr string, nullStr string, diags *diag.Diagnostics) types.String {
	if inStr == nullStr {
		return types.StringNull()
	}
	return StringValueOrNull(ctx, inStr, diags)
}

// MapValueOrNull returns a types.Map based on the supplied elements. If the
// supplied elements is empty, the returned types.Map will be flagged as null.
func MapValueOrNull[T any](ctx context.Context, elementType attr.Type, elements map[string]T, diags *diag.Diagnostics) types.Map {
	if len(elements) == 0 {
		return types.MapNull(elementType)
	}

	result, d := types.MapValueFrom(ctx, elementType, elements)
	diags.Append(d...)
	return result
}

// ListValueOrNull returns a types.List based on the supplied elements. If the
// supplied elements is empty, the returned types.List will be flagged as null.
func ListValueOrNull[T any](ctx context.Context, elementType attr.Type, elements []T, diags *diag.Diagnostics) types.List {
	if len(elements) == 0 {
		return types.ListNull(elementType)
	}

	result, d := types.ListValueFrom(ctx, elementType, elements)
	diags.Append(d...)
	return result
}

// SetValueOrNull returns a types.Set based on the supplied elements. If the
// supplied elements is empty, the returned types.Set will be flagged as null.
func SetValueOrNull[T any](ctx context.Context, elementType attr.Type, elements []T, diags *diag.Diagnostics) types.Set {
	if len(elements) == 0 {
		return types.SetNull(elementType)
	}

	result, d := types.SetValueFrom(ctx, elementType, elements)
	diags.Append(d...)
	return result
}

// ObjectValueOrNull returns a types.Object based on the supplied attributes. If the
// supplied attributes is nil, the returned types.Object will be flagged as null.
func ObjectValueOrNull(ctx context.Context, attrTypes map[string]attr.Type, attributes any, diags *diag.Diagnostics) types.Object {
	if attributes == nil {
		return types.ObjectNull(attrTypes)
	}

	result, d := types.ObjectValueFrom(ctx, attrTypes, attributes)
	diags.Append(d...)
	return result
}
