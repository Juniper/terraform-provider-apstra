package apstravalidator

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"strings"
)

type CollectionValidator interface {
	//validator.List
	//validator.Map
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

//func (o attributeConflictValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
//	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
//		return
//	}
//
//	uniqueStrings := make(map[string]bool)
//
//	for i, element := range req.ConfigValue.Elements() {
//		objectPath := req.Path.AtListIndex(i)
//
//		objectValuable, ok := element.(basetypes.ObjectValuable)
//		if !ok {
//			resp.Diagnostics.AddAttributeError(
//				req.Path,
//				"Invalid Validator for Element Value",
//				"While performing schema-based validation, an unexpected error occurred. "+
//					"The attribute declares a Object values validator, however its values do not implement the types.ObjectValuable interface for custom Object types. "+
//					"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
//					fmt.Sprintf("Path: %s\n", req.Path.String())+
//					fmt.Sprintf("Element Type: %T\n", req.ConfigValue.ElementType(ctx))+
//					fmt.Sprintf("Element Value Type: %T\n", element),
//			)
//
//			return
//		}
//
//		objectValue, diags := objectValuable.ToObjectValue(ctx)
//		resp.Diagnostics.Append(diags...)
//		if diags.HasError() {
//			return
//		}
//
//		for attrName, attrValue := range objectValue.Attributes() {
//			if attrName != o.keyAttrs || attrValue.IsUnknown() {
//				continue
//			}
//
//			if uniqueStrings[attrValue.String()] {
//				resp.Diagnostics.AddAttributeError(
//					objectPath,
//					fmt.Sprintf("%s collision", o.keyAttrs),
//					fmt.Sprintf("Two objects cannot use the same %s", o.keyAttrs),
//				)
//			} else {
//				uniqueStrings[attrValue.String()] = true
//			}
//			break // the correct attribute has been found; move on to the next
//		}
//	}
//}
//
//func (o attributeConflictValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
//	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
//		return
//	}
//
//	uniqueStrings := make(map[string]bool)
//
//	for mapKey, element := range req.ConfigValue.Elements() {
//		objectPath := req.Path.AtMapKey(mapKey)
//
//		objectValuable, ok := element.(basetypes.ObjectValuable)
//		if !ok {
//			resp.Diagnostics.AddAttributeError(
//				req.Path,
//				"Invalid Validator for Element Value",
//				"While performing schema-based validation, an unexpected error occurred. "+
//					"The attribute declares a Object values validator, however its values do not implement the types.ObjectValuable interface for custom Object types. "+
//					"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
//					fmt.Sprintf("Path: %s\n", req.Path.String())+
//					fmt.Sprintf("Element Type: %T\n", req.ConfigValue.ElementType(ctx))+
//					fmt.Sprintf("Element Value Type: %T\n", element),
//			)
//
//			return
//		}
//
//		objectValue, diags := objectValuable.ToObjectValue(ctx)
//		resp.Diagnostics.Append(diags...)
//		if diags.HasError() {
//			return
//		}
//
//		for attrName, attrValue := range objectValue.Attributes() {
//			if attrName != o.keyAttrs || attrValue.IsUnknown() {
//				continue
//			}
//
//			if uniqueStrings[attrValue.String()] {
//				resp.Diagnostics.AddAttributeError(
//					objectPath,
//					fmt.Sprintf("%s collision", o.keyAttrs),
//					fmt.Sprintf("Two objects cannot use the same %s", o.keyAttrs),
//				)
//			} else {
//				uniqueStrings[attrValue.String()] = true
//			}
//			break // the correct attribute has been found; move on to the next
//		}
//	}
//}

func (o attributeConflictValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// map of key attribute names
	keyAttributeNames := make(map[string]bool, len(o.keyAttrs))
	for _, key := range o.keyAttrs {
		keyAttributeNames[key] = true
	}

	// map of found value combinations delimited by ':'
	//    base64(key1):base64(key2):...:base64(keyN)
	foundKeyValueCombinations := make(map[string]bool) // found value combinations

	for _, element := range req.ConfigValue.Elements() { // loop over set members
		objectPath := req.Path.AtSetValue(element)

		objectValuable, ok := element.(basetypes.ObjectValuable)
		if !ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Validator for Element Value",
				"While performing schema-based validation, an unexpected error occurred. "+
					"The attribute declares a Object values validator, however its values do not implement the types.ObjectValuable interface for custom Object types. "+
					"This is likely an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Path: %s\n", req.Path.String())+
					fmt.Sprintf("Element Type: %T\n", req.ConfigValue.ElementType(ctx))+
					fmt.Sprintf("Element Value Type: %T\n", element),
			)

			return
		}

		objectValue, diags := objectValuable.ToObjectValue(ctx)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		if len(o.keyAttrs) == 0 {
			// the caller didn't specify any "key" attributes, so we'll use all of them
			for k := range objectValue.Attributes() {
				o.keyAttrs = append(o.keyAttrs, k)
				keyAttributeNames[k] = true
			}
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
			sb.WriteString(keyValuesMap[o.keyAttrs[0]]) // keyAttrs always has at least 1 entry
			for i := range o.keyAttrs[1:] {
				sb.WriteString(":" + keyValuesMap[o.keyAttrs[i]])
			}

			if foundKeyValueCombinations[sb.String()] { // seen this value before?
				resp.Diagnostics.AddAttributeError(
					objectPath,
					fmt.Sprintf("%s collision", o.keyAttrs),
					fmt.Sprintf("Two objects cannot use the same %s", o.keyAttrs),
				)
			} else {
				foundKeyValueCombinations[sb.String()] = true // log the name for future collision checks
			}
			break // all of the the required attribute have been found; move on to the next set member
		}
	}
}

func UniquteValueCombinationsAt(attrNames ...string) CollectionValidator {
	return attributeConflictValidator{
		keyAttrs: attrNames,
	}
}
