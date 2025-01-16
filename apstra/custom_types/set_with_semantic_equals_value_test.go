package customtypes_test

import (
	"context"
	"fmt"
	"testing"

	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestSetWithSemanticEquals_SetSemanticEquals(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val      customtypes.SetWithSemanticEquals
		other    basetypes.SetValuable
		expected bool
	}

	testCases := map[string]testCase{
		"identical": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
			expected: true,
		},
		"identical_with_null": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
				customtypes.NewStringWithAltValuesNull(),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
				customtypes.NewStringWithAltValuesNull(),
			}),
			expected: true,
		},
		"identical_with_unknown": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
				customtypes.NewStringWithAltValuesUnknown(),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
				customtypes.NewStringWithAltValuesUnknown(),
			}),
			expected: true,
		},
		"identical_with_null_and_unknown": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
				customtypes.NewStringWithAltValuesNull(),
				customtypes.NewStringWithAltValuesUnknown(),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
				customtypes.NewStringWithAltValuesNull(),
				customtypes.NewStringWithAltValuesUnknown(),
			}),
			expected: true,
		},
		"values_okay_forward": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("one"),
				customtypes.NewStringWithAltValuesValue("TWO"),
				customtypes.NewStringWithAltValuesValue("iii"),
			}),
			expected: true,
		},
		"values_okay_backward": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("one"),
				customtypes.NewStringWithAltValuesValue("TWO"),
				customtypes.NewStringWithAltValuesValue("iii"),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
			expected: true,
		},
		"values_not_okay_forward": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("four"),
				customtypes.NewStringWithAltValuesValue("FIVE"),
				customtypes.NewStringWithAltValuesValue("vi"),
			}),
			expected: false,
		},
		"values_not_okay_backward": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("four"),
				customtypes.NewStringWithAltValuesValue("FIVE"),
				customtypes.NewStringWithAltValuesValue("vi"),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
			expected: false,
		},
		"wrong_set_valuable_type": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("four"),
				customtypes.NewStringWithAltValuesValue("FIVE"),
				customtypes.NewStringWithAltValuesValue("vi"),
			}),
			other: basetypes.NewSetValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
			expected: false,
		},
		"wrong_element_type": {
			val: customtypes.NewSetWithSemanticEqualsValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
			other: customtypes.NewSetWithSemanticEqualsValueMust(types.StringType, []attr.Value{
				types.StringValue("one"),
				types.StringValue("TWO"),
				types.StringValue("iii"),
			}),
			expected: false,
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			result, diags := tCase.val.SetSemanticEquals(context.Background(), tCase.other)
			require.Equal(t, tCase.expected, result)
			require.Falsef(t, diags.HasError(), "test case produced %d diagnostic errors: %s", diags.ErrorsCount())
			require.Equalf(t, 0, diags.WarningsCount(), "test case produced %d diagnostic warnings: %s", diags.WarningsCount(), diags.ErrorsCount())
		})
	}
}

func TestSetWithSemanticEquals_ValidateAttribute(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in            basetypes.SetValue
		expectedDiags diag.Diagnostics
	}{
		"empty": {
			in: types.SetValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{}),
		},
		"non-empty": {
			in: types.SetValueMust(customtypes.StringWithAltValuesType{}, []attr.Value{
				customtypes.NewStringWithAltValuesValue("1", "one", "ONE", "i"),
				customtypes.NewStringWithAltValuesValue("2", "two", "TWO", "ii"),
				customtypes.NewStringWithAltValuesValue("3", "three", "THREE", "iii"),
			}),
		},
		"null": {
			in: types.SetNull(customtypes.StringWithAltValuesType{}),
		},
		"unknown": {
			in: types.SetUnknown(customtypes.StringWithAltValuesType{}),
		},
		"empty_wrong_type": {
			in: types.SetValueMust(types.StringType, []attr.Value{}),
			expectedDiags: []diag.Diagnostic{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid element type in set",
					fmt.Sprintf("Members of SetWithSemanticEquals must implement semantic equality. Type %T is present "+
						"in the set, but it does not implement semantic equality. This is always an error in the provider. ", types.String{}),
				),
			},
		},
		"non-empty_wrong_type": {
			in: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(2),
				types.Int64Value(3),
			}),
			expectedDiags: []diag.Diagnostic{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid element type in set",
					fmt.Sprintf("Members of SetWithSemanticEquals must implement semantic equality. Type %T is present "+
						"in the set, but it does not implement semantic equality. This is always an error in the provider. ", types.Int64{}),
				),
			},
		},
		"null_wrong_type": {
			in: types.SetNull(types.Int32Type),
			expectedDiags: []diag.Diagnostic{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid element type in set",
					fmt.Sprintf("Members of SetWithSemanticEquals must implement semantic equality. Type %T is present "+
						"in the set, but it does not implement semantic equality. This is always an error in the provider. ", types.Int32{}),
				),
			},
		},
		"unknown_wrong_type": {
			in: types.SetUnknown(types.NumberType),
			expectedDiags: []diag.Diagnostic{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid element type in set",
					fmt.Sprintf("Members of SetWithSemanticEquals must implement semantic equality. Type %T is present "+
						"in the set, but it does not implement semantic equality. This is always an error in the provider. ", types.Number{}),
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := xattr.ValidateAttributeRequest{Path: path.Root("test")}
			var response xattr.ValidateAttributeResponse
			customtypes.SetWithSemanticEquals{SetValue: testCase.in}.ValidateAttribute(context.Background(), request, &response)
			if diff := cmp.Diff(response.Diagnostics, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
