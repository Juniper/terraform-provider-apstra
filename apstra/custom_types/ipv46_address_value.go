package customtypes

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*IPv46Address)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*IPv46Address)(nil)
)

type IPv46Address struct {
	basetypes.StringValue
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
