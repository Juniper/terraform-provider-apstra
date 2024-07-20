package customtypes

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*IPv46Prefix)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*IPv46Prefix)(nil)
	_ xattr.ValidateableAttribute                = (*IPv46Prefix)(nil)
)

type IPv46Prefix struct {
	basetypes.StringValue
}

func (v IPv46Prefix) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if !v.Type(ctx).TerraformType(ctx).Is(tftypes.String) {
		err := fmt.Errorf("expected String value, received %T with value: %v", v, v)
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"IPv46 Prefix Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+err.Error(),
		)
	}

	if v.IsNull() || v.IsUnknown() {
		return
	}

	valueString := v.ValueString()

	prefix, err := netip.ParsePrefix(valueString)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv46 Prefix String Value",
			"A string value was provided that is not valid IPv4 or IPv6 prefix string format.\n\n"+
				"Given Value: "+valueString+"\n"+
				"Error: "+err.Error(),
		)

		return
	}

	if !prefix.IsValid() && (!prefix.Addr().Is4() && !prefix.Addr().Is6()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv46 Prefix String Value",
			"A string value was provided that is not valid IPv4 or IPv6 prefix string format.\n\n"+
				"Given Value: "+valueString+"\n",
		)

		return
	}

	if prefix.Masked().Addr().Compare(prefix.Addr()) != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv46 Prefix String Value",
			"A string value was provided that does not represent a base address of an IPv4 or IPv6 prefix.\n\n"+
				"Given Value: "+valueString+"\n"+
				"Base Address: "+prefix.Masked().String(),
		)

		return
	}
}

func (v IPv46Prefix) Type(_ context.Context) attr.Type {
	return IPv46PrefixType{}
}

func (v IPv46Prefix) Equal(o attr.Value) bool {
	other, ok := o.(IPv46Prefix)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v IPv46Prefix) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(IPv46Prefix)
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

	newPrefix, _ := netip.ParsePrefix(newValue.ValueString())
	currentPrefix, _ := netip.ParsePrefix(v.ValueString())

	return currentPrefix.String() == newPrefix.String(), diags
}

func (v IPv46Prefix) ValueIPv46Prefix() (netip.Prefix, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("IPv46Prefix ValueIPv46Prefix Error", "prefix string value is null"))
		return netip.Prefix{}, diags
	}

	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("IPv46Prefix ValueIPv46Prefix Error", "prefix string value is unknown"))
		return netip.Prefix{}, diags
	}

	ipv46Prefix, err := netip.ParsePrefix(v.ValueString())
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("IPv46Prefix ValueIPv46Prefix Error", err.Error()))
		return netip.Prefix{}, diags
	}

	return ipv46Prefix, nil
}

func (v IPv46Prefix) Is4() bool {
	if v.IsUnknown() || v.IsNull() {
		return false
	}

	prefix, _ := netip.ParsePrefix(v.ValueString())
	return prefix.IsValid() && prefix.Addr().Is4()
}

func (v IPv46Prefix) Is6() bool {
	if v.IsUnknown() || v.IsNull() {
		return false
	}

	prefix, _ := netip.ParsePrefix(v.ValueString())
	return prefix.IsValid() && prefix.Addr().Is6()
}

func NewIPv46PrefixNull() IPv46Prefix {
	return IPv46Prefix{
		StringValue: basetypes.NewStringNull(),
	}
}

func NewIPv46PrefixUnknown() IPv46Prefix {
	return IPv46Prefix{
		StringValue: basetypes.NewStringUnknown(),
	}
}

func NewIPv46PrefixValue(value string) IPv46Prefix {
	return IPv46Prefix{
		StringValue: basetypes.NewStringValue(value),
	}
}

func NewIPv46PrefixPointerValue(value *string) IPv46Prefix {
	return IPv46Prefix{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}
