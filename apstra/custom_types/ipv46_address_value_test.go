package customtypes_test

import (
	"context"
	"net/netip"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestIPv46AddressStringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentIpAddr customtypes.IPv46Address
		givenIpAddr   basetypes.StringValuable
		expectedMatch bool
		expectedDiags diag.Diagnostics
	}{
		"not equal - IPv6 address mismatch": {
			currentIpAddr: customtypes.NewIPv46AddressValue("0:0:0:0:0:0:0:0"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("0:0:0:0:0:0:0:1"),
			expectedMatch: false,
		},
		"not equal - IPv6 address compressed mismatch": {
			currentIpAddr: customtypes.NewIPv46AddressValue("FF01::"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("FF01::1"),
			expectedMatch: false,
		},
		"not equal - IPv4-Mapped IPv6 address mismatch": {
			currentIpAddr: customtypes.NewIPv46AddressValue("::FFFF:192.168.255.255"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("::FFFF:192.168.255.254"),
			expectedMatch: false,
		},
		"semantically equal - byte-for-byte match": {
			currentIpAddr: customtypes.NewIPv46AddressValue("0:0:0:0:0:0:0:0"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("0:0:0:0:0:0:0:0"),
			expectedMatch: true,
		},
		"semantically equal - case insensitive": {
			currentIpAddr: customtypes.NewIPv46AddressValue("2001:0DB8:0000:0000:0008:0800:200C:417A"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("2001:0db8:0000:0000:0008:0800:200c:417a"),
			expectedMatch: true,
		},
		"semantically equal - IPv4-Mapped byte-for-byte match": {
			currentIpAddr: customtypes.NewIPv46AddressValue("::FFFF:192.168.255.255"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("::FFFF:192.168.255.255"),
			expectedMatch: true,
		},
		"semantically equal - compressed all zeroes match": {
			currentIpAddr: customtypes.NewIPv46AddressValue("0:0:0:0:0:0:0:0"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("::"),
			expectedMatch: true,
		},
		"semantically equal - compressed all leading zeroes match": {
			currentIpAddr: customtypes.NewIPv46AddressValue("2001:0DB8:0000:0000:0008:0800:200C:417A"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("2001:DB8::8:800:200C:417A"),
			expectedMatch: true,
		},
		"semantically equal - start compressed match": {
			currentIpAddr: customtypes.NewIPv46AddressValue("::101"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("0:0:0:0:0:0:0:101"),
			expectedMatch: true,
		},
		"semantically equal - middle compressed match": {
			currentIpAddr: customtypes.NewIPv46AddressValue("2001:DB8::8:800:200C:417A"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("2001:DB8:0:0:8:800:200C:417A"),
			expectedMatch: true,
		},
		"semantically equal - end compressed match": {
			currentIpAddr: customtypes.NewIPv46AddressValue("FF01:0:0:0:0:0:0:0"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("FF01::"),
			expectedMatch: true,
		},
		"semantically equal - IPv4-Mapped compressed match": {
			currentIpAddr: customtypes.NewIPv46AddressValue("0:0:0:0:0:FFFF:192.168.255.255"),
			givenIpAddr:   customtypes.NewIPv46AddressValue("::FFFF:192.168.255.255"),
			expectedMatch: true,
		},
		"error - not given IPv6Address IPv6 value": {
			currentIpAddr: customtypes.NewIPv46AddressValue("0:0:0:0:0:0:0:0"),
			givenIpAddr:   basetypes.NewStringValue("0:0:0:0:0:0:0:0"),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected value type was received while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Expected Value Type: customtypes.IPv46Address\n"+
						"Got Value Type: basetypes.StringValue",
				),
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			match, diags := testCase.currentIpAddr.StringSemanticEquals(context.Background(), testCase.givenIpAddr)

			if testCase.expectedMatch != match {
				t.Errorf("Expected StringSemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestIPv46AddressValueIPv4Address(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ipValue        iptypes.IPv4Address
		expectedIpAddr netip.Addr
		expectedDiags  diag.Diagnostics
	}{
		"IPv4 address value is null ": {
			ipValue: iptypes.NewIPv4AddressNull(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"IPv4Address ValueIPv4Address Error",
					"IPv4 address string value is null",
				),
			},
		},
		"IPv4 address value is unknown ": {
			ipValue: iptypes.NewIPv4AddressUnknown(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"IPv4Address ValueIPv4Address Error",
					"IPv4 address string value is unknown",
				),
			},
		},
		"valid IPv4 address ": {
			ipValue:        iptypes.NewIPv4AddressValue("192.0.2.1"),
			expectedIpAddr: netip.MustParseAddr("192.0.2.1"),
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ipAddr, diags := testCase.ipValue.ValueIPv4Address()

			if ipAddr != testCase.expectedIpAddr {
				t.Errorf("Unexpected difference in netip.Addr, got: %s, expected: %s", ipAddr, testCase.expectedIpAddr)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestIPv6AddressValueIPv6Address(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ipValue        iptypes.IPv6Address
		expectedIpAddr netip.Addr
		expectedDiags  diag.Diagnostics
	}{
		"IPv6 address value is null ": {
			ipValue: iptypes.NewIPv6AddressNull(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"IPv6Address ValueIPv6Address Error",
					"IPv6 address string value is null",
				),
			},
		},
		"IPv6 address value is unknown ": {
			ipValue: iptypes.NewIPv6AddressUnknown(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"IPv6Address ValueIPv6Address Error",
					"IPv6 address string value is unknown",
				),
			},
		},
		"valid IPv6 address ": {
			ipValue:        iptypes.NewIPv6AddressValue("2001:DB8::8:800:200C:417A"),
			expectedIpAddr: netip.MustParseAddr("2001:DB8::8:800:200C:417A"),
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ipAddr, diags := testCase.ipValue.ValueIPv6Address()

			if ipAddr != testCase.expectedIpAddr {
				t.Errorf("Unexpected difference in netip.Addr, got: %s, expected: %s", ipAddr, testCase.expectedIpAddr)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
