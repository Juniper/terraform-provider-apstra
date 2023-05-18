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
	"strings"
)

var _ NineTypesValidator = mustBeWhenValueAtValidator{}

// mustBeWhenValueAtValidator validates that each set member validates against each of the value validators.
type mustBeWhenValueAtValidator struct {
	expression        path.Expression
	triggerValue      attr.Value
	boolValidators    []validator.Bool
	float64Validators []validator.Float64
	int64Validators   []validator.Int64
	listValidators    []validator.List
	mapValidators     []validator.Map
	numberValidators  []validator.Number
	objectValidators  []validator.Object
	setValidators     []validator.Set
	stringValidators  []validator.String
}

type mustBeWhenValueAtValidatorRequest struct {
	Path           path.Path
	PathExpression path.Expression
	Config         tfsdk.Config
	ConfigValue    attr.Value
}

type mustBeWhenValueAtValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o mustBeWhenValueAtValidator) Description(ctx context.Context) string {
	var descriptions []string

	for _, v := range o.boolValidators {
		descriptions = append(descriptions, v.Description(ctx))
	}
	for _, v := range o.float64Validators {
		descriptions = append(descriptions, v.Description(ctx))
	}
	for _, v := range o.int64Validators {
		descriptions = append(descriptions, v.Description(ctx))
	}
	for _, v := range o.listValidators {
		descriptions = append(descriptions, v.Description(ctx))
	}
	for _, v := range o.mapValidators {
		descriptions = append(descriptions, v.Description(ctx))
	}
	for _, v := range o.numberValidators {
		descriptions = append(descriptions, v.Description(ctx))
	}
	for _, v := range o.objectValidators {
		descriptions = append(descriptions, v.Description(ctx))
	}
	for _, v := range o.setValidators {
		descriptions = append(descriptions, v.Description(ctx))
	}
	for _, v := range o.stringValidators {
		descriptions = append(descriptions, v.Description(ctx))
	}

	return fmt.Sprintf("element must satisfy all validations: %s when value at %q is %q",
		strings.Join(descriptions, " + "), o.expression.Resolve().String(), o.triggerValue)
}

func (o mustBeWhenValueAtValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o mustBeWhenValueAtValidator) validate(ctx context.Context, req mustBeWhenValueAtValidatorRequest, resp *mustBeWhenValueAtValidatorResponse) {
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

			if !o.triggerValue.Equal(mpVal) {
				resp.Diagnostics.Append(
					validatordiag.InvalidAttributeValueDiagnostic(
						mp, fmt.Sprintf("value must be %q", o.triggerValue.String()), mpVal.String()),
				)
			}
		}
	}
}

func (o mustBeWhenValueAtValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o mustBeWhenValueAtValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o mustBeWhenValueAtValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o mustBeWhenValueAtValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o mustBeWhenValueAtValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o mustBeWhenValueAtValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o mustBeWhenValueAtValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o mustBeWhenValueAtValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o mustBeWhenValueAtValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateRequest := mustBeWhenValueAtValidatorRequest{
		Path:           req.Path,
		PathExpression: req.PathExpression,
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
	}
	validateResponse := mustBeWhenValueAtValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

// MustBeWhenValueAtBool ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtBool(e path.Expression, v attr.Value, validators ...validator.Bool) validator.Bool {
	return mustBeWhenValueAtValidator{
		expression:     e,
		triggerValue:   v,
		boolValidators: validators,
	}
}

// MustBeWhenValueAtFloat64 ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtFloat64(e path.Expression, v attr.Value, validators ...validator.Float64) validator.Float64 {
	return mustBeWhenValueAtValidator{
		expression:        e,
		triggerValue:      v,
		float64Validators: validators,
	}
}

// MustBeWhenValueAtInt64 ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtInt64(e path.Expression, v attr.Value, validators ...validator.Int64) validator.Int64 {
	return mustBeWhenValueAtValidator{
		expression:      e,
		triggerValue:    v,
		int64Validators: validators,
	}
}

// MustBeWhenValueAtList ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtList(e path.Expression, v attr.Value, validators ...validator.List) validator.List {
	return mustBeWhenValueAtValidator{
		expression:     e,
		triggerValue:   v,
		listValidators: validators,
	}
}

// MustBeWhenValueAtMap ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtMap(e path.Expression, v attr.Value, validators ...validator.Map) validator.Map {
	return mustBeWhenValueAtValidator{
		expression:    e,
		triggerValue:  v,
		mapValidators: validators,
	}
}

// MustBeWhenValueAtNumber ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtNumber(e path.Expression, v attr.Value, validators ...validator.Number) validator.Number {
	return mustBeWhenValueAtValidator{
		expression:       e,
		triggerValue:     v,
		numberValidators: validators,
	}
}

// MustBeWhenValueAtObject ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtObject(e path.Expression, v attr.Value, validators ...validator.Object) validator.Object {
	return mustBeWhenValueAtValidator{
		expression:       e,
		triggerValue:     v,
		objectValidators: validators,
	}
}

// MustBeWhenValueAtSet ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtSet(e path.Expression, v attr.Value, validators ...validator.Set) validator.Set {
	return mustBeWhenValueAtValidator{
		expression:    e,
		triggerValue:  v,
		setValidators: validators,
	}
}

// MustBeWhenValueAtString ensures that the value configured for a different attribute has the specified value.
func MustBeWhenValueAtString(e path.Expression, trigger attr.Value, validators ...validator.String) validator.String {
	return mustBeWhenValueAtValidator{
		expression:       e,
		triggerValue:     trigger,
		stringValidators: validators,
	}
}
