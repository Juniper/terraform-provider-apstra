package importvalidation

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
)

type identityWithAttributes interface {
	Attributes() map[string]identityschema.Attribute
}

// CheckRequiredFields ensures that all fields marked RequiredForImport in the given identity's schema are
// present and known. It uses reflection to inspect the struct fields and their corresponding tfsdk tags.
// If any required field is missing or unknown, it adds an error to the provided diagnostics.
func CheckRequiredFields(_ context.Context, identity identityWithAttributes, diags *diag.Diagnostics) {
	structValue := reflect.ValueOf(identity)
	if structValue.Kind() == reflect.Pointer {
		if structValue.IsNil() {
			diags.AddError("Identity validation failed", "identity value is nil")
			return
		}
		structValue = structValue.Elem()
	}

	if structValue.Kind() != reflect.Struct {
		diags.AddError(
			"Identity validation failed",
			fmt.Sprintf("expected identity to be a struct or pointer to struct, got %T", identity),
		)
		return
	}

	for k, v := range identity.Attributes() {
		if !v.IsRequired() {
			continue
		}

		fieldValue, ok := fieldByTfsdkTag(structValue, k)
		if !ok {
			diags.AddError(
				"Missing required identity field",
				fmt.Sprintf("no struct field found with tfsdk tag %q", k),
			)
			continue
		}

		if !fieldValue.IsValid() || !fieldValue.CanInterface() {
			diags.AddError(
				"Invalid identity field",
				fmt.Sprintf("field with tfsdk tag %q cannot be inspected", k),
			)
			continue
		}

		fieldAttrValue, ok := fieldValue.Interface().(attr.Value)
		if !ok {
			diags.AddError(
				"Invalid identity field type",
				fmt.Sprintf("field with tfsdk tag %q does not implement attr.Value", k),
			)
			continue
		}

		if fieldAttrValue.IsNull() || fieldAttrValue.IsUnknown() {
			diags.AddError(
				"Missing required identity value",
				fmt.Sprintf("field %q must be known and non-null", k),
			)
		}
	}
}

func fieldByTfsdkTag(structValue reflect.Value, key string) (reflect.Value, bool) {
	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		tag := fieldType.Tag.Get("tfsdk")
		if tag == "" || tag == "-" {
			continue
		}

		tagName := strings.Split(tag, ",")[0]
		if tagName == key {
			return structValue.Field(i), true
		}
	}

	return reflect.Value{}, false
}
