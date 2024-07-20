package customtypes_test

import (
	"context"
	"net/netip"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"

	"github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestIPv46PrefixStringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentPrefix customtypes.IPv46Prefix
		givenPrefix   basetypes.StringValuable
		expectedMatch bool
		expectedDiags diag.Diagnostics
	}{
		"not equal - IPv6 prefix mismatch": {
			currentPrefix: customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/128"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:1/128"),
			expectedMatch: false,
		},
		"not equal - IPv6 mask longer": {
			currentPrefix: customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/127"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/128"),
			expectedMatch: false,
		},
		"not equal - IPv6 mask shorter": {
			currentPrefix: customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/128"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/127"),
			expectedMatch: false,
		},
		"not equal - IPv6 prefix compressed mismatch": {
			currentPrefix: customtypes.NewIPv46PrefixValue("FF01::/128"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("FF01::1/128"),
			expectedMatch: false,
		},
		"not equal - IPv4-Mapped IPv6 prefix mismatch": {
			currentPrefix: customtypes.NewIPv46PrefixValue("::FFFF:192.168.255.255/128"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("::FFFF:192.168.255.254/128"),
			expectedMatch: false,
		},
		"semantically equal - byte-for-byte match": {
			currentPrefix: customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/0"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/0"),
			expectedMatch: true,
		},
		"semantically equal - case insensitive": {
			currentPrefix: customtypes.NewIPv46PrefixValue("2001:0DB8:0000:0000:0008:0800:200C:417A/128"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("2001:0db8:0000:0000:0008:0800:200c:417a/128"),
			expectedMatch: true,
		},
		"semantically equal - IPv4-Mapped byte-for-byte match": {
			currentPrefix: customtypes.NewIPv46PrefixValue("::FFFF:192.168.0.0/112"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("::FFFF:192.168.0.0/112"),
			expectedMatch: true,
		},
		"semantically equal - compressed all zeroes match": {
			currentPrefix: customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/0"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("::/0"),
			expectedMatch: true,
		},
		"semantically equal - compressed all leading zeroes match": {
			currentPrefix: customtypes.NewIPv46PrefixValue("2001:0DB8:0000:0000:0008:0800:200C:417A/128"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("2001:DB8::8:800:200C:417A/128"),
			expectedMatch: true,
		},
		"semantically equal - start compressed match": {
			currentPrefix: customtypes.NewIPv46PrefixValue("::101/128"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:101/128"),
			expectedMatch: true,
		},
		"semantically equal - middle compressed match": {
			currentPrefix: customtypes.NewIPv46PrefixValue("2001:DB8::8:800:200C:0/112"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("2001:DB8:0:0:8:800:200C:0/112"),
			expectedMatch: true,
		},
		"semantically equal - end compressed match": {
			currentPrefix: customtypes.NewIPv46PrefixValue("FF01:0:0:0:0:0:0:0/64"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("FF01::/64"),
			expectedMatch: true,
		},
		"semantically equal - IPv4-Mapped compressed match": {
			currentPrefix: customtypes.NewIPv46PrefixValue("0:0:0:0:0:FFFF:192.168.255.255/128"),
			givenPrefix:   customtypes.NewIPv46PrefixValue("::FFFF:192.168.255.255/128"),
			expectedMatch: true,
		},
		"error - not given IPv6Prefix IPv6 value": {
			currentPrefix: customtypes.NewIPv46PrefixValue("0:0:0:0:0:0:0:0/0"),
			givenPrefix:   basetypes.NewStringValue("0:0:0:0:0:0:0:0/0"),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected value type was received while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Expected Value Type: customtypes.IPv46Prefix\n"+
						"Got Value Type: basetypes.StringValue",
				),
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			match, diags := testCase.currentPrefix.StringSemanticEquals(context.Background(), testCase.givenPrefix)

			if testCase.expectedMatch != match {
				t.Errorf("Expected StringSemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestIPv46PrefixValueIPv4Prefix(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ipValue          cidrtypes.IPv4Prefix
		expectedIpPrefix netip.Prefix
		expectedDiags    diag.Diagnostics
	}{
		"IPv4 prefix value is null ": {
			ipValue: cidrtypes.NewIPv4PrefixNull(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"IPv4Prefix ValueIPv4Prefix Error",
					"IPv4 CIDR string value is null",
				),
			},
		},
		"IPv4 prefix value is unknown ": {
			ipValue: cidrtypes.NewIPv4PrefixUnknown(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"IPv4Prefix ValueIPv4Prefix Error",
					"IPv4 CIDR string value is unknown",
				),
			},
		},
		"valid IPv4 prefix ": {
			ipValue:          cidrtypes.NewIPv4PrefixValue("192.0.2.0/24"),
			expectedIpPrefix: netip.MustParsePrefix("192.0.2.0/24"),
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ipPrefix, diags := testCase.ipValue.ValueIPv4Prefix()

			if ipPrefix != testCase.expectedIpPrefix {
				t.Errorf("Unexpected difference in netip.Prefix, got: %s, expected: %s", ipPrefix, testCase.expectedIpPrefix)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestIPv6PrefixValueIPv6Prefix(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		prefixValue      cidrtypes.IPv6Prefix
		expectedIpPrefix netip.Prefix
		expectedDiags    diag.Diagnostics
	}{
		"IPv6 prefix value is null ": {
			prefixValue: cidrtypes.NewIPv6PrefixNull(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"IPv6Prefix ValueIPv6Prefix Error",
					"IPv6 CIDR string value is null",
				),
			},
		},
		"IPv6 prefix value is unknown ": {
			prefixValue: cidrtypes.NewIPv6PrefixUnknown(),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"IPv6Prefix ValueIPv6Prefix Error",
					"IPv6 CIDR string value is unknown",
				),
			},
		},
		"valid IPv6 prefix ": {
			prefixValue:      cidrtypes.NewIPv6PrefixValue("2001:DB8:1:2::/64"),
			expectedIpPrefix: netip.MustParsePrefix("2001:DB8:1:2::/64"),
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ipPrefix, diags := testCase.prefixValue.ValueIPv6Prefix()

			if ipPrefix != testCase.expectedIpPrefix {
				t.Errorf("Unexpected difference in netip.Addr, got: %s, expected: %s", ipPrefix, testCase.expectedIpPrefix)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
