package customtypes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*IPv46Address)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*IPv46Address)(nil)
	_ xattr.ValidateableAttribute                = (*IPv46Address)(nil)
)

type IPv46Address struct {
	basetypes.StringValue
}

func (v IPv46Address) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if !v.Type(ctx).TerraformType(ctx).Is(tftypes.String) {
		err := fmt.Errorf("expected String value, received %T with value: %v", v, v)
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"IPv46 Address Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+err.Error(),
		)
	}

	if v.IsNull() || v.IsUnknown() {
		return
	}

	valueString := v.ValueString()

	ipAddr, err := netip.ParseAddr(valueString)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv46 Address String Value",
			"A string value was provided that is not valid IPv4 or IPv6 string format.\n\n"+
				"Given Value: "+valueString+"\n"+
				"Error: "+err.Error(),
		)

		return
	}

	if !ipAddr.IsValid() && (!ipAddr.Is4() && !ipAddr.Is6()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv46 Address String Value",
			"A string value was provided that is not valid IPv4 or IPv6 string format.\n\n"+
				"Given Value: "+valueString+"\n",
		)

		return
	}
}

func (v IPv46Address) Type(_ context.Context) attr.Type {
	return IPv46AddressType{}
}

func (v IPv46Address) Equal(o attr.Value) bool {
	other, ok := o.(IPv46Address)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v IPv46Address) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(IPv46Address)
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

	newIpAddr, _ := netip.ParseAddr(newValue.ValueString())
	currentIpAddr, _ := netip.ParseAddr(v.ValueString())

	return currentIpAddr == newIpAddr, diags
}

func (v IPv46Address) ValueIPv46Address() (netip.Addr, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("IPv46Address ValueIPv46Address Error", "address string value is null"))
		return netip.Addr{}, diags
	}

	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("IPv46Address ValueIPv46Address Error", "address string value is unknown"))
		return netip.Addr{}, diags
	}

	ipv46Addr, err := netip.ParseAddr(v.ValueString())
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("IPv46Address ValueIPv46Address Error", err.Error()))
		return netip.Addr{}, diags
	}

	return ipv46Addr, nil
}

func NewIPv46AddressNull() IPv46Address {
	return IPv46Address{
		StringValue: basetypes.NewStringNull(),
	}
}

func NewIPv46AddressUnknown() IPv46Address {
	return IPv46Address{
		StringValue: basetypes.NewStringUnknown(),
	}
}

func NewIPv46AddressValue(value string) IPv46Address {
	return IPv46Address{
		StringValue: basetypes.NewStringValue(value),
	}
}

func NewIPv46AddressPointerValue(value *string) IPv46Address {
	return IPv46Address{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}
