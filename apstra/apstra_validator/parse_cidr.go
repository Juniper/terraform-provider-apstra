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

var _ validator.String = ParseCidrValidator{}

type ParseCidrValidator struct {
	requireIpv4 bool
	requireIpv6 bool
}

type ParseCidrValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type ParseCidrValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o ParseCidrValidator) Description(_ context.Context) string {
	switch {
	case o.requireIpv4 && o.requireIpv6:
		return "Ensures that the supplied value can be parsed as both an IPv4 and IPv6 prefix - this usage is likely a mistake in the provider code"
	case o.requireIpv4:
		return "Ensures that the supplied can be parsed as an IPv4 prefix"
	case o.requireIpv6:
		return "Ensures that the supplied can be parsed as an IPv6 prefix"
	default:
		return "Ensures that the supplied can be parsed as either an IPv4 or IPv6 prefix"
	}
}

func (o ParseCidrValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ParseCidrValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	ip, _, err := net.ParseCIDR(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("error parsing %q as a CIDR-notation prefix", value),
			err.Error())
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

func ParseCidr(requireIpv4 bool, requireIpv6 bool) validator.String {
	return ParseCidrValidator{
		requireIpv4: requireIpv4,
		requireIpv6: requireIpv6,
	}
}
