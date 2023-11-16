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

// This type of validator must satisfy all types.
var (
	_ NineTypesValidator = AtMostNOfValidator{}
)

type NineTypesValidator interface {
	validator.Bool
	validator.Float64
	validator.Int64
	validator.List
	validator.Map
	validator.Number
	validator.Object
	validator.Set
	validator.String
}

// AtMostNOfValidator is the underlying struct implementing AtMostNOf.
type AtMostNOfValidator struct {
	N               int
	PathExpressions path.Expressions
}

type AtMostNOfValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	N              int
	Path           path.Path
	PathExpression path.Expression
}

type AtMostNOfValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o AtMostNOfValidator) Description(ctx context.Context) string {
	return o.MarkdownDescription(ctx)
}

func (o AtMostNOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that at most %d attributes from this collection is set: %s", o.N, o.PathExpressions)
}

func (o AtMostNOfValidator) Validate(ctx context.Context, req AtMostNOfValidatorRequest, resp *AtMostNOfValidatorResponse) {
	expressions := req.PathExpression.MergeExpressions(o.PathExpressions...)

	var notNullPaths []path.Path
	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		resp.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			var mpVal attr.Value
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if !mpVal.IsNull() {
				notNullPaths = append(notNullPaths, mp)
			}
		}
	}

	if len(notNullPaths) <= req.N {
		return // this is the desired outcome: fewer non-null paths than the limit.
	}

	resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
		req.Path,
		fmt.Sprintf("At most %d attributes out of %s may be specified, but %d non-null attributes were found",
			req.N, expressions, len(notNullPaths)),
	))
}

func (o AtMostNOfValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtMostNOfValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtMostNOfValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtMostNOfValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtMostNOfValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtMostNOfValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtMostNOfValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtMostNOfValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o AtMostNOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := AtMostNOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		N:              o.N,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostNOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

// AtMostNOf checks that of a set of path.Expression,
// including the attribute this validator is applied to,
// at most 'n' have a non-null value.
//
// Any relative path.Expression will be resolved using the attribute being
// validated.
func AtMostNOf(n int, expressions ...path.Expression) NineTypesValidator {
	return AtMostNOfValidator{
		N:               n,
		PathExpressions: expressions,
	}
}
