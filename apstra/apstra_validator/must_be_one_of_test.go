package apstravalidator_test

import (
	"context"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math/big"
	"testing"
)

func TestMustBeOneOfValidator(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		req       apstravalidator.MustBeOneOfValidatorRequest
		OneOf     []attr.Value
		expErrors bool
	}

	testCases := map[string]testCase{
		"bool positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.BoolValue(true),
			},
			OneOf:     []attr.Value{types.BoolValue(true)},
			expErrors: false,
		},
		"bool negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.BoolValue(true),
			},
			OneOf:     []attr.Value{types.BoolNull()},
			expErrors: true,
		},
		"float positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Float64Value(1),
			},
			expErrors: false,
			OneOf:     []attr.Value{types.Float64Value(1), types.Float64Value(1.5)},
		},
		"float negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Float64Value(1),
			},
			expErrors: true,
			OneOf:     []attr.Value{types.Float64Value(.5), types.Float64Value(1.5)},
		},
		"number positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Int64Value(1),
			},
			expErrors: false,
			OneOf:     []attr.Value{types.Int64Value(1), types.Int64Value(2)},
		},
		"number negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.NumberValue(big.NewFloat(1)),
			},
			expErrors: true,
			OneOf:     []attr.Value{types.NumberValue(big.NewFloat(1.1)), types.NumberValue(big.NewFloat(1.2))},
		},
		"string negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.StringValue("1"),
			},
			expErrors: false,
			OneOf:     []attr.Value{types.StringValue("1"), types.StringValue("2")},
		},
		"string positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.StringValue("1"),
			},
			expErrors: true,
			OneOf:     []attr.Value{types.StringValue("3"), types.StringValue("2")},
		},
		"int64 negative": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Int64Value(1),
			},
			expErrors: false,
			OneOf:     []attr.Value{types.Int64Value(1), types.Int64Value(2)},
		},
		"int64 positive": {
			req: apstravalidator.MustBeOneOfValidatorRequest{
				ConfigValue: types.Int64Value(1),
			},
			expErrors: true,
			OneOf:     []attr.Value{types.Int64Value(3), types.Int64Value(2)},
		},
	}

	for name, test := range testCases {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res apstravalidator.MustBeOneOfValidatorResponse

			v := apstravalidator.MustBeOneOf(test.OneOf)
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
