package apstravalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"net"
)

var _ validator.String = ParseMacValidator{}

type ParseMacValidator struct{}

func (o ParseMacValidator) Description(_ context.Context) string {
	return "Ensures that the supplied can be parsed as a MAC address"
}

func (o ParseMacValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ParseMacValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	hwAddr, err := net.ParseMAC(req.ConfigValue.ValueString())
	if err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			"Value must be a valid, 48-bit MAC address",
			req.ConfigValue.ValueString(),
		))
	}

	if len(hwAddr) != 6 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			"Value must be a valid, 48-bit MAC address",
			req.ConfigValue.ValueString(),
		))
	}
}

func ParseMac() validator.String {
	return ParseMacValidator{}
}
