package apstravalidator

import (
	"context"
	"fmt"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"strings"
)

var _ NineTypesValidator = whenValueAtMustBeValidator{}

// whenValueAtMustBeValidator validates that each set member validates against each of the value validators.
type whenValueAtMustBeValidator struct {
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

type whenValueAtMustBeValidatorRequest struct {
	Path                path.Path
	PathExpression      path.Expression
	Config              tfsdk.Config
	ConfigValue         attr.Value
	TypeSpecificRequest any
}

type whenValueAtMustBeValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o whenValueAtMustBeValidator) Description(ctx context.Context) string {
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

func (o whenValueAtMustBeValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o whenValueAtMustBeValidator) validate(ctx context.Context, req whenValueAtMustBeValidatorRequest, resp *whenValueAtMustBeValidatorResponse) {
	expressions := req.PathExpression.MergeExpressions(o.expression)
	for i := range expressions {
		expressions[i] = expressions[i].Resolve()
	}

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			var mpVal attr.Value
			diags = req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			// Nothing to do when "value at" doesn't match the trigger value
			if !o.triggerValue.Equal(mpVal) {
				continue
			}

			// collect validator diagnostics here
			var vd diag.Diagnostics

			for _, v := range o.boolValidators {
				vr := new(validator.BoolResponse)
				v.ValidateBool(ctx, req.TypeSpecificRequest.(validator.BoolRequest), vr)
				vd.Append(vr.Diagnostics...)
			}
			for _, v := range o.float64Validators {
				vr := new(validator.Float64Response)
				v.ValidateFloat64(ctx, req.TypeSpecificRequest.(validator.Float64Request), vr)
				vd.Append(vr.Diagnostics...)
			}
			for _, v := range o.int64Validators {
				vr := new(validator.Int64Response)
				v.ValidateInt64(ctx, req.TypeSpecificRequest.(validator.Int64Request), vr)
				vd.Append(vr.Diagnostics...)
			}
			for _, v := range o.listValidators {
				vr := new(validator.ListResponse)
				v.ValidateList(ctx, req.TypeSpecificRequest.(validator.ListRequest), vr)
				vd.Append(vr.Diagnostics...)
			}
			for _, v := range o.mapValidators {
				vr := new(validator.MapResponse)
				v.ValidateMap(ctx, req.TypeSpecificRequest.(validator.MapRequest), vr)
				vd.Append(vr.Diagnostics...)
			}
			for _, v := range o.numberValidators {
				vr := new(validator.NumberResponse)
				v.ValidateNumber(ctx, req.TypeSpecificRequest.(validator.NumberRequest), vr)
				vd.Append(vr.Diagnostics...)
			}
			for _, v := range o.objectValidators {
				vr := new(validator.ObjectResponse)
				v.ValidateObject(ctx, req.TypeSpecificRequest.(validator.ObjectRequest), vr)
				vd.Append(vr.Diagnostics...)
			}
			for _, v := range o.setValidators {
				vr := new(validator.SetResponse)
				v.ValidateSet(ctx, req.TypeSpecificRequest.(validator.SetRequest), vr)
				vd.Append(vr.Diagnostics...)
			}
			for _, v := range o.stringValidators {
				vr := new(validator.StringResponse)
				v.ValidateString(ctx, req.TypeSpecificRequest.(validator.StringRequest), vr)
				vd.Append(vr.Diagnostics...)
			}

			diagWrapMsg := fmt.Sprintf("When attribute %s is %s", mp, o.triggerValue)
			utils.WrapEachDiagnostic(diagWrapMsg, vd...)
			resp.Diagnostics.Append(vd...)
		}
	}
}

func (o whenValueAtMustBeValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o whenValueAtMustBeValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o whenValueAtMustBeValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o whenValueAtMustBeValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o whenValueAtMustBeValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o whenValueAtMustBeValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o whenValueAtMustBeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o whenValueAtMustBeValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (o whenValueAtMustBeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateRequest := whenValueAtMustBeValidatorRequest{
		Path:                req.Path,
		PathExpression:      req.PathExpression,
		Config:              req.Config,
		ConfigValue:         req.ConfigValue,
		TypeSpecificRequest: req,
	}
	validateResponse := whenValueAtMustBeValidatorResponse{}
	o.validate(ctx, validateRequest, &validateResponse)
	resp.Diagnostics.Append(validateResponse.Diagnostics...)
}

// WhenValueAtMustBeBool ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeBool(e path.Expression, v attr.Value, validators ...validator.Bool) validator.Bool {
	return whenValueAtMustBeValidator{
		expression:     e,
		triggerValue:   v,
		boolValidators: validators,
	}
}

// WhenValueAtMustBeFloat64 ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeFloat64(e path.Expression, v attr.Value, validators ...validator.Float64) validator.Float64 {
	return whenValueAtMustBeValidator{
		expression:        e,
		triggerValue:      v,
		float64Validators: validators,
	}
}

// WhenValueAtMustBeInt64 ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeInt64(e path.Expression, v attr.Value, validators ...validator.Int64) validator.Int64 {
	return whenValueAtMustBeValidator{
		expression:      e,
		triggerValue:    v,
		int64Validators: validators,
	}
}

// WhenValueAtMustBeList ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeList(e path.Expression, v attr.Value, validators ...validator.List) validator.List {
	return whenValueAtMustBeValidator{
		expression:     e,
		triggerValue:   v,
		listValidators: validators,
	}
}

// WhenValueAtMustBeMap ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeMap(e path.Expression, v attr.Value, validators ...validator.Map) validator.Map {
	return whenValueAtMustBeValidator{
		expression:    e,
		triggerValue:  v,
		mapValidators: validators,
	}
}

// WhenValueAtMustBeNumber ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeNumber(e path.Expression, v attr.Value, validators ...validator.Number) validator.Number {
	return whenValueAtMustBeValidator{
		expression:       e,
		triggerValue:     v,
		numberValidators: validators,
	}
}

// WhenValueAtMustBeObject ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeObject(e path.Expression, v attr.Value, validators ...validator.Object) validator.Object {
	return whenValueAtMustBeValidator{
		expression:       e,
		triggerValue:     v,
		objectValidators: validators,
	}
}

// WhenValueAtMustBeSet ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeSet(e path.Expression, v attr.Value, validators ...validator.Set) validator.Set {
	return whenValueAtMustBeValidator{
		expression:    e,
		triggerValue:  v,
		setValidators: validators,
	}
}

// WhenValueAtMustBeString ensures that the value configured for a different attribute has the specified value.
func WhenValueAtMustBeString(e path.Expression, trigger attr.Value, validators ...validator.String) validator.String {
	return whenValueAtMustBeValidator{
		expression:       e,
		triggerValue:     trigger,
		stringValidators: validators,
	}
}
