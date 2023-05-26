package apstravalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"net"
)

var _ validator.String = ParseIpOrCidrValidator{}

type ParseIpOrCidrValidator struct {
	requireIpv4 bool
	requireIpv6 bool
}

func (o ParseIpOrCidrValidator) Description(_ context.Context) string {
	switch {
	case o.requireIpv4 && o.requireIpv6:
		return "Ensures that the supplied value can be parsed as both an IPv4 and IPv6 address or prefix - this usage is likely a mistake in the provider code"
	case o.requireIpv4:
		return "Ensures that the supplied value can be parsed as an IPv4 address or prefix"
	case o.requireIpv6:
		return "Ensures that the supplied value can be parsed as an IPv6 address or prefix"
	default:
		return "Ensures that the supplied value can be parsed as either an IPv4 or IPv6 address or prefix"
	}
}

func (o ParseIpOrCidrValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ParseIpOrCidrValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var ip net.IP
	var ipNet *net.IPNet
	var err error

	value := req.ConfigValue.ValueString()

	// try parsing as an IP address (not CIDR notation)
	ip = net.ParseIP(value)
	if ip == nil {

		// try parsing as a CIDR prefix
		ip, ipNet, err = net.ParseCIDR(value)
		if err != nil || ip == nil || ipNet == nil {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				"value is not a valid CIDR notation prefix",
				value))
			return
		}

		// ensure it's the zero address
		if !ipNet.IP.Equal(ip) {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				fmt.Sprintf("value is not a valid CIDR base address (did you mean %q?)", ipNet.String()),
				value,
			))
		}
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

func ParseIpOrCidr(requireIpv4 bool, requireIpv6 bool) validator.String {
	return ParseIpOrCidrValidator{
		requireIpv4: requireIpv4,
		requireIpv6: requireIpv6,
	}
}
