package customtypes_test

import (
	"context"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestIPv46AddressTypeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in            tftypes.Value
		expectedDiags diag.Diagnostics
	}{
		"empty-struct": {
			in: tftypes.Value{},
		},
		"null": {
			in: tftypes.NewValue(tftypes.String, nil),
		},
		"unknown": {
			in: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"valid IPv4 address - broadcast": {
			in: tftypes.NewValue(tftypes.String, "255.255.255.255"),
		},
		"valid IPv4 address - loopback": {
			in: tftypes.NewValue(tftypes.String, "127.0.0.1"),
		},
		"valid IPv4 address - multicast": {
			in: tftypes.NewValue(tftypes.String, "224.1.2.3"),
		},
		"valid IPv4 address - zeros": {
			in: tftypes.NewValue(tftypes.String, "0.0.0.0"),
		},
		"valid IPv6 address - unspecified": {
			in: tftypes.NewValue(tftypes.String, "::"),
		},
		"valid IPv6 address - full": {
			in: tftypes.NewValue(tftypes.String, "1:2:3:4:5:6:7:8"),
		},
		"valid IPv6 address - trailing double colon": {
			in: tftypes.NewValue(tftypes.String, "FF01::"),
		},
		"valid IPv6 address - leading double colon": {
			in: tftypes.NewValue(tftypes.String, "::8:800:200C:417A"),
		},
		"valid IPv6 address - middle double colon": {
			in: tftypes.NewValue(tftypes.String, "2001:DB8::8:800:200C:417A"),
		},
		"valid IPv6 address - lowercase": {
			in: tftypes.NewValue(tftypes.String, "2001:db8::8:800:200c:417a"),
		},
		"valid IPv6 address - IPv4-Mapped": {
			in: tftypes.NewValue(tftypes.String, "::FFFF:192.168.255.255"),
		},
		"valid IPv6 address - IPv4-Compatible": {
			in: tftypes.NewValue(tftypes.String, "::127.0.0.1"),
		},
		"invalid IPv6 address - invalid colon end": {
			in: tftypes.NewValue(tftypes.String, "0:0:0:0:0:0:0:"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid IPv46 Address String Value",
					"A string value was provided that is not valid IPv4 or IPv6 string format.\n\n"+
						"Given Value: 0:0:0:0:0:0:0:\n"+
						"Error: ParseAddr(\"0:0:0:0:0:0:0:\"): colon must be followed by more characters (at \":\")",
				),
			},
		},
		"invalid IPv6 address - too many colons": {
			in: tftypes.NewValue(tftypes.String, "0:0::1::"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid IPv46 Address String Value",
					"A string value was provided that is not valid IPv4 or IPv6 string format.\n\n"+
						"Given Value: 0:0::1::\n"+
						"Error: ParseAddr(\"0:0::1::\"): multiple :: in address (at \":\")",
				),
			},
		},
		"invalid IPv6 address - trailing numbers": {
			in: tftypes.NewValue(tftypes.String, "0:0:0:0:0:0:0:1:99"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid IPv46 Address String Value",
					"A string value was provided that is not valid IPv4 or IPv6 string format.\n\n"+
						"Given Value: 0:0:0:0:0:0:0:1:99\n"+
						"Error: ParseAddr(\"0:0:0:0:0:0:0:1:99\"): trailing garbage after address (at \"99\")",
				),
			},
		},
		"wrong-value-type": {
			in: tftypes.NewValue(tftypes.Number, 123),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"IPv46 Address Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected String value, received tftypes.Value with value: tftypes.Number<\"123\">",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := customtypes.IPv46AddressType{}.Validate(context.Background(), testCase.in, path.Root("test"))

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestIPv46AddressTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in          tftypes.Value
		expectation attr.Value
		expectedErr string
	}{
		"true": {
			in:          tftypes.NewValue(tftypes.String, "FF01::101"),
			expectation: customtypes.NewIPv46AddressValue("FF01::101"),
		},
		"unknown": {
			in:          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectation: customtypes.NewIPv46AddressUnknown(),
		},
		"null": {
			in:          tftypes.NewValue(tftypes.String, nil),
			expectation: customtypes.NewIPv46AddressNull(),
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

			got, err := customtypes.IPv46AddressType{}.ValueFromTerraform(ctx, testCase.in)
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
