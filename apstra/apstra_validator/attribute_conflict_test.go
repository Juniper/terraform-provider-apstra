package apstravalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"testing"
)

func TestAttributeConflictValidator_ValidateSet(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		keyAttrNames []string
		attrTypes    map[string]attr.Type
		attrValues   []attr.Value
		expectError  bool
	}

	testCases := map[string]testCase{
		"one_key_no_collision": {
			keyAttrNames: []string{"key1"},
			attrTypes: map[string]attr.Type{
				"key1": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("bar"),
					},
				),
			},
			expectError: false,
		},
		"one_key_collision": {
			keyAttrNames: []string{"key1"},
			attrTypes: map[string]attr.Type{
				"key1": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
					},
				),
			},
			expectError: true,
		},
		"one_key_plus_extras_no_collision": {
			keyAttrNames: []string{"key1"},
			attrTypes: map[string]attr.Type{
				"key1":   types.StringType,
				"extra1": types.StringType,
				"extra2": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1":   types.StringType,
						"extra1": types.StringType,
						"extra2": types.StringType,
					},
					map[string]attr.Value{
						"key1":   types.StringValue("foo"),
						"extra1": types.StringValue("foo"),
						"extra2": types.StringValue("foo"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1":   types.StringType,
						"extra1": types.StringType,
						"extra2": types.StringType,
					},
					map[string]attr.Value{
						"key1":   types.StringValue("bar"),
						"extra1": types.StringValue("bar"),
						"extra2": types.StringValue("bar"),
					},
				),
			},
			expectError: false,
		},
		"one_key_plus_extras_collision": {
			keyAttrNames: []string{"key1"},
			attrTypes: map[string]attr.Type{
				"key1":   types.StringType,
				"extra1": types.StringType,
				"extra2": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1":   types.StringType,
						"extra1": types.StringType,
						"extra2": types.StringType,
					},
					map[string]attr.Value{
						"key1":   types.StringValue("foo"),
						"extra1": types.StringValue("bar"),
						"extra2": types.StringValue("baz"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1":   types.StringType,
						"extra1": types.StringType,
						"extra2": types.StringType,
					},
					map[string]attr.Value{
						"key1":   types.StringValue("foo"),
						"extra1": types.StringValue("bar"),
						"extra2": types.StringValue("baz"),
					},
				),
			},
			expectError: true,
		},
		"three_keys_no_collision": {
			keyAttrNames: []string{"key1", "key2", "key3"},
			attrTypes: map[string]attr.Type{
				"key1": types.StringType,
				"key2": types.StringType,
				"key3": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
						"key2": types.StringType,
						"key3": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
						"key2": types.StringValue("bar"),
						"key3": types.StringValue("baz"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
						"key2": types.StringType,
						"key3": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("bar"),
						"key2": types.StringValue("baz"),
						"key3": types.StringValue("foo"),
					},
				),
			},
			expectError: false,
		},
		"three_keys_with_extras_no_collision": {
			keyAttrNames: []string{"key1", "key2", "key3"},
			attrTypes: map[string]attr.Type{
				"key1":   types.StringType,
				"key2":   types.StringType,
				"key3":   types.StringType,
				"extra1": types.StringType,
				"extra2": types.StringType,
				"extra3": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1":   types.StringType,
						"key2":   types.StringType,
						"key3":   types.StringType,
						"extra1": types.StringType,
						"extra2": types.StringType,
						"extra3": types.StringType,
					},
					map[string]attr.Value{
						"key1":   types.StringValue("foo"),
						"key2":   types.StringValue("bar"),
						"key3":   types.StringValue("baz"),
						"extra1": types.StringValue("foo"),
						"extra2": types.StringValue("bar"),
						"extra3": types.StringValue("baz"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1":   types.StringType,
						"key2":   types.StringType,
						"key3":   types.StringType,
						"extra1": types.StringType,
						"extra2": types.StringType,
						"extra3": types.StringType,
					},
					map[string]attr.Value{
						"key1":   types.StringValue("bar"),
						"key2":   types.StringValue("baz"),
						"key3":   types.StringValue("foo"),
						"extra1": types.StringValue("foo"),
						"extra2": types.StringValue("bar"),
						"extra3": types.StringValue("baz"),
					},
				),
			},
			expectError: false,
		},
		"three_keys_collision": {
			keyAttrNames: []string{"key1", "key2", "key3"},
			attrTypes: map[string]attr.Type{
				"key1": types.StringType,
				"key2": types.StringType,
				"key3": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
						"key2": types.StringType,
						"key3": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
						"key2": types.StringValue("bar"),
						"key3": types.StringValue("baz"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
						"key2": types.StringType,
						"key3": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
						"key2": types.StringValue("bar"),
						"key3": types.StringValue("baz"),
					},
				),
			},
			expectError: true,
		},
		"three_keys_with_extras_collision": {
			keyAttrNames: []string{"key1", "key2", "key3"},
			attrTypes: map[string]attr.Type{
				"key1":   types.StringType,
				"key2":   types.StringType,
				"key3":   types.StringType,
				"extra1": types.StringType,
				"extra2": types.StringType,
				"extra3": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1":   types.StringType,
						"key2":   types.StringType,
						"key3":   types.StringType,
						"extra1": types.StringType,
						"extra2": types.StringType,
						"extra3": types.StringType,
					},
					map[string]attr.Value{
						"key1":   types.StringValue("foo"),
						"key2":   types.StringValue("bar"),
						"key3":   types.StringValue("baz"),
						"extra1": types.StringValue("foo"),
						"extra2": types.StringValue("bar"),
						"extra3": types.StringValue("baz"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1":   types.StringType,
						"key2":   types.StringType,
						"key3":   types.StringType,
						"extra1": types.StringType,
						"extra2": types.StringType,
						"extra3": types.StringType,
					},
					map[string]attr.Value{
						"key1":   types.StringValue("foo"),
						"key2":   types.StringValue("bar"),
						"key3":   types.StringValue("baz"),
						"extra1": types.StringValue("foo"),
						"extra2": types.StringValue("bar"),
						"extra3": types.StringValue("baz"),
					},
				),
			},
			expectError: true,
		},
		"all_keys_no_collision": {
			keyAttrNames: []string{},
			attrTypes: map[string]attr.Type{
				"key1": types.StringType,
				"key2": types.StringType,
				"key3": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
						"key2": types.StringType,
						"key3": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
						"key2": types.StringValue("bar"),
						"key3": types.StringValue("baz"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
						"key2": types.StringType,
						"key3": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("bar"),
						"key2": types.StringValue("baz"),
						"key3": types.StringValue("foo"),
					},
				),
			},
			expectError: false,
		},
		"all_keys_collision": {
			keyAttrNames: []string{},
			attrTypes: map[string]attr.Type{
				"key1": types.StringType,
				"key2": types.StringType,
				"key3": types.StringType,
			},
			attrValues: []attr.Value{
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
						"key2": types.StringType,
						"key3": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
						"key2": types.StringValue("bar"),
						"key3": types.StringValue("baz"),
					},
				),
				types.ObjectValueMust(
					map[string]attr.Type{
						"key1": types.StringType,
						"key2": types.StringType,
						"key3": types.StringType,
					},
					map[string]attr.Value{
						"key1": types.StringValue("foo"),
						"key2": types.StringValue("bar"),
						"key3": types.StringValue("baz"),
					},
				),
			},
			expectError: true,
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			request := validator.SetRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    types.SetValueMust(types.ObjectType{AttrTypes: tCase.attrTypes}, tCase.attrValues),
			}
			response := validator.SetResponse{}
			validator := UniquteValueCombinationsAt(tCase.keyAttrNames...)
			validator.ValidateSet(ctx, request, &response)

			if !response.Diagnostics.HasError() && tCase.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !tCase.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}

			if response.Diagnostics.HasError() {
				for _, error := range response.Diagnostics.Errors() {
					log.Println(validator.Description(ctx))
					log.Println(error.Summary())
					log.Println(error.Detail())
				}
			}
		})
	}
}
