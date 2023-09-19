package apstravalidator

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"strings"
)

type CollectionValidator interface {
	validator.List
	validator.Map
	validator.Set
}

var _ CollectionValidator = attributeConflictValidator{}

// attributeConflictValidator ensures that no two elements of a list, map, or
// set of objects use the same value across all attributes enumerated in
// keyAttrs.
//
// For example, if keyAttrs contains just {"name"}, then having two objects
// with `name: "foo"` will produce a validation error.
//
// If keyAttrs contains {"protocol", "port"} then having two objects with
// `protocol: "TCP"` and `port: 80` will produce a validation error.
//
// If keyAttrs is empty, then values across all attributes are evaluated.
type attributeConflictValidator struct {
	keyAttrs []string
}

func (o attributeConflictValidator) Description(_ context.Context) string {
	if len(o.keyAttrs) == 0 {
		return "Ensure that no two collection (list/map/set) members share values for all attributes"
	}

	return fmt.Sprintf(
		"Ensure that no two collection (list/map/set) members share values for these attributes: [%s]",
		strings.Join(o.keyAttrs, " "),
	)
}

func (o attributeConflictValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o attributeConflictValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	foundKeyValueCombinations := make(map[string]bool)
	for i, element := range req.ConfigValue.Elements() { // loop over set members
		validateRequest := attributeConflictValidateElementRequest{
			elementValue:              element,
			elementPath:               req.Path.AtListIndex(i),
			foundKeyValueCombinations: foundKeyValueCombinations,
			path:                      req.Path,
		}
		validateResponse := attributeConflictValidateElementResponse{}
		o.validateElement(ctx, validateRequest, &validateResponse)
		resp.Diagnostics.Append(validateResponse.Diagnostics...)
	}
}

func (o attributeConflictValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	foundKeyValueCombinations := make(map[string]bool)
	for mapKey, element := range req.ConfigValue.Elements() { // loop over set members
		validateRequest := attributeConflictValidateElementRequest{
			elementValue:              element,
			elementPath:               req.Path.AtMapKey(mapKey),
			foundKeyValueCombinations: foundKeyValueCombinations,
			path:                      req.Path,
		}
		validateResponse := attributeConflictValidateElementResponse{}
		o.validateElement(ctx, validateRequest, &validateResponse)
		resp.Diagnostics.Append(validateResponse.Diagnostics...)
	}
}

func (o attributeConflictValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	foundKeyValueCombinations := make(map[string]bool)
	for _, element := range req.ConfigValue.Elements() { // loop over set members
		validateRequest := attributeConflictValidateElementRequest{
			elementValue:              element,
			elementPath:               req.Path.AtSetValue(element),
			foundKeyValueCombinations: foundKeyValueCombinations,
			path:                      req.Path,
		}
		validateResponse := attributeConflictValidateElementResponse{}
		o.validateElement(ctx, validateRequest, &validateResponse)
		resp.Diagnostics.Append(validateResponse.Diagnostics...)
	}
}

func UniqueValueCombinationsAt(attrNames ...string) CollectionValidator {
	return attributeConflictValidator{
		keyAttrs: attrNames,
	}
}

type attributeConflictValidateElementRequest struct {
	elementValue              attr.Value
	elementPath               path.Path
	foundKeyValueCombinations map[string]bool
	path                      path.Path
}

type attributeConflictValidateElementResponse struct {
	Diagnostics diag.Diagnostics
}

func (o *attributeConflictValidator) validateElement(ctx context.Context, req attributeConflictValidateElementRequest, resp *attributeConflictValidateElementResponse) {
	objectValuable, ok := req.elementValue.(basetypes.ObjectValuable)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.path,
			"Invalid Validator for Element Value",
			"While performing schema-based validation, an unexpected error occurred. "+
				"The attribute declares a Object values validator, however its values do not implement the types.ObjectValuable interface for custom Object types. "+
				"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
				fmt.Sprintf("Path: %s\n", req.path.String())+
				fmt.Sprintf("Element Type: %T\n", req.elementValue.Type(ctx))+
				fmt.Sprintf("Element Value Type: %T\n", req.elementValue),
		)

		return
	}

	objectValue, d := objectValuable.ToObjectValue(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if the caller didn't specify any "key" attributes we use all of them
	if len(o.keyAttrs) == 0 {
		for k := range objectValue.Attributes() {
			o.keyAttrs = append(o.keyAttrs, k)
		}
	}

	// map of key attribute names used to quickly recognize whether an attribute is interesting
	keyAttributeNames := make(map[string]bool, len(o.keyAttrs))
	for _, key := range o.keyAttrs {
		keyAttributeNames[key] = true
	}

	keyValuesMap := make(map[string]string, len(keyAttributeNames))
	for attrName, attrValue := range objectValue.Attributes() { // loop over set member attributes
		if !keyAttributeNames[attrName] {
			continue // attribute is not interesting
		}

		if attrValue.IsUnknown() {
			return // cannot validate when attribute is unknown
		}

		keyValuesMap[attrName] = base64.StdEncoding.EncodeToString([]byte(attrValue.String()))
		if len(keyValuesMap) < len(keyAttributeNames) {
			continue // keep going until we fill keyValuesMap
		}

		sb := strings.Builder{}
		for i := range o.keyAttrs {
			if i == 0 {
				sb.WriteString(keyValuesMap[o.keyAttrs[i]])
			} else {
				sb.WriteString(":" + keyValuesMap[o.keyAttrs[i]])
			}
		}

		if req.foundKeyValueCombinations[sb.String()] { // seen this value before?
			resp.Diagnostics.AddAttributeError(
				req.elementPath,
				fmt.Sprintf("%s collision", o.keyAttrs),
				fmt.Sprintf("Two objects cannot use the same %s", o.keyAttrs),
			)
		} else {
			req.foundKeyValueCombinations[sb.String()] = true // log the name for future collision checks
		}
		break // all of the the required attribute have been found; move on to the next set member
	}

}
