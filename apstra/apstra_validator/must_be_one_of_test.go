package apstravalidator_test

import (
	"context"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math/big"
	"testing"
)

func TestMustBeOneOfValidator(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		req       apstravalidator.MustBeOneOfValidatorRequest
		expErrors bool
	}

	testCases := map[string]testCase{
		"bool positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.BoolValue(true),
				OneOf:       []attr.Value{types.BoolValue(true)},
			},
			expErrors: false,
		},
		"bool negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.BoolValue(true),
				OneOf:       []attr.Value{types.BoolValue(false)},
			},
			expErrors: true,
		},
		"float positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.Float64Value(1),
				OneOf:       []attr.Value{types.Float64Value(1), types.Float64Value(1.5)},
			},
			expErrors: false,
		},
		"float negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.Float64Value(1),
				OneOf:       []attr.Value{types.Float64Value(.5), types.Float64Value(1.5)},
			},
			expErrors: true,
		},
		"number positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.Int64Value(1),
				OneOf:       []attr.Value{types.Int64Value(1), types.Int64Value(2)},
			},
			expErrors: false,
		},
		"number negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.NumberValue(big.NewFloat(1)),
				OneOf:       []attr.Value{types.NumberValue(big.NewFloat(1.1)), types.NumberValue(big.NewFloat(1.2))},
			},
			expErrors: true,
		},
		"string negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.StringValue("1"),
				OneOf:       []attr.Value{types.StringValue("1"), types.StringValue("2")},
			},
			expErrors: false,
		}, "string positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.StringValue("1"),
				OneOf:       []attr.Value{types.StringValue("3"), types.StringValue("2")},
			},
			expErrors: true,
		},
		"int64 negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.Int64Value(1),
				OneOf:       []attr.Value{types.Int64Value(1), types.Int64Value(2)},
			},
			expErrors: false,
		}, "list positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.ListValue(types.Int64Type, []attr.Value{}),
				OneOf:       []attr.Value{types.nt64Value(3), types.Int64Value(2)},
			},
			expErrors: true,
		},
		"list negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.Int64Value(1),
				OneOf:       []attr.Value{types.Int64Value(1), types.Int64Value(2)},
			},
			expErrors: false,
		}, "int64 positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.Int64Value(1),
				OneOf:       []attr.Value{types.nt64Value(3), types.Int64Value(2)},
			},
			expErrors: true,
		},
		"int64 negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.Int64Value(1),
				OneOf:       []attr.Value{types.Int64Value(1), types.Int64Value(2)},
			},
			expErrors: false,
		}, "int64 positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				Path:        path.Path{},
				ConfigValue: types.Int64Value(1),
				OneOf:       []attr.Value{types.nt64Value(3), types.Int64Value(2)},
			},
			expErrors: true,
		},

		// "catch_forbidden_value": {
		// 	req: apstravalidator.ForbiddenWhenValueIsRequest{
		// 		Path:           path.Root("bar"),
		// 		PathExpression: path.MatchRoot("bar"),
		// 		ConfigValue:    types.StringValue("bar value"),
		// 		Config: tfsdk.Config{
		// 			Schema: schema.Schema{
		// 				Attributes: map[string]schema.Attribute{
		// 					"foo": schema.Int64Attribute{},
		// 					"bar": schema.StringAttribute{},
		// 				},
		// 			},
		// 			Raw: tftypes.NewValue(tftypes.Object{
		// 				AttributeTypes: map[string]tftypes.Type{
		// 					"foo": tftypes.Number,
		// 					"bar": tftypes.String,
		// 				},
		// 			}, map[string]tftypes.Value{
		// 				"foo": tftypes.NewValue(tftypes.Number, 42),
		// 				"bar": tftypes.NewValue(tftypes.String, "bar value"),
		// 			}),
		// 		},
		// 	},
		// 	other:      path.MatchRoot("foo"),
		// 	otherValue: types.Int64Value(42),
		// 	expErrors:  true,
		// },
		// "catch_forbidden_null": {
		// 	req: apstravalidator.ForbiddenWhenValueIsRequest{
		// 		Path:           path.Root("bar"),
		// 		PathExpression: path.MatchRoot("bar"),
		// 		ConfigValue:    types.StringValue("bar value"),
		// 		Config: tfsdk.Config{
		// 			Schema: schema.Schema{
		// 				Attributes: map[string]schema.Attribute{
		// 					"foo": schema.Int64Attribute{},
		// 					"bar": schema.StringAttribute{},
		// 				},
		// 			},
		// 			Raw: tftypes.NewValue(tftypes.Object{
		// 				AttributeTypes: map[string]tftypes.Type{
		// 					"foo": tftypes.Number,
		// 					"bar": tftypes.String,
		// 				},
		// 			}, map[string]tftypes.Value{
		// 				"foo": tftypes.NewValue(tftypes.Number, nil),
		// 				"bar": tftypes.NewValue(tftypes.String, "bar value"),
		// 			}),
		// 		},
		// 	},
		// 	other:      path.MatchRoot("foo"),
		// 	otherValue: types.Int64Null(),
		// 	expErrors:  true,
		// },
	}

	for name, test := range testCases {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res apstravalidator.MustBeOneOfValidatorResponse

			apstravalidator.MustBeOneOfValidator{
				OneOf: test.req.OneOf,
			}.Validate(ctx, test.req, &res)

			if test.expErrors && !res.Diagnostics.HasError() {
				t.Fatal("expected error(s), got none")
			}

			if !test.expErrors && res.Diagnostics.HasError() {
				t.Fatalf("not expecting errors, got %d: %v", res.Diagnostics.ErrorsCount(), res.Diagnostics)
			}
		})
	}
}
