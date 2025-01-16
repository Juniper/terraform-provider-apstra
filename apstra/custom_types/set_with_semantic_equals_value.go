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
	_ basetypes.SetValuable                   = (*SetWithSemanticEquals)(nil)
	_ basetypes.SetValuableWithSemanticEquals = (*SetWithSemanticEquals)(nil)
	_ xattr.ValidateableAttribute             = (*SetWithSemanticEquals)(nil)
)

type SetWithSemanticEquals struct {
	basetypes.SetValue
	ignoreLength bool
}

func (s SetWithSemanticEquals) SetSemanticEquals(ctx context.Context, other basetypes.SetValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	o, ok := other.(SetWithSemanticEquals)
	if !ok {
		return false, diags
	}

	// A set with no elementType is an invalid state
	if s.ElementType(ctx) == nil || o.ElementType(ctx) == nil {
		return false, diags
	}

	if !s.ElementType(ctx).Equal(o.ElementType(ctx)) {
		return false, diags
	}

	if s.IsNull() != o.IsNull() {
		return false, diags
	}

	if s.IsUnknown() != o.IsUnknown() {
		return false, diags
	}

	if s.IsNull() || s.IsUnknown() {
		return true, diags
	}

	if !s.ignoreLength && len(s.Elements()) != len(o.Elements()) {
		return false, diags
	}

	for _, elem := range s.Elements() {
		if !o.contains(ctx, elem, &diags) {
			return false, diags
		}
	}

	for _, elem := range o.Elements() {
		if !s.contains(ctx, elem, &diags) {
			return false, diags
		}
	}

	return true, diags
}

func (s SetWithSemanticEquals) contains(ctx context.Context, v attr.Value, diags *diag.Diagnostics) bool {
	for _, elem := range s.Elements() {
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

func (s SetWithSemanticEquals) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	v := s.ElementType(ctx).ValueType(ctx)

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
		fmt.Sprintf("Members of SetWithSemanticEquals must implement semantic equality. Type %T is present "+
			"in the set, but it does not implement semantic equality. This is always an error in the provider. ", v),
	)
}

func NewSetWithSemanticEqualsNull(elementType attr.Type) SetWithSemanticEquals {
	return SetWithSemanticEquals{
		SetValue:     basetypes.NewSetNull(elementType),
		ignoreLength: false,
	}
}

func NewSetWithSemanticEqualsUnknown(elementType attr.Type) SetWithSemanticEquals {
	return SetWithSemanticEquals{
		SetValue:     basetypes.NewSetUnknown(elementType),
		ignoreLength: false,
	}
}

func NewSetWithSemanticEqualsValue(elementType attr.Type, elements []attr.Value) (SetWithSemanticEquals, diag.Diagnostics) {
	setValue, diags := basetypes.NewSetValue(elementType, elements)
	if diags.HasError() {
		return NewSetWithSemanticEqualsUnknown(elementType), diags
	}

	return SetWithSemanticEquals{
		SetValue:     setValue,
		ignoreLength: false,
	}, diags
}

func NewSetWithSemanticEqualsValueFrom(ctx context.Context, elementType attr.Type, elements any) (SetWithSemanticEquals, diag.Diagnostics) {
	setValue, diags := basetypes.NewSetValueFrom(ctx, elementType, elements)
	if diags.HasError() {
		return NewSetWithSemanticEqualsUnknown(elementType), diags
	}

	return SetWithSemanticEquals{
		SetValue:     setValue,
		ignoreLength: false,
	}, diags
}

func NewSetWithSemanticEqualsValueMust(elementType attr.Type, elements []attr.Value) SetWithSemanticEquals {
	return SetWithSemanticEquals{
		SetValue:     basetypes.NewSetValueMust(elementType, elements),
		ignoreLength: false,
	}
}
