package apstravalidator

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = ParseCidrValidator{}

type ParseCidrValidator struct {
	requireIpv4 bool
	requireIpv6 bool
}

func (o ParseCidrValidator) Description(_ context.Context) string {
	switch {
	case o.requireIpv4 && o.requireIpv6:
		return "Ensures that the supplied value can be parsed as both an IPv4 and IPv6 prefix - this usage is likely a mistake in the provider code"
	case o.requireIpv4:
		return "Ensures that the supplied value can be parsed as an IPv4 prefix"
	case o.requireIpv6:
		return "Ensures that the supplied value can be parsed as an IPv6 prefix"
	default:
		return "Ensures that the supplied value can be parsed as either an IPv4 or IPv6 prefix"
	}
}

func (o ParseCidrValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ParseCidrValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	ip, ipNet, err := net.ParseCIDR(value)
	if err != nil || ip == nil || ipNet == nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			"value is not a valid CIDR notation prefix",
			value))
		return
	}

	if !ipNet.IP.Equal(ip) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			fmt.Sprintf("value is not a valid CIDR base address (did you mean %q?)", ipNet.String()),
			value,
		))
	}

	switch {
	case o.requireIpv4 && len(ip.To4()) != net.IPv4len:
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, "value is not an IPv4 CIDR prefix", value))
	case o.requireIpv6 && len(ip) != net.IPv6len:
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, "value is not an IPv6 CIDR prefix", value))
	}
}

func ParseCidr(requireIpv4 bool, requireIpv6 bool) validator.String {
	return ParseCidrValidator{
		requireIpv4: requireIpv4,
		requireIpv6: requireIpv6,
	}
}
