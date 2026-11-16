package importvalidation

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
)

const (
	failSummaryGeneral      = "Identity validation failed"
	failSummaryMissingField = "Missing required identity field"
	failSummaryInvalidField = "Invalid identity field"
	failSummaryInvalidType  = "Invalid identity field type"
	failSummaryMissingValue = "Missing required identity value"

	failDetailNilIdentity   = "identity value is nil"
	failDetailNotStruct     = "expected identity to be a struct or pointer to struct, got: "
	failDetailNoTag         = "no struct field found with tfsdk tag: "
	failDetailCannotInspect = "unable to inspect field with tfsdk tag: "
	failDetailNotAttrValue  = "field does not implement attr.Value, tfsdk tag: "
	failDetailNullOrUnknown = "required field must be known and non-null, tfsdk tag: "
)

type identityWithAttributes interface {
	Attributes() map[string]identityschema.Attribute
}

// CheckRequiredFields ensures that all fields marked RequiredForImport in the given identity's schema are
// present and known. It uses reflection to inspect the struct fields and their corresponding tfsdk tags.
// If any required field is missing or unknown, it adds an error to the provided diagnostics.
func CheckRequiredFields(_ context.Context, identity identityWithAttributes, diags *diag.Diagnostics) {
	structValue := reflect.ValueOf(identity)

	// Dereference pointer types if necessary. Value must not be nil.
	if structValue.Kind() == reflect.Pointer {
		if structValue.IsNil() {
			diags.AddError(failSummaryGeneral, failDetailNilIdentity)
			return
		}
		structValue = structValue.Elem()
	}

	// After dereferencing, the value must be a struct.
	if structValue.Kind() != reflect.Struct {
		diags.AddError(
			failSummaryGeneral,
			failDetailNotStruct+fmt.Sprintf("%T", identity),
		)
		return
	}

	// Loop over attributes claimed by the identity's schema.
	for tfsdkTag, attribute := range identity.Attributes() {
		if !attribute.IsRequiredForImport() {
			continue // Attribute not required, so we don't care about it.
		}

		// Find the struct element which corresponds to the tfsdk tag.
		fieldValue, ok := fieldByTfsdkTag(structValue, tfsdkTag)
		if !ok {
			diags.AddError(
				failSummaryMissingField,
				failDetailNoTag+strconv.Quote(tfsdkTag),
			)
			continue
		}

		// Make sure we can inspect the struct element.
		if !fieldValue.IsValid() || !fieldValue.CanInterface() {
			diags.AddError(
				failSummaryInvalidField,
				failDetailCannotInspect+strconv.Quote(tfsdkTag),
			)
			continue
		}

		// Every struct element should implement attr.Value.
		fieldAttrValue, ok := fieldValue.Interface().(attr.Value)
		if !ok {
			diags.AddError(
				failSummaryInvalidType,
				failDetailNotAttrValue+strconv.Quote(tfsdkTag),
			)
			continue
		}

		// This is why we're here. Make sure that fields marked RequiredForImport are known and non-null.
		if fieldAttrValue.IsNull() || fieldAttrValue.IsUnknown() {
			diags.AddError(
				failSummaryMissingValue,
				failDetailNullOrUnknown+strconv.Quote(tfsdkTag),
			)
		}
	}
}

// fieldByTfsdkTag returns the field of structValue whose `tfsdk` struct tag
// name matches desired, and true. If no field matches, it returns the zero
// reflect.Value and false.
//
// Only the portion of the tag before the first comma is compared, so options
// like `tfsdk:"name,omitempty"` match "name". Fields with a missing or "-" tag
// are skipped. If several fields share a tag name, the first one wins.
//
// structValue must be of Kind reflect.Struct (not a pointer); other kinds
// cause a panic. The returned value may not be accessible via Interface()
// (e.g. an unexported field), so callers should check CanInterface().
func fieldByTfsdkTag(structValue reflect.Value, desired string) (reflect.Value, bool) {
	structType := structValue.Type()

	// Loop over the struct elemeents.
	for i := range structType.NumField() {
		// Discover the tfsdk tag for this struct element.
		tag := structType.Field(i).Tag.Get("tfsdk")
		if tag == "" || tag == "-" {
			continue // no tag or explicitly ignored: skip this field
		}

		// Compare the tag name (portion before the first comma) to the desired tag.
		if desired == strings.Split(tag, ",")[0] {
			return structValue.Field(i), true
		}
	}

	// Matching struct element not found.
	return reflect.Value{}, false
}
