package apstravalidator

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ NineTypesValidator = whenValueSetValidator{}

// whenValueSetValidator validates that each set member validates against each of the value validators.
type whenValueSetValidator struct {
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

func (o whenValueSetValidator) Description(ctx context.Context) string {
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

	return fmt.Sprintf("element must satisfy all validations: %s because value is set", strings.Join(descriptions, " + "))
}

func (o whenValueSetValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o whenValueSetValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.boolValidators {
		vResponse := new(validator.BoolResponse)
		v.ValidateBool(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueSetValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.float64Validators {
		vResponse := new(validator.Float64Response)
		v.ValidateFloat64(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueSetValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.int64Validators {
		vResponse := new(validator.Int64Response)
		v.ValidateInt64(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueSetValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.listValidators {
		vResponse := new(validator.ListResponse)
		v.ValidateList(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueSetValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.mapValidators {
		vResponse := new(validator.MapResponse)
		v.ValidateMap(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueSetValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.numberValidators {
		vResponse := new(validator.NumberResponse)
		v.ValidateNumber(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueSetValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.objectValidators {
		vResponse := new(validator.ObjectResponse)
		v.ValidateObject(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueSetValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.setValidators {
		vResponse := new(validator.SetResponse)
		v.ValidateSet(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueSetValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// unknown or null value means there's nothing to do
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, v := range o.stringValidators {
		vResponse := new(validator.StringResponse)
		v.ValidateString(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is set", req.Path), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

// WhenValueSetBool runs the supplied validators only when the configured
// value is set.
func WhenValueSetBool(validators ...validator.Bool) validator.Bool {
	return whenValueSetValidator{
		boolValidators: validators,
	}
}

// WhenValueSetFloat64 runs the supplied validators only when the configured
// value is set.
func WhenValueSetFloat64(validators ...validator.Float64) validator.Float64 {
	return whenValueSetValidator{
		float64Validators: validators,
	}
}

// WhenValueSetInt64 runs the supplied validators only when the configured
// value is set.
func WhenValueSetInt64(validators ...validator.Int64) validator.Int64 {
	return whenValueSetValidator{
		int64Validators: validators,
	}
}

// WhenValueSetList runs the supplied validators only when the configured
// value is set.
func WhenValueSetList(validators ...validator.List) validator.List {
	return whenValueSetValidator{
		listValidators: validators,
	}
}

// WhenValueSetMap runs the supplied validators only when the configured
// value is set.
func WhenValueSetMap(validators ...validator.Map) validator.Map {
	return whenValueSetValidator{
		mapValidators: validators,
	}
}

// WhenValueSetNumber runs the supplied validators only when the configured
// value is set.
func WhenValueSetNumber(validators ...validator.Number) validator.Number {
	return whenValueSetValidator{
		numberValidators: validators,
	}
}

// WhenValueSetObject runs the supplied validators only when the configured
// value is set.
func WhenValueSetObject(validators ...validator.Object) validator.Object {
	return whenValueSetValidator{
		objectValidators: validators,
	}
}

// WhenValueSetSet runs the supplied validators only when the configured
// value is set.
func WhenValueSetSet(validators ...validator.Set) validator.Set {
	return whenValueSetValidator{
		setValidators: validators,
	}
}

// WhenValueSetString runs the supplied validators only when the configured
// value is set.
func WhenValueSetString(validators ...validator.String) validator.String {
	return whenValueSetValidator{
		stringValidators: validators,
	}
}
