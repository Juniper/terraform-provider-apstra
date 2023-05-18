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

var _ NineTypesValidator = valueAtMustBeValidator{}

// valueAtMustBeValidator validates that each set member validates against each of the value validators.
type valueAtMustBeValidator struct {
	expression path.Expression
	value      attr.Value
	nullOk     bool
}

type valueAtMustBeValidatorRequest struct {
	Path           path.Path
	PathExpression path.Expression
	Config         tfsdk.Config
	ConfigValue    attr.Value
}

type valueAtMustBeValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o valueAtMustBeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("element at %q must be: %q (null is %t)", o.expression, o.value.String(), o.nullOk)
}

func (o valueAtMustBeValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o valueAtMustBeValidator) validate(ctx context.Context, req valueAtMustBeValidatorRequest, resp *valueAtMustBeValidatorResponse) {
	expressions := req.PathExpression.MergeExpressions(o.expression)
	for i := range expressions {
		expressions[i] = expressions[i].Resolve()
	}

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, mp := range matchedPaths {
			var mpVal attr.Value
			diags = req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if o.nullOk && mpVal.IsNull() {
				return
			}

			if !o.value.Equal(mpVal) {
				resp.Diagnostics.Append(
					validatordiag.InvalidAttributeValueDiagnostic(
						mp,
						fmt.Sprintf("must be %s", o.value),
						mpVal.String()),
				)
			}
		}
	}
}

func (o valueAtMustBeValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o valueAtMustBeValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o valueAtMustBeValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o valueAtMustBeValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o valueAtMustBeValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o valueAtMustBeValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o valueAtMustBeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o valueAtMustBeValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o valueAtMustBeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateRequest := valueAtMustBeValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := valueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

// ValueAtMustBeBool ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeBool(e path.Expression, v attr.Value, nullOk bool) validator.Bool {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}

// ValueAtMustBeFloat64 ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeFloat64(e path.Expression, v attr.Value, nullOk bool) validator.Float64 {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}

// ValueAtMustBeInt64 ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeInt64(e path.Expression, v attr.Value, nullOk bool) validator.Int64 {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}

// ValueAtMustBeList ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeList(e path.Expression, v attr.Value, nullOk bool) validator.List {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}

// ValueAtMustBeMap ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeMap(e path.Expression, v attr.Value, nullOk bool) validator.Map {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}

// ValueAtMustBeNumber ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeNumber(e path.Expression, v attr.Value, nullOk bool) validator.Number {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}

// ValueAtMustBeObject ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeObject(e path.Expression, v attr.Value, nullOk bool) validator.Object {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}

// ValueAtMustBeSet ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeSet(e path.Expression, v attr.Value, nullOk bool) validator.Set {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}

// ValueAtMustBeString ensures that the value configured for a different attribute has the specified value.
func ValueAtMustBeString(e path.Expression, v attr.Value, nullOk bool) validator.String {
	return valueAtMustBeValidator{
		expression: e,
		value:      v,
		nullOk:     nullOk,
	}
}
