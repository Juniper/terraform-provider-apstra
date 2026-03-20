package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*StringWithEmptyEqualsNull)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*StringWithEmptyEqualsNull)(nil)
)

type StringWithEmptyEqualsNull struct {
	basetypes.StringValue
}

func (v StringWithEmptyEqualsNull) Type(_ context.Context) attr.Type {
	return StringWithEmptyEqualsNullType{}
}

func (v StringWithEmptyEqualsNull) Equal(o attr.Value) bool {
	other, ok := o.(StringWithEmptyEqualsNull)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals implements the semantic equality check. According to this
// (https://discuss.hashicorp.com/t/can-semantic-equality-check-in-custom-types-be-asymmetrical/60644/2?u=hqnvylrx)
// semantic equality checks on custom types are always implementeed as oldValue.SemanticEquals(ctx, newValue)
func (v StringWithEmptyEqualsNull) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(StringWithEmptyEqualsNull)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	// check for actual equality
	if v.Equal(newValue) {
		return true, diags
	}

	// values are semantically equal if one is "" and the other is Null
	if (v.IsNull() && newValue.ValueString() == "") || (newValue.IsNull() && v.ValueString() == "") {
		return true, diags
	}

	return false, diags
}

func NewStringWithEmptyEqualsNullNull() StringWithEmptyEqualsNull {
	return StringWithEmptyEqualsNull{
		StringValue: basetypes.NewStringNull(),
	}
}

func NewStringWithEmptyEqualsNullUnknown() StringWithEmptyEqualsNull {
	return StringWithEmptyEqualsNull{
		StringValue: basetypes.NewStringUnknown(),
	}
}

func NewStringWithEmptyEqualsNullValue(value string) StringWithEmptyEqualsNull {
	return StringWithEmptyEqualsNull{
		StringValue: basetypes.NewStringValue(value),
	}
}

// NewStringWithEmptyEqualsNullPointerValue creates a new StringWithEmptyEqualsNull from a *string.
// If value is nil, the resulting StringWithEmptyEqualsNull will be Null.
func NewStringWithEmptyEqualsNullPointerValue(value *string) StringWithEmptyEqualsNull {
	if value == nil {
		return NewStringWithEmptyEqualsNullNull()
	}

	return NewStringWithEmptyEqualsNullValue(*value)
}
