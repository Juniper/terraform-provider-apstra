package apstravalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"math/big"
)

var (
	_ validator.Float64 = DifferentFromValidator{}
	_ validator.Int64   = DifferentFromValidator{}
	_ validator.Number  = DifferentFromValidator{}
)

// AtLeastProductOfValidator is the underlying struct implementing AtLeastProductOf.
type AtLeastProductOfValidator struct {
	PathExpression path.Expression
	Multiplier     float64
}

type AtLeastProductOfValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type AtLeastProductOfValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o AtLeastProductOfValidator) Description(ctx context.Context) string {
	return o.MarkdownDescription(ctx)
}

func (o AtLeastProductOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that if an attribute is set, it's value is at least "+
		"%f * any value at: %q",
		o.Multiplier, o.PathExpression)
}

func (o AtLeastProductOfValidator) Validate(ctx context.Context, req AtLeastProductOfValidatorRequest, resp *AtLeastProductOfValidatorResponse) {
	// If attribute configuration is null or unknown, there is nothing
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// convert this value to *big.Float
	var thisBF *big.Float
	switch t := req.ConfigValue.(type) {
	case basetypes.Float64Value:
		thisBF = big.NewFloat(t.ValueFloat64())
	case basetypes.Int64Value:
		thisBF = big.NewFloat(float64(t.ValueInt64()))
	case basetypes.NumberValue:
		thisBF = t.ValueBigFloat()
	}

	multiplier := big.NewFloat(o.Multiplier)

	matchedPaths, diags := req.Config.PathMatches(ctx, o.PathExpression)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	for _, mp := range matchedPaths {
		var mpVal attr.Value
		diags = req.Config.GetAttribute(ctx, mp, &mpVal)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Unknown and Null attributes can't be multiplied
		if mpVal.IsNull() || mpVal.IsUnknown() {
			continue
		}

		// convert the value we're comparing to *big.Float
		var mpBF *big.Float
		switch t := mpVal.(type) {
		case basetypes.Float64Value:
			mpBF = big.NewFloat(t.ValueFloat64())
		case basetypes.Int64Value:
			mpBF = big.NewFloat(float64(t.ValueInt64()))
		case basetypes.NumberValue:
			mpBF = t.ValueBigFloat()
		}

		if thisBF.Cmp(mpBF.Mul(mpBF, multiplier)) == -1 {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
				req.Path,
				fmt.Sprintf("value must be at least %f times the value at %s (%s), got %s",
					multiplier, mp, mpVal, req.ConfigValue),
			))
		}
	}
}

// AtLeastProductOf checks that a set of path.Expression have values equal to the
// current attribute when the current attribute is non-null.
//
// Relative path.Expression will be resolved using the attribute being
// validated.
func AtLeastProductOf(multiplier float64, expression path.Expression) *AtLeastProductOfValidator {
	return &AtLeastProductOfValidator{
		PathExpression: expression,
		Multiplier:     multiplier,
	}
}

func (o AtLeastProductOfValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateReq := AtLeastProductOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtLeastProductOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtLeastProductOfValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := AtLeastProductOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtLeastProductOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtLeastProductOfValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateReq := AtLeastProductOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtLeastProductOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}
