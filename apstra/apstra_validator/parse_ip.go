package apstravalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"net"
)

var _ validator.String = ParseIpValidator{}

type ParseIpValidator struct {
	requireIpv4 bool
	requireIpv6 bool
}

type ParseIpValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type ParseIpValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o ParseIpValidator) Description(_ context.Context) string {
	switch {
	case o.requireIpv4 && o.requireIpv6:
		return "Ensures that the supplied value can be parsed as both an IPv4 and IPv6 address - this usage is likely a mistake in the provider code"
	case o.requireIpv4:
		return "Ensures that the supplied can be parsed as an IPv4 address"
	case o.requireIpv6:
		return "Ensures that the supplied can be parsed as an IPv6 address"
	default:
		return "Ensures that the supplied can be parsed as either an IPv4 or IPv6 address"
	}
}

func (o ParseIpValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ParseIpValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	ip := net.ParseIP(value)
	if ip == nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, "value must be an IP address", value))
		return
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

func ParseIp(requireIpv4 bool, requireIpv6 bool) validator.String {
	return ParseIpValidator{
		requireIpv4: requireIpv4,
		requireIpv6: requireIpv6,
	}
}
