package apstravalidator

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = ParseSpeedValidator{}

type ParseSpeedValidator struct{}

func (o ParseSpeedValidator) Description(_ context.Context) string {
	return "Ensures that user submitted speed is in format xxG"
}

func (o ParseSpeedValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ParseSpeedValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	validMap := make(map[string]struct{})
	validStrings := []string{
		"100M",
		"1G",
		"10G",
		"25G",
		"40G",
		"50G",
		"100G",
		"200G",
		"400G",
		"800G",
	}

	for _, s := range validStrings {
		validMap[s] = struct{}{}
	}

	value := req.ConfigValue.ValueString()
	if _, ok := validMap[value]; !ok {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, fmt.Sprintf("value must be one of '%s'", strings.Join(validStrings, "', '")), value))
		return
	}
}

func ParseSpeed() validator.String {
	return ParseSpeedValidator{}
}
