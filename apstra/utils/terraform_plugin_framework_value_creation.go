package utils

import (
	"context"
	"fmt"
	"net"
	"reflect"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/constraints"
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

	switch in := in.(type) {
	case *apstra.VNI:
		return types.Int64Value(int64(*in))
	case *apstra.Vlan:
		return types.Int64Value(int64(*in))
	case *int:
		return types.Int64Value(int64(*in))
	case *int8:
		return types.Int64Value(int64(*in))
	case *int16:
		return types.Int64Value(int64(*in))
	case *int32:
		return types.Int64Value(int64(*in))
	case *int64:
		return types.Int64Value(*in)
	case *uint:
		return types.Int64Value(int64(*in))
	case *uint8:
		return types.Int64Value(int64(*in))
	case *uint16:
		return types.Int64Value(int64(*in))
	case *uint32:
		return types.Int64Value(int64(*in))
	case *uint64:
		return types.Int64Value(int64(*in))
	case apstra.VNI:
		return types.Int64Value(int64(in))
	case apstra.Vlan:
		return types.Int64Value(int64(in))
	case int:
		return types.Int64Value(int64(in))
	case int8:
		return types.Int64Value(int64(in))
	case int16:
		return types.Int64Value(int64(in))
	case int32:
		return types.Int64Value(int64(in))
	case int64:
		return types.Int64Value(in)
	case uint:
		return types.Int64Value(int64(in))
	case uint8:
		return types.Int64Value(int64(in))
	case uint16:
		return types.Int64Value(int64(in))
	case uint32:
		return types.Int64Value(int64(in))
	case uint64:
		return types.Int64Value(int64(in))

	default:
		diags.AddError("cannot convert interface to int64",
			fmt.Sprintf("value is type %s", reflect.TypeOf(in).String()))
	}

	return types.Int64Null()
}

// Int64PointerValue returns a types.Int64 based on a pointer to any signed or
// unsigned integer.
func Int64PointerValue[T constraints.Integer](v *T) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}

	return types.Int64Value(int64(*v))
}

func Ipv4AddrValue(v net.IP) iptypes.IPv4Address {
	if v == nil || v.String() == "<nil>" {
		return iptypes.NewIPv4AddressNull()
	}

	return iptypes.NewIPv4AddressValue(v.String())
}

func Ipv6AddrValue(v net.IP) iptypes.IPv6Address {
	if v == nil || v.String() == "<nil>" {
		return iptypes.NewIPv6AddressNull()
	}

	return iptypes.NewIPv6AddressValue(v.String())
}

func Ipv4PrefixPointerValue(v *net.IPNet) cidrtypes.IPv4Prefix {
	if v == nil || v.String() == "<nil>" {
		return cidrtypes.NewIPv4PrefixNull()
	}

	return cidrtypes.NewIPv4PrefixValue(v.String())
}

func Ipv6PrefixPointerValue(v *net.IPNet) cidrtypes.IPv6Prefix {
	if v == nil || v.String() == "<nil>" {
		return cidrtypes.NewIPv6PrefixNull()
	}

	return cidrtypes.NewIPv6PrefixValue(v.String())
}

func Int64AttrValueFromPtr[A constraints.Integer](a *A) types.Int64 {
	if a == nil {
		return types.Int64Null()
	}

	return types.Int64Value(int64(*a))
}
