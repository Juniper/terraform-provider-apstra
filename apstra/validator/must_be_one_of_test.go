package apstravalidator_test

import (
	"context"
	"math/big"
	"testing"

	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMustBeOneOfValidator(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		req       apstravalidator.MustBeOneOfValidatorRequest
		oneOf     []attr.Value
		expErrors bool
	}

	testCases := map[string]testCase{
		"bool_negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.BoolValue(true),
			},
			oneOf:     []attr.Value{types.BoolValue(true)},
			expErrors: false,
		},
		"bool_positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.BoolValue(true),
			},
			oneOf:     []attr.Value{types.BoolNull()},
			expErrors: true,
		},
		"float_negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Float64Value(1),
			},
			expErrors: false,
			oneOf:     []attr.Value{types.Float64Value(1), types.Float64Value(1.5)},
		},
		"float_positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Float64Unknown(),
			},
			expErrors: true,
			oneOf:     []attr.Value{types.Float64Value(.5), types.Float64Value(1.5)},
		},
		"number_negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Int64Value(1),
			},
			expErrors: false,
			oneOf:     []attr.Value{types.Int64Value(1), types.Int64Value(2)},
		},
		"number_positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.NumberValue(big.NewFloat(1)),
			},
			expErrors: true,
			oneOf:     []attr.Value{types.NumberValue(big.NewFloat(1.1)), types.NumberValue(big.NewFloat(1.2))},
		},
		"string_negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.StringValue("1"),
			},
			expErrors: false,
			oneOf:     []attr.Value{types.StringValue("1"), types.StringValue("2")},
		},
		"string_positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Int64Value(2),
			},
			expErrors: true,
			oneOf:     []attr.Value{types.StringValue("3"), types.StringValue("2")},
		},
		"int64_negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Int64Value(1),
			},
			expErrors: false,
			oneOf:     []attr.Value{types.Int64Value(1), types.Int64Value(2)},
		},
		"int64_positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Int64Value(1),
			},
			expErrors: true,
			oneOf:     []attr.Value{types.Int64Value(3), types.Int64Value(2)},
		},
	}

	for name, test := range testCases {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res apstravalidator.MustBeOneOfValidatorResponse

			v := apstravalidator.MustBeOneOf(test.oneOf)
			v.(apstravalidator.MustBeOneOfValidator).Validate(ctx, test.req, &res)

			if test.expErrors && !res.Diagnostics.HasError() {
				t.Fatal("expected error(s), got none")
			}

			if !test.expErrors && res.Diagnostics.HasError() {
				t.Fatalf("not expecting errors, got %d: %v", res.Diagnostics.ErrorsCount(), res.Diagnostics)
			}
		})
	}
}
