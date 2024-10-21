package apstravalidator

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ NineTypesValidator = whenValueIsValidator{}

// whenValueIsValidator validates that each set member validates against each of the value validators.
type whenValueIsValidator struct {
	trigger           attr.Value
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

func (o whenValueIsValidator) Description(ctx context.Context) string {
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

	return fmt.Sprintf("element must satisfy all validations: %s becuase value is %s", strings.Join(descriptions, " + "), o.trigger)
}

func (o whenValueIsValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o whenValueIsValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.boolValidators {
		vResponse := new(validator.BoolResponse)
		v.ValidateBool(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueIsValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.float64Validators {
		vResponse := new(validator.Float64Response)
		v.ValidateFloat64(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueIsValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.int64Validators {
		vResponse := new(validator.Int64Response)
		v.ValidateInt64(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueIsValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.listValidators {
		vResponse := new(validator.ListResponse)
		v.ValidateList(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueIsValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.mapValidators {
		vResponse := new(validator.MapResponse)
		v.ValidateMap(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueIsValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.numberValidators {
		vResponse := new(validator.NumberResponse)
		v.ValidateNumber(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueIsValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.objectValidators {
		vResponse := new(validator.ObjectResponse)
		v.ValidateObject(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueIsValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.setValidators {
		vResponse := new(validator.SetResponse)
		v.ValidateSet(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

func (o whenValueIsValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// validating unknown values doesn't make sense, but null is fair game
	if req.ConfigValue.IsUnknown() {
		return
	}

	// further validation not required if config value doesn't match the trigger
	if !o.trigger.Equal(req.ConfigValue) {
		return
	}

	for _, v := range o.stringValidators {
		vResponse := new(validator.StringResponse)
		v.ValidateString(ctx, req, vResponse)
		utils.WrapEachDiagnostic(fmt.Sprintf("When attribute %s is %s", req.Path, o.trigger), vResponse.Diagnostics...)
		resp.Diagnostics.Append(vResponse.Diagnostics...)
	}
}

// WhenValueIsBool runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsBool(trigger types.Bool, validators ...validator.Bool) validator.Bool {
	return whenValueIsValidator{
		trigger:        trigger,
		boolValidators: validators,
	}
}

// WhenValueIsFloat64 runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsFloat64(trigger types.Float64, validators ...validator.Float64) validator.Float64 {
	return whenValueIsValidator{
		trigger:           trigger,
		float64Validators: validators,
	}
}

// WhenValueIsInt64 runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsInt64(trigger types.Int64, validators ...validator.Int64) validator.Int64 {
	return whenValueIsValidator{
		trigger:         trigger,
		int64Validators: validators,
	}
}

// WhenValueIsList runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsList(trigger types.List, validators ...validator.List) validator.List {
	return whenValueIsValidator{
		trigger:        trigger,
		listValidators: validators,
	}
}

// WhenValueIsMap runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsMap(trigger types.Map, validators ...validator.Map) validator.Map {
	return whenValueIsValidator{
		trigger:       trigger,
		mapValidators: validators,
	}
}

// WhenValueIsNumber runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsNumber(trigger types.Number, validators ...validator.Number) validator.Number {
	return whenValueIsValidator{
		trigger:          trigger,
		numberValidators: validators,
	}
}

// WhenValueIsObject runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsObject(trigger types.Object, validators ...validator.Object) validator.Object {
	return whenValueIsValidator{
		trigger:          trigger,
		objectValidators: validators,
	}
}

// WhenValueIsSet runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsSet(trigger types.Set, validators ...validator.Set) validator.Set {
	return whenValueIsValidator{
		trigger:       trigger,
		setValidators: validators,
	}
}

// WhenValueIsString runs the supplied validators only when the configured
// value is equal to the trigger value.
func WhenValueIsString(trigger types.String, validators ...validator.String) validator.String {
	return whenValueIsValidator{
		trigger:          trigger,
		stringValidators: validators,
	}
}
