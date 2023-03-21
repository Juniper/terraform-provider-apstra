package apstravalidator

import (
	"context"
	"fmt"
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

func (o ParseIpValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	ip := net.ParseIP(value)
	if ip == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"input validation error",
			fmt.Sprintf("is %q an IP address?", req.ConfigValue.String()))
	}

	if o.requireIpv4 && len(ip) != 4 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"input validation error",
			fmt.Sprintf("is %q an IPv4 address?", req.ConfigValue.String()))
	}

	if o.requireIpv4 && len(ip) != 16 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"input validation error",
			fmt.Sprintf("is %q an IPv6 address?", req.ConfigValue.String()))
	}
}

func ParseIp(requireIpv4 bool, requireIpv6 bool) validator.String {
	return ParseIpValidator{
		requireIpv4: requireIpv4,
		requireIpv6: requireIpv6,
	}
}
