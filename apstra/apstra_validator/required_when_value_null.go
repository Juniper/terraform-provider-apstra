package apstravalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ NineTypesValidator = RequiredWhenValueNullValidator{}

type RequiredWhenValueNullValidator struct {
	expression path.Expression
}

type RequiredWhenValueNullRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type RequiredWhenValueNullResponse struct {
	Diagnostics diag.Diagnostics
}

func (o RequiredWhenValueNullValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that a value is supplied when attribute at %s is null", o.expression.String())
}

func (o RequiredWhenValueNullValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o RequiredWhenValueNullValidator) Validate(ctx context.Context, req RequiredWhenValueNullRequest, resp *RequiredWhenValueNullResponse) {
	// can't proceed while value is unknown
	if req.ConfigValue.IsUnknown() {
		return
	}

	// if we have a value there's no need for further investigation
	if !req.ConfigValue.IsNull() {
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

			// Unknown attributes can't be validated
			if mpVal.IsUnknown() {
				return
			}

			// If the specified attribute isn't null, then we're done
			if !mpVal.IsNull() {
				return
			}

			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing required attribute",
				fmt.Sprintf("Attribute %q required when attribute %q is null.", req.Path, mp),
			)
		}
	}
}

func (o RequiredWhenValueNullValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RequiredWhenValueNullValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RequiredWhenValueNullValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RequiredWhenValueNullValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RequiredWhenValueNullValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RequiredWhenValueNullValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RequiredWhenValueNullValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RequiredWhenValueNullValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o RequiredWhenValueNullValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := RequiredWhenValueNullRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}

	validateResp := &RequiredWhenValueNullResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func RequiredWhenValueNull(expression path.Expression) RequiredWhenValueNullValidator {
	return RequiredWhenValueNullValidator{
		expression: expression,
	}
}
