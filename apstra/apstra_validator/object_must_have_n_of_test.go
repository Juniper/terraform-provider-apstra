package apstravalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestObjectMustHaveNOf(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		n               int
		checkAttributes []string
		attrTypes       map[string]attr.Type
		attributes      map[string]attr.Value
		expAtLeastErr   bool
		expAtMostErr    bool
		expExactlyErr   bool
	}

	testCases := map[string]testCase{
		"zero": {
			n:               0,
			checkAttributes: nil,
			attrTypes:       nil,
			attributes:      nil,
			expAtLeastErr:   false,
			expAtMostErr:    false,
			expExactlyErr:   false,
		},
		"1:1:1": {
			n:               1,
			checkAttributes: []string{"a1"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
			},
			expAtLeastErr: false,
			expAtMostErr:  false,
			expExactlyErr: false,
		},
		"1:1:2": {
			n:               1,
			checkAttributes: []string{"a1"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
			},
			expAtLeastErr: false,
			expAtMostErr:  false,
			expExactlyErr: false,
		},
		"1:3:3": {
			n:               1,
			checkAttributes: []string{"a1", "a2", "a3"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
				"a3": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
				"a3": types.StringValue("baz"),
			},
			expAtLeastErr: false,
			expAtMostErr:  true,
			expExactlyErr: true,
		},
		"1:3:4": {
			n:               1,
			checkAttributes: []string{"a1", "a2", "a3"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
				"a3": types.StringType,
				"a4": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
				"a3": types.StringValue("baz"),
				"a4": types.StringValue("bang"),
			},
			expAtLeastErr: false,
			expAtMostErr:  true,
			expExactlyErr: true,
		},
		"2:3": {
			n:               2,
			checkAttributes: []string{"a1", "a2", "a3"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
				"a3": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
				"a3": types.StringValue("baz"),
			},
			expAtLeastErr: false,
			expAtMostErr:  true,
			expExactlyErr: true,
		},
		"2:3:4": {
			n:               2,
			checkAttributes: []string{"a1", "a2", "a3"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
				"a3": types.StringType,
				"a4": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
				"a3": types.StringValue("baz"),
				"a4": types.StringValue("bang"),
			},
			expAtLeastErr: false,
			expAtMostErr:  true,
			expExactlyErr: true,
		},
		"3:3:3": {
			n:               3,
			checkAttributes: []string{"a1", "a2", "a3"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
				"a3": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
				"a3": types.StringValue("baz"),
			},
			expAtLeastErr: false,
			expAtMostErr:  false,
			expExactlyErr: false,
		},
		"3:3:4": {
			n:               3,
			checkAttributes: []string{"a1", "a2", "a3"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
				"a3": types.StringType,
				"a4": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
				"a3": types.StringValue("baz"),
				"a4": types.StringValue("bang"),
			},
			expAtLeastErr: false,
			expAtMostErr:  false,
			expExactlyErr: false,
		},
		"4:3:4": {
			n:               4,
			checkAttributes: []string{"a1", "a2", "a3"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
				"a3": types.StringType,
				"a4": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
				"a3": types.StringValue("baz"),
				"a4": types.StringValue("bang"),
			},
			expAtLeastErr: true,
			expAtMostErr:  true,
			expExactlyErr: true,
		},
		"bad_data": {
			n:               1,
			checkAttributes: []string{"b1", "b2"},
			attrTypes: map[string]attr.Type{
				"a1": types.StringType,
				"a2": types.StringType,
			},
			attributes: map[string]attr.Value{
				"a1": types.StringValue("foo"),
				"a2": types.StringValue("bar"),
			},
			expAtLeastErr: true,
			expAtMostErr:  true,
			expExactlyErr: true,
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			var resp validator.ObjectResponse
			req := validator.ObjectRequest{
				Path:        path.Root("test"),
				ConfigValue: types.ObjectValueMust(tCase.attrTypes, tCase.attributes),
			}

			resp = validator.ObjectResponse{}
			atLeastValidator := AtLeastNAttributes(tCase.n, tCase.checkAttributes...)
			atLeastValidator.ValidateObject(ctx, req, &resp)
			if resp.Diagnostics.HasError() && !tCase.expAtLeastErr {
				t.Fatal("got an error in the 'at least' case where none was expected")
			}
			if !resp.Diagnostics.HasError() && tCase.expAtLeastErr {
				t.Fatal("got no error in the 'at least' case where one was expected")
			}

			resp = validator.ObjectResponse{}
			atMostValidator := AtMostNAttributes(tCase.n, tCase.checkAttributes...)
			atMostValidator.ValidateObject(ctx, req, &resp)
			if resp.Diagnostics.HasError() && !tCase.expAtMostErr {
				t.Fatal("got an error in the 'at most' case where none was expected")
			}
			if !resp.Diagnostics.HasError() && tCase.expAtMostErr {
				t.Fatal("got no error in the 'at most' case where one was expected")
			}

			resp = validator.ObjectResponse{}
			exactlyValidator := ExactlyNAttributes(tCase.n, tCase.checkAttributes...)
			exactlyValidator.ValidateObject(ctx, req, &resp)
			if resp.Diagnostics.HasError() && !tCase.expExactlyErr {
				t.Fatal("got an error in the 'exactly' case where none was expected")
			}
			if !resp.Diagnostics.HasError() && tCase.expExactlyErr {
				t.Fatal("got no error in the 'exactly' case where one was expected")
			}

			//atMostValidator := AtMostNAttributes(tCase.n, tCase.checkAttributes...)
			//exactlyValidator := ExactlyNAttributes(tCase.n, tCase.checkAttributes...)
		})
	}
}
