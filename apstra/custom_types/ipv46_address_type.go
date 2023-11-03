package customtypes

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*IPv46AddressType)(nil)
	_ xattr.TypeWithValidate  = (*IPv46AddressType)(nil)
)

type IPv46AddressType struct {
	basetypes.StringType
}

// String returns a human readable string of the type name.
func (t IPv46AddressType) String() string {
	return "customtypes.IPv46AddressType"
}

// ValueType returns the Value type.
func (t IPv46AddressType) ValueType(_ context.Context) attr.Value {
	return IPv46Address{}
}

// Equal returns true if the given type is equivalent.
func (t IPv46AddressType) Equal(o attr.Type) bool {
	other, ok := o.(IPv46AddressType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t IPv46AddressType) Validate(_ context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.Type() == nil {
		return diags
	}

	if !in.Type().Is(tftypes.String) {
		err := fmt.Errorf("expected String value, received %T with value: %v", in, in)
		diags.AddAttributeError(
			path,
			"IPv46 Address Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var valueString string

	if err := in.As(&valueString); err != nil {
		diags.AddAttributeError(
			path,
			"IPv46 Address Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+err.Error(),
		)

		return diags
	}

	ipAddr, err := netip.ParseAddr(valueString)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Invalid IPv46 Address String Value",
			"A string value was provided that is not valid IPv4 or IPv6 string format.\n\n"+
				"Given Value: "+valueString+"\n"+
				"Error: "+err.Error(),
		)

		return diags
	}

	if !ipAddr.IsValid() && (!ipAddr.Is4() && !ipAddr.Is6()) {
		diags.AddAttributeError(
			path,
			"Invalid IPv46 Address String Value",
			"A string value was provided that is not valid IPv4 or IPv6 string format.\n\n"+
				"Given Value: "+valueString+"\n",
		)

		return diags
	}

	return diags
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t IPv46AddressType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return IPv46Address{
		StringValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t IPv46AddressType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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
