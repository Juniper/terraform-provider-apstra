package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*IPv46PrefixType)(nil)
	_ attr.Type               = (*IPv46PrefixType)(nil)
)

type IPv46PrefixType struct {
	basetypes.StringType
}

// String returns a human readable string of the type name.
func (t IPv46PrefixType) String() string {
	return "customtypes.IPv46PrefixType"
}

// ValueType returns the Value type.
func (t IPv46PrefixType) ValueType(_ context.Context) attr.Value {
	return IPv46Prefix{}
}

// Equal returns true if the given type is equivalent.
func (t IPv46PrefixType) Equal(o attr.Type) bool {
	other, ok := o.(IPv46PrefixType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t IPv46PrefixType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return IPv46Prefix{
		StringValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t IPv46PrefixType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
