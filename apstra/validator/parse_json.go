package apstravalidator

import (
	"context"

	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = ParseJsonValidator{}

type ParseJsonValidator struct{}

func (o ParseJsonValidator) Description(_ context.Context) string {
	return "Ensures that the supplied value is a valid JSON"
}

func (o ParseJsonValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ParseJsonValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !utils.IsJSON(req.ConfigValue) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path, "value must be JSON", req.ConfigValue.ValueString()))
	}
}

func ParseJson() validator.String {
	return ParseJsonValidator{}
}
