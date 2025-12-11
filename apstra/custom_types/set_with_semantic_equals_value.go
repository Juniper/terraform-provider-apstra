package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.SetValuableWithSemanticEquals = (*SetWithSemanticEqualsValue)(nil)
	_ xattr.ValidateableAttribute             = (*SetWithSemanticEqualsValue)(nil)
)

type SetWithSemanticEqualsValue struct {
	basetypes.SetValue
}

func (o SetWithSemanticEqualsValue) Equal(v attr.Value) bool {
	other, ok := v.(SetWithSemanticEqualsValue)
	if !ok {
		return false
	}

	return o.SetValue.Equal(other.SetValue)
}

func (o SetWithSemanticEqualsValue) Type(_ context.Context) attr.Type {
	return SetWithSemanticEqualsType{basetypes.SetType{ElemType: o.ElementType(nil)}}
}

func (o SetWithSemanticEqualsValue) SetSemanticEquals(ctx context.Context, v basetypes.SetValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	other, ok := v.(SetWithSemanticEqualsValue)
	if !ok {
		return false, diags
	}

	// A set with no elementType is an invalid state
	if o.ElementType(ctx) == nil || other.ElementType(ctx) == nil {
		return false, diags
	}

	if !o.ElementType(ctx).Equal(other.ElementType(ctx)) {
		return false, diags
	}

	if o.IsNull() != other.IsNull() {
		return false, diags
	}

	if o.IsUnknown() != other.IsUnknown() {
		return false, diags
	}

	if o.IsNull() || other.IsUnknown() {
		return true, diags
	}

	for _, elem := range other.Elements() {
		if !o.semanticContains(ctx, elem, &diags) {
			return false, diags
		}
	}

	for _, elem := range o.Elements() {
		if !other.semanticContains(ctx, elem, &diags) {
			return false, diags
		}
	}

	return true, diags
}

func (o SetWithSemanticEqualsValue) semanticContains(ctx context.Context, v attr.Value, diags *diag.Diagnostics) bool {
	for _, elem := range o.Elements() {
		if elem, ok := elem.(basetypes.BoolValuableWithSemanticEquals); ok {
			ok, d := elem.BoolSemanticEquals(ctx, v.(basetypes.BoolValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.DynamicValuableWithSemanticEquals); ok {
			ok, d := elem.DynamicSemanticEquals(ctx, v.(basetypes.DynamicValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.Float32ValuableWithSemanticEquals); ok {
			ok, d := elem.Float32SemanticEquals(ctx, v.(basetypes.Float32ValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.Float64ValuableWithSemanticEquals); ok {
			ok, d := elem.Float64SemanticEquals(ctx, v.(basetypes.Float64ValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.Int32ValuableWithSemanticEquals); ok {
			ok, d := elem.Int32SemanticEquals(ctx, v.(basetypes.Int32ValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.Int64ValuableWithSemanticEquals); ok {
			ok, d := elem.Int64SemanticEquals(ctx, v.(basetypes.Int64ValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.ListValuableWithSemanticEquals); ok {
			ok, d := elem.ListSemanticEquals(ctx, v.(basetypes.ListValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.MapValuableWithSemanticEquals); ok {
			ok, d := elem.MapSemanticEquals(ctx, v.(basetypes.MapValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.NumberValuableWithSemanticEquals); ok {
			ok, d := elem.NumberSemanticEquals(ctx, v.(basetypes.NumberValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.ObjectValuableWithSemanticEquals); ok {
			ok, d := elem.ObjectSemanticEquals(ctx, v.(basetypes.ObjectValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.SetValuableWithSemanticEquals); ok {
			ok, d := elem.SetSemanticEquals(ctx, v.(basetypes.SetValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}

		if elem, ok := elem.(basetypes.StringValuableWithSemanticEquals); ok {
			ok, d := elem.StringSemanticEquals(ctx, v.(basetypes.StringValuableWithSemanticEquals))
			diags.Append(d...)
			if diags.HasError() {
				return false
			}
			if ok {
				return true
			}
		}
	}

	return false
}

func (o SetWithSemanticEqualsValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	v := o.ElementType(ctx).ValueType(ctx)

	if _, ok := v.(basetypes.BoolValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.DynamicValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.Float32ValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.Float64ValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.Int32ValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.Int64ValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.ListValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.MapValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.NumberValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.ObjectValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.SetValuableWithSemanticEquals); ok {
		return
	}
	if _, ok := v.(basetypes.StringValuableWithSemanticEquals); ok {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid element type in set",
		fmt.Sprintf("Members of SetWithSemanticEqualsValue must implement semantic equality. Type %T is present "+
			"in the set, but it does not implement semantic equality. This is always an error in the provider. ", v),
	)
}

func NewSetWithSemanticEqualsNull(elementType attr.Type) SetWithSemanticEqualsValue {
	return SetWithSemanticEqualsValue{
		SetValue: basetypes.NewSetNull(elementType),
	}
}

func NewSetWithSemanticEqualsUnknown(elementType attr.Type) SetWithSemanticEqualsValue {
	return SetWithSemanticEqualsValue{
		SetValue: basetypes.NewSetUnknown(elementType),
	}
}

func NewSetWithSemanticEqualsValue(elementType attr.Type, elements []attr.Value) (SetWithSemanticEqualsValue, diag.Diagnostics) {
	setValue, diags := basetypes.NewSetValue(elementType, elements)
	if diags.HasError() {
		return NewSetWithSemanticEqualsUnknown(elementType), diags
	}

	return SetWithSemanticEqualsValue{
		SetValue: setValue,
	}, diags
}

func NewSetWithSemanticEqualsValueFrom(ctx context.Context, elementType attr.Type, elements any) (SetWithSemanticEqualsValue, diag.Diagnostics) {
	setValue, diags := basetypes.NewSetValueFrom(ctx, elementType, elements)
	if diags.HasError() {
		return NewSetWithSemanticEqualsUnknown(elementType), diags
	}

	return SetWithSemanticEqualsValue{
		SetValue: setValue,
	}, diags
}

func NewSetWithSemanticEqualsValueMust(elementType attr.Type, elements []attr.Value) SetWithSemanticEqualsValue {
	return SetWithSemanticEqualsValue{
		SetValue: basetypes.NewSetValueMust(elementType, elements),
	}
}
