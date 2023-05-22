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
)

var (
	_ NineTypesValidator = ForbiddenWhenValueIsValidator{}
)

type ForbiddenWhenValueIsValidator struct {
	expression path.Expression
	value      string
}

type ForbiddenWhenValueIsRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type ForbiddenWhenValueIsResponse struct {
	Diagnostics diag.Diagnostics
}

func (o ForbiddenWhenValueIsValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that no value is supplied when attribute at %q has value %q", o.expression.String(), o.value)
}

func (o ForbiddenWhenValueIsValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ForbiddenWhenValueIsValidator) Validate(ctx context.Context, req ForbiddenWhenValueIsRequest, resp *ForbiddenWhenValueIsResponse) {
	// can't proceed while value is unknown
	if req.ConfigValue.IsUnknown() {
		return
	}

	// if we don't have a value there's no need for further investigation
	if req.ConfigValue.IsNull() {
		return
	}

	mergedExpressions := req.PathExpression.MergeExpressions(o.expression)

	for _, expression := range mergedExpressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this apstra_validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(req.Path) {
				continue
			}

			var mpVal attr.Value
			diags = req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue // Collect all errors
			}

			// Unknown and Null attributes can't satisfy the valueIs condition
			if mpVal.IsNull() || mpVal.IsUnknown() {
				return
			}

			if mpVal.String() == o.value {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
					req.Path,
					fmt.Sprintf("value not permitted when %q has value %q, got %q", mp, mpVal, req.ConfigValue),
				))
			}
		}
	}
}

func (o ForbiddenWhenValueIsValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o ForbiddenWhenValueIsValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o ForbiddenWhenValueIsValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o ForbiddenWhenValueIsValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o ForbiddenWhenValueIsValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o ForbiddenWhenValueIsValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o ForbiddenWhenValueIsValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o ForbiddenWhenValueIsValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o ForbiddenWhenValueIsValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := ForbiddenWhenValueIsRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &ForbiddenWhenValueIsResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func BoolForbiddenWhenValueIs(expression path.Expression, value string) validator.Bool {
	return ForbiddenWhenValueIsValidator{
		expression: expression,
		value:      value,
	}
}

func Float64ForbiddenWhenValueIs(expression path.Expression, value string) validator.Float64 {
	return ForbiddenWhenValueIsValidator{
		expression: expression,
		value:      value,
	}
}

func Int64ForbiddenWhenValueIs(expression path.Expression, value string) validator.Int64 {
	return ForbiddenWhenValueIsValidator{
		expression: expression,
		value:      value,
	}
}

func MapForbiddenWhenValueIs(expression path.Expression, value string) validator.Map {
	return ForbiddenWhenValueIsValidator{
		expression: expression,
		value:      value,
	}
}

func NumberForbiddenWhenValueIs(expression path.Expression, value string) validator.Number {
	return ForbiddenWhenValueIsValidator{
		expression: expression,
		value:      value,
	}
}

func ObjectForbiddenWhenValueIs(expression path.Expression, value string) validator.Object {
	return ForbiddenWhenValueIsValidator{
		expression: expression,
		value:      value,
	}
}

func SetForbiddenWhenValueIs(expression path.Expression, value string) validator.Set {
	return ForbiddenWhenValueIsValidator{
		expression: expression,
		value:      value,
	}
}

func StringForbiddenWhenValueIs(expression path.Expression, value string) validator.String {
	return ForbiddenWhenValueIsValidator{
		expression: expression,
		value:      value,
	}
}
