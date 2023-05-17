package utils

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"reflect"
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

func Int64ValueOrNull(_ context.Context, in any, diags *diag.Diagnostics) types.Int64 {
	// when in is nil, return a null attr.Value
	if in == nil {
		return types.Int64Null()
	}

	// when in is a nil pointer, return a null attr.Value
	if reflect.TypeOf(in).Kind() == reflect.Ptr && reflect.ValueOf(in).IsNil() {
		return types.Int64Null()
	}

	switch in.(type) {
	case *apstra.Vlan:
		return types.Int64Value(int64(*in.(*apstra.Vlan)))
	case *int:
		return types.Int64Value(int64(*in.(*int)))
	case *int8:
		return types.Int64Value(int64(*in.(*int8)))
	case *int16:
		return types.Int64Value(int64(*in.(*int16)))
	case *int32:
		return types.Int64Value(int64(*in.(*int32)))
	case *int64:
		return types.Int64Value(*in.(*int64))
	case *uint:
		return types.Int64Value(int64(*in.(*uint)))
	case *uint8:
		return types.Int64Value(int64(*in.(*uint8)))
	case *uint16:
		return types.Int64Value(int64(*in.(*uint16)))
	case *uint32:
		return types.Int64Value(int64(*in.(*uint32)))
	case *uint64:
		return types.Int64Value(int64(*in.(*uint64)))
	case apstra.Vlan:
		return types.Int64Value(int64(in.(apstra.Vlan)))
	case int:
		return types.Int64Value(int64(in.(int)))
	case int8:
		return types.Int64Value(int64(in.(int8)))
	case int16:
		return types.Int64Value(int64(in.(int16)))
	case int32:
		return types.Int64Value(int64(in.(int32)))
	case int64:
		return types.Int64Value(in.(int64))
	case uint:
		return types.Int64Value(int64(in.(uint)))
	case uint8:
		return types.Int64Value(int64(in.(uint8)))
	case uint16:
		return types.Int64Value(int64(in.(uint16)))
	case uint32:
		return types.Int64Value(int64(in.(uint32)))
	case uint64:
		return types.Int64Value(int64(in.(uint64)))

	default:
		diags.AddError("cannot convert interface to int64",
			fmt.Sprintf("value is type %s", reflect.TypeOf(in).String()))
	}

	return types.Int64Null()
}
