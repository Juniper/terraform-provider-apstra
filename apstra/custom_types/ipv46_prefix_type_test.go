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

func TestIPv46PrefixTypeValidate(t *testing.T) {
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
		"valid IPv4 prefix - all zeros": {
			in: types.StringValue("0.0.0.0/0"),
		},
		"valid IPv4 prefix - all ones": {
			in: types.StringValue("255.255.255.255/32"),
		},
		"valid IPv4 prefix - loopback block": {
			in: types.StringValue("127.0.0.0/8"),
		},
		"valid IPv4 prefix - multicast block": {
			in: types.StringValue("224.0.0.0/4"),
		},
		"valid IPv4 prefix - 10slash8": {
			in: types.StringValue("10.0.0.0/8"),
		},
		"valid IPv4 prefix - 10slash24": {
			in: types.StringValue("10.1.2.0/24"),
		},
		"valid IPv6 address - unspecified": {
			in: types.StringValue("::/0"),
		},
		"valid IPv6 host prefix": {
			in: types.StringValue("1:2:3:4:5:6:7:8/128"),
		},
		"valid IPv6 prefix - trailing double colon": {
			in: types.StringValue("FF01::/64"),
		},
		"valid IPv6 prefix - leading double colon": {
			in: types.StringValue("::8:0:0:0:0/64"),
		},
		"valid IPv6 host prefix - middle double colon": {
			in: types.StringValue("2001:DB8::8:800:200C:417A/128"),
		},
		"valid IPv6 host prefix - lowercase": {
			in: types.StringValue("2001:db8::8:800:200c:417a/128"),
		},
		"valid IPv6 prefix - IPv4-Mapped": {
			in: types.StringValue("::FFFF:192.168.0.0/112"),
		},
		"valid IPv6 prefix - IPv4-Compatible": {
			in: types.StringValue("::127.0.0.1/128"),
		},
		"invalid IPv4 prefix - not base address": {
			in: types.StringValue("192.168.1.1/24"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid IPv46 Prefix String Value",
					"A string value was provided that does not represent a base address of an IPv4 or IPv6 prefix.\n\n"+
						"Given Value: 192.168.1.1/24\n"+
						"Base Address: 192.168.1.0/24",
				),
			},
		},
		"invalid IPv6 prefix - not base address": {
			in: types.StringValue("2001:db8::1/64"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid IPv46 Prefix String Value",
					"A string value was provided that does not represent a base address of an IPv4 or IPv6 prefix.\n\n"+
						"Given Value: 2001:db8::1/64\n"+
						"Base Address: 2001:db8::/64",
				),
			},
		},
		"invalid IPv6 prefix - invalid colon end": {
			in: types.StringValue("0:0:0:0:0:0:0:/0"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid IPv46 Prefix String Value",
					"A string value was provided that is not valid IPv4 or IPv6 prefix string format.\n\n"+
						"Given Value: 0:0:0:0:0:0:0:/0\n"+
						"Error: netip.ParsePrefix(\"0:0:0:0:0:0:0:/0\"): ParseAddr(\"0:0:0:0:0:0:0:\"): colon must be followed by more characters (at \":\")",
				),
			},
		},
		"invalid IPv6 prefix - too many colons": {
			in: types.StringValue("0:0::1::/64"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid IPv46 Prefix String Value",
					"A string value was provided that is not valid IPv4 or IPv6 prefix string format.\n\n"+
						"Given Value: 0:0::1::/64\n"+
						"Error: netip.ParsePrefix(\"0:0::1::/64\"): ParseAddr(\"0:0::1::\"): multiple :: in address (at \":\")",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := xattr.ValidateAttributeRequest{Path: path.Root("test")}
			var response xattr.ValidateAttributeResponse
			customtypes.IPv46Prefix{StringValue: testCase.in}.ValidateAttribute(context.Background(), request, &response)
			if diff := cmp.Diff(response.Diagnostics, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestIPv46PrefixTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in          tftypes.Value
		expectation attr.Value
		expectedErr string
	}{
		"true": {
			in:          tftypes.NewValue(tftypes.String, "FF01::101/128"),
			expectation: customtypes.NewIPv46PrefixValue("FF01::101/128"),
		},
		"unknown": {
			in:          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectation: customtypes.NewIPv46PrefixUnknown(),
		},
		"null": {
			in:          tftypes.NewValue(tftypes.String, nil),
			expectation: customtypes.NewIPv46PrefixNull(),
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

			got, err := customtypes.IPv46PrefixType{}.ValueFromTerraform(ctx, testCase.in)
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
