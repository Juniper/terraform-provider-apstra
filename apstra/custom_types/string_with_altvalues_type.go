package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.StringTypable = (*StringWithAltValuesType)(nil)

type StringWithAltValuesType struct {
	basetypes.StringType
}

// Equal returns true if the given type is equivalent.
func (o StringWithAltValuesType) Equal(other attr.Type) bool {
	_, ok := other.(StringWithAltValuesType)

	return ok
}

func (o StringWithAltValuesType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return NewStringWithAltValuesUnknown(), nil
	}

	if in.IsNull() {
		return NewStringWithAltValuesNull(), nil
	}

	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}

	return NewStringWithAltValuesValue(s), nil
}

// ValueType returns the Value type.
func (o StringWithAltValuesType) ValueType(_ context.Context) attr.Value {
	return StringWithAltValues{}
}

// String returns a human readable string of the type name.
func (o StringWithAltValuesType) String() string {
	return "customtypes.StringWithAltValues"
}

// ValueFromString returns a StringValuable type given a StringValue.
func (o StringWithAltValuesType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return StringWithAltValues{
		StringValue: in,
	}, nil
}
