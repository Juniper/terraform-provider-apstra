package apstravalidator_test

import (
	"context"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

func TestForbiddenWhenValueIsValidator(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		req        apstravalidator.ForbiddenWhenValueIsRequest
		other      path.Expression
		otherValue attr.Value
		expErrors  bool
	}

	testCases := map[string]testCase{
		"base": {
			req: apstravalidator.ForbiddenWhenValueIsRequest{
				Path:           path.Root("bar"),
				PathExpression: path.MatchRoot("bar"),
				ConfigValue:    types.StringValue("bar value"),
				Config: tfsdk.Config{
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"foo": schema.Int64Attribute{},
							"bar": schema.StringAttribute{},
						},
					},
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"foo": tftypes.Number,
							"bar": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.Number, 42),
						"bar": tftypes.NewValue(tftypes.String, "bar value"),
					}),
				},
			},
			other:      path.MatchRoot("foo"),
			otherValue: types.Int64Value(0),
			expErrors:  false,
		},
		"catch_forbidden_value": {
			req: apstravalidator.ForbiddenWhenValueIsRequest{
				Path:           path.Root("bar"),
				PathExpression: path.MatchRoot("bar"),
				ConfigValue:    types.StringValue("bar value"),
				Config: tfsdk.Config{
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"foo": schema.Int64Attribute{},
							"bar": schema.StringAttribute{},
						},
					},
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"foo": tftypes.Number,
							"bar": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.Number, 42),
						"bar": tftypes.NewValue(tftypes.String, "bar value"),
					}),
				},
			},
			other:      path.MatchRoot("foo"),
			otherValue: types.Int64Value(42),
			expErrors:  true,
		},
	}

	for name, test := range testCases {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res apstravalidator.ForbiddenWhenValueIsResponse

			apstravalidator.ForbiddenWhenValueIsValidator{
				Expression: test.other,
				Value:      test.otherValue,
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
