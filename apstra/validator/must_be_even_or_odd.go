package apstravalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Int64 = MustBeEvenOrOddValidator{}

type MustBeEvenOrOddValidator struct {
	even bool
}

func (o MustBeEvenOrOddValidator) Description(_ context.Context) string {
	if o.even {
		return "Ensures that the supplied value is even"
	} else {
		return "Ensures that the supplied value is odd"
	}
}

func (o MustBeEvenOrOddValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o MustBeEvenOrOddValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueInt64()

	if o.even && value%2 != 0 {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				"value must be even",
				req.ConfigValue.String(),
			))
	}

	if !o.even && value%2 != 1 {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				"value must be odd",
				req.ConfigValue.String(),
			))
	}
}

func MustBeEvenOrOdd(even bool) validator.Int64 {
	return MustBeEvenOrOddValidator{
		even: even,
	}
}
