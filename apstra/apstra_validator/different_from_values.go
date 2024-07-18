package apstravalidator

import (
	"context"
	"fmt"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ validator.Bool    = DifferentFromValuesValidator{}
	_ validator.Float64 = DifferentFromValuesValidator{}
	_ validator.Int64   = DifferentFromValuesValidator{}
	_ validator.List    = DifferentFromValuesValidator{}
	_ validator.Map     = DifferentFromValuesValidator{}
	_ validator.Number  = DifferentFromValuesValidator{}
	_ validator.Object  = DifferentFromValuesValidator{}
	_ validator.Set     = DifferentFromValuesValidator{}
	_ validator.String  = DifferentFromValuesValidator{}
)

// DifferentFromValuesValidator is the underlying struct implementing DifferentFromValues.
type DifferentFromValuesValidator struct {
	PathExpressions path.Expressions
}

type DifferentFromValuesValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type DifferentFromValuesValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o DifferentFromValuesValidator) Description(ctx context.Context) string {
	return o.MarkdownDescription(ctx)
}

func (o DifferentFromValuesValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that if an attribute is set, these don't share the same value: %q", o.PathExpressions)
}

func (o DifferentFromValuesValidator) Validate(ctx context.Context, req DifferentFromValuesValidatorRequest, resp *DifferentFromValuesValidatorResponse) {
	// If attribute configuration isn't known, there is nothing else to validate
	if !utils.HasValue(req.ConfigValue) {
		return
	}

	for _, expression := range req.PathExpression.MergeExpressions(o.PathExpressions...) {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		// loop over matched paths. Each one should be a "collection" (list/map/set)
		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this apstra_validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(req.Path) {
				continue
			}

			var mpVal attr.Value
			diags = req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)

			// Unknown and Null attributes can't have value collisions
			if mpVal.IsNull() || mpVal.IsUnknown() {
				continue
			}

			var elements []attr.Value

			// extract the collection values into elements
			switch mpVal := mpVal.(type) {
			case basetypes.ListValue:
				elements = mpVal.Elements()
			case basetypes.MapValue:
				mapElements := mpVal.Elements()
				elements = make([]attr.Value, len(mapElements))
				var i int
				for _, v := range elements {
					elements[i] = v
					i++
				}
			case basetypes.SetValue:
				elements = mpVal.Elements()
			default:
				panic(fmt.Sprintf("DifferentFromValuesValidator unhandled attr.Value implementation: %t", mpVal))
			}

			// compare each element against the value in question
			for _, v := range elements {
				// Unknown and Null attributes can't have value collisions
				if v.IsNull() || v.IsUnknown() {
					continue
				}

				if req.ConfigValue.Equal(v) {
					resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
						req.Path,
						fmt.Sprintf("attribute %q must not have any values which match attribute %q (%s)", mp, req.Path, req.ConfigValue.String()),
					))
				}
			}
		}
	}
}

// DifferentFromValues checks that none of the values of the collection
// (list/map/set) attributes found at the supplied expressions are Equal() to
// this attribute's value.
func DifferentFromValues(expressions ...path.Expression) *DifferentFromValuesValidator {
	return &DifferentFromValuesValidator{
		PathExpressions: expressions,
	}
}

func (o DifferentFromValuesValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o DifferentFromValuesValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o DifferentFromValuesValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o DifferentFromValuesValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o DifferentFromValuesValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o DifferentFromValuesValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o DifferentFromValuesValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o DifferentFromValuesValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o DifferentFromValuesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := DifferentFromValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &DifferentFromValuesValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}
