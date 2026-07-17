package customtypes_test

import (
	"context"
	"testing"

	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRouteTargetTypeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in            basetypes.StringValue
		expectedDiags diag.Diagnostics
	}{
		"null": {
			in: types.StringNull(),
		},
		"unknown": {
			in: types.StringUnknown(),
		},
		"valid zeros": {
			in: types.StringValue("0:0"),
		},
		"valid ones": {
			in: types.StringValue("0:0"),
		},
		"valid max16 max32": {
			in: types.StringValue("65535:4294967295"),
		},
		"valid max32 max16": {
			in: types.StringValue("4294967295:65535"),
		},
		"valid IPv4 address": {
			in: types.StringValue("127.0.0.1:127"),
		},
		"invalid 33:16": {
			in: types.StringValue("4294967296:65535"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Route Target String Value",
					"A string value was provided that is not a valid Route Type string format.\n\n"+
						"Given Value: 4294967296:65535\n"+
						"Error: cannot parse 1st part of route target \"4294967296:65535\"\n",
				),
			},
		},
		"invalid 32:17": {
			in: types.StringValue("4294967295:65536"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Route Target String Value",
					"A string value was provided that is not a valid Route Type string format.\n\n"+
						"Given Value: 4294967295:65536\n"+
						"Error: parsing 2nd part of route target \"4294967295:65536\": strconv.ParseUint: parsing \"65536\": value out of range\n",
				),
			},
		},
		"invalid 16:33": {
			in: types.StringValue("65535:4294967296"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Route Target String Value",
					"A string value was provided that is not a valid Route Type string format.\n\n"+
						"Given Value: 65535:4294967296\n"+
						"Error: parsing 2nd part of route target \"65535:4294967296\": strconv.ParseUint: parsing \"4294967296\": value out of range\n",
				),
			},
		},
		"invalid 17:32": {
			in: types.StringValue("65536:4294967295"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Route Target String Value",
					"A string value was provided that is not a valid Route Type string format.\n\n"+
						"Given Value: 65536:4294967295\n"+
						"Error: parsing 2nd part of route target \"65536:4294967295\": strconv.ParseUint: parsing \"4294967295\": value out of range\n",
				),
			},
		},
		"invalid IPv4 address": {
			in: types.StringValue("1.2.3.4.5:1"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Route Target String Value",
					"A string value was provided that is not a valid Route Type string format.\n\n"+
						"Given Value: 1.2.3.4.5:1\n"+
						"Error: cannot parse 1st part of route target \"1.2.3.4.5:1\"\n",
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := xattr.ValidateAttributeRequest{Path: path.Root("test")}
			var response xattr.ValidateAttributeResponse
			customtypes.RouteTarget{StringValue: testCase.in}.ValidateAttribute(context.Background(), request, &response)
			if diff := cmp.Diff(response.Diagnostics, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestRouteTargetTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in          tftypes.Value
		expectation attr.Value
		expectedErr string
	}{
		"true": {
			in:          tftypes.NewValue(tftypes.String, "1:1"),
			expectation: customtypes.NewRouteTargetValue("1:1"),
		},
		"unknown": {
			in:          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectation: customtypes.NewRouteTargetUnknown(),
		},
		"null": {
			in:          tftypes.NewValue(tftypes.String, nil),
			expectation: customtypes.NewRouteTargetNull(),
		},
		"wrongType": {
			in:          tftypes.NewValue(tftypes.Number, 123),
			expectedErr: "can't unmarshal tftypes.Number into *string, expected string",
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := customtypes.RouteTargetType{}.ValueFromTerraform(ctx, testCase.in)
			if err != nil {
				if testCase.expectedErr == "" {
					t.Fatalf("Unexpected error: %s", err)
				}
				if testCase.expectedErr != err.Error() {
					t.Fatalf("Expected error to be %q, got %q", testCase.expectedErr, err.Error())
				}
				return
			}
			if err == nil && testCase.expectedErr != "" {
				t.Fatalf("Expected error to be %q, didn't get an error", testCase.expectedErr)
			}
			if !got.Equal(testCase.expectation) {
				t.Errorf("Expected %+v, got %+v", testCase.expectation, got)
			}
			if testCase.expectation.IsNull() != testCase.in.IsNull() {
				t.Errorf("Expected null-ness match: expected %t, got %t", testCase.expectation.IsNull(), testCase.in.IsNull())
			}
			if testCase.expectation.IsUnknown() != !testCase.in.IsKnown() {
				t.Errorf("Expected unknown-ness match: expected %t, got %t", testCase.expectation.IsUnknown(), !testCase.in.IsKnown())
			}
		})
	}
}
