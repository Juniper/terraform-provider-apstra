package apstravalidator

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// This type of validator must satisfy all types.
var (
	_ NineTypesValidator = MustBeOneOfValidator{}
)

// MustBeOneOfValidator is the underlying struct implementing MustBeOneOf.
type MustBeOneOfValidator struct {
	oneOf []attr.Value
}

type MustBeOneOfValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	oneOf          []attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type MustBeOneOfValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (o MustBeOneOfValidator) Description(ctx context.Context) string {
	return o.MarkdownDescription(ctx)
}

func (o MustBeOneOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that the value is one of the following : %q", o.oneOf)
}

func (o MustBeOneOfValidator) Validate(ctx context.Context, req MustBeOneOfValidatorRequest, resp *MustBeOneOfValidatorResponse) {
	oneOfUs := false
	for _, v := range req.oneOf {
		if req.ConfigValue == v {
			oneOfUs = true
		}
	}

	if !oneOfUs {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			fmt.Sprintf("Must be one of : %s", attrArrToString(req.oneOf)),
			req.ConfigValue.String(),
		))
	}
}

func attrArrToString(arr []attr.Value) string {
	r := make([]string, len(arr))
	for i, j := range arr {
		r[i] = j.String()
	}
	return strings.Join(r, ",")
}

func (o MustBeOneOfValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o MustBeOneOfValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o MustBeOneOfValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o MustBeOneOfValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o MustBeOneOfValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o MustBeOneOfValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o MustBeOneOfValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o MustBeOneOfValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (o MustBeOneOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := MustBeOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		oneOf:          o.oneOf,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &MustBeOneOfValidatorResponse{}

	o.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

// MustBeOneOf checks that of a set of path.Expression,
// including the attribute this validator is applied to,
// at most 'n' have a non-null value.
//
// Any relative path.Expression will be resolved using the attribute being
// validated.
func MustBeOneOf(oneOf []attr.Value) NineTypesValidator {
	return MustBeOneOfValidator{
		oneOf: oneOf,
	}
}
