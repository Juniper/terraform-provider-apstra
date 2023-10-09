package blueprint

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"net"
)

var _ validator.Object = &PrefixFilterValidator{}

// PrefixFilterValidator is the underlying struct implementing DifferentFrom.
type PrefixFilterValidator struct {
	PathExpressions path.Expressions
}

func (o PrefixFilterValidator) Description(ctx context.Context) string {
	return o.MarkdownDescription(ctx)
}

func (o PrefixFilterValidator) MarkdownDescription(_ context.Context) string {
	return "Ensure that `prefix`, `ge_mask` and `le_mask` fit together into a valid route filter"
}

func (o PrefixFilterValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var pf prefixFilter
	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &pf, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if pf.Prefix.IsUnknown() || pf.LeMask.IsUnknown() || pf.GeMask.IsUnknown() {
		return
	}

	ip, ipNet, err := net.ParseCIDR(pf.Prefix.ValueString())
	if err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			"value is not a valid CIDR notation prefix",
			pf.Prefix.ValueString()))
		return
	}

	if !ipNet.IP.Equal(ip) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			fmt.Sprintf("value is not a valid CIDR base address (did you mean %q?)", ipNet.String()),
			pf.Prefix.ValueString(),
		))
	}

	prefixLen, maskBits := ipNet.Mask.Size()

	var leKnown, geKnown bool
	if !pf.LeMask.IsNull() && !pf.LeMask.IsUnknown() {
		leKnown = true
	}
	if !pf.GeMask.IsNull() && !pf.GeMask.IsUnknown() {
		geKnown = true
	}

	if geKnown && pf.GeMask.ValueInt64() > int64(maskBits) { // ge_mask bigger than v4/v6 numbers?
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, fmt.Sprintf("'ge_mask' must be <= %d", maskBits), pf.GeMask.String()))
	}

	if geKnown && pf.GeMask.ValueInt64() <= int64(prefixLen) {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(req.Path,
				fmt.Sprintf("'ge_mask' must be larger/longer/more specific than prefix length %d", prefixLen),
				pf.GeMask.String()))
	}

	if leKnown && pf.LeMask.ValueInt64() > int64(maskBits) { // le_mask bigger than v4/v6 numbers?
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, fmt.Sprintf("'le_mask' must be <= %d", maskBits), pf.LeMask.String()))
	}

	if leKnown && pf.LeMask.ValueInt64() <= int64(prefixLen) {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(req.Path,
				fmt.Sprintf("'le_mask' must be larger/longer/more specific than prefix length %d", prefixLen),
				pf.LeMask.String()))
	}

	if geKnown && leKnown && pf.GeMask.ValueInt64() >= pf.LeMask.ValueInt64() {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(req.Path,
				fmt.Sprintf("'le_mask' must be larger/longer/more specific than 'ge_mask' (%d)", prefixLen),
				pf.LeMask.String()))
	}
}

func prefixFilterValidator() validator.Object {
	return &PrefixFilterValidator{}
}
