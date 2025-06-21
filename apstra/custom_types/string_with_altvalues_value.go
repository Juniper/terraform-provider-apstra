package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ basetypes.StringValuableWithSemanticEquals = (*StringWithAltValues)(nil)

type StringWithAltValues struct {
	basetypes.StringValue
	altValues []basetypes.StringValue
}

func (v StringWithAltValues) Equal(o attr.Value) bool {
	other, ok := o.(StringWithAltValues)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals implements the semantic equality check. According to this
// (https://discuss.hashicorp.com/t/can-semantic-equality-check-in-custom-types-be-asymmetrical/60644/2?u=hqnvylrx)
// semantic equality checks on custom types are always implementeed as oldValue.SemanticEquals(ctx, newValue)
func (v StringWithAltValues) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(StringWithAltValues)
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

	// check new value against our "main" value
	if v.Equal(newValue) {
		return true, diags
	}

	// check new value against our "alt" values
	for _, altValue := range v.altValues {
		if altValue.Equal(newValue.StringValue) {
			return true, diags
		}
	}

	// check our value against new "alt" values
	for _, altValue := range newValue.altValues {
		if altValue.Equal(v.StringValue) {
			return true, diags
		}
	}

	return false, diags
}

func (v StringWithAltValues) Type(_ context.Context) attr.Type {
	return StringWithAltValuesType{}
}

func NewStringWithAltValuesNull() StringWithAltValues {
	return StringWithAltValues{
		StringValue: basetypes.NewStringNull(),
	}
}

func NewStringWithAltValuesUnknown() StringWithAltValues {
	return StringWithAltValues{
		StringValue: basetypes.NewStringUnknown(),
	}
}

func NewStringWithAltValuesValue(value string, alt ...string) StringWithAltValues {
	altValues := make([]types.String, len(alt))
	for i, a := range alt {
		altValues[i] = types.StringValue(a)
	}

	return StringWithAltValues{
		StringValue: basetypes.NewStringValue(value),
		altValues:   altValues,
	}
}
