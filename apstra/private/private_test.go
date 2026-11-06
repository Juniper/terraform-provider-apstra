package private_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

type testPS map[string][]byte

func (ps testPS) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if result, ok := ps[key]; ok {
		return result, nil
	}

	return nil, nil
}

func (ps testPS) SetKey(_ context.Context, key string, b []byte) diag.Diagnostics {
	if len(b) == 0 {
		delete(ps, key)
		return nil
	}

	ps[key] = b
	return nil
}

func TestPrivateStateSetLoad(t *testing.T) {
	type testCase struct {
		sl           private.SetterLoader
		slJSON       string
		expectedKeys []string
		roundTrip    func(context.Context, private.State, *diag.Diagnostics)
	}

	testCases := map[string]testCase{
		"EphemeralApiToken": {
			sl: &private.EphemeralApiToken{
				Token:         "token-value",
				ExpiresAt:     time.Date(2029, time.January, 20, 17, 0, 0, 0, time.UTC),
				WarnThreshold: time.Minute * 10,
				DoNotLogOut:   true,
			},
			slJSON:       `{"token":"token-value","expires_at":"2029-01-20T17:00:00Z","warn_threshold":600000000000,"do_not_log_out":true}`,
			expectedKeys: []string{"private.EphemeralApiToken", "EphemeralApiToken"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.EphemeralApiToken](ctx, ps, diags)
			},
		},
		"ResourceDatacenterIpLinkAddressingInterfaceAddressing": {
			sl: &private.ResourceDatacenterIpLinkAddressingInterfaceAddressing{
				SwitchIpv4:  enum.InterfaceNumberingIpv4TypeNumbered,
				SwitchIpv6:  enum.InterfaceNumberingIpv6TypeLinkLocal,
				GenericIpv4: enum.InterfaceNumberingIpv4TypeNumbered,
				GenericIpv6: enum.InterfaceNumberingIpv6TypeLinkLocal,
			},
			slJSON:       `{"generic_ipv4":"numbered", "generic_ipv6":"link_local", "switch_ipv4":"numbered", "switch_ipv6":"link_local"}`,
			expectedKeys: []string{"private.ResourceDatacenterIpLinkAddressingInterfaceAddressing", "ResourceDatacenterIpLinkAddressingInterfaceAddressing"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterIpLinkAddressingInterfaceAddressing](ctx, ps, diags)
			},
		},
		"ResourceDatacenterIpLinkAddressingInterfaceIds": {
			sl: &private.ResourceDatacenterIpLinkAddressingInterfaceIds{
				SwitchInterface:  "switch-if-id",
				GenericInterface: "generic-if-id",
			},
			slJSON:       `{"switch_interface":"switch-if-id","generic_interface":"generic-if-id"}`,
			expectedKeys: []string{"private.ResourceDatacenterIpLinkAddressingInterfaceIds", "ResourceDatacenterIpLinkAddressingInterfaceIds"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterIpLinkAddressingInterfaceIds](ctx, ps, diags)
			},
		},
		"ResourceDatacenterRoutingZoneLoopbackAddresses": {
			sl: &private.ResourceDatacenterRoutingZoneLoopbackAddresses{
				"loopback-id-one": {
					HasIpv4: true,
					HasIpv6: false,
				},
				"loopback-id-two": {
					HasIpv4: false,
					HasIpv6: true,
				},
			},
			slJSON:       `{"loopback-id-one":{"has_ipv4":true,"has_ipv6":false},"loopback-id-two":{"has_ipv4":false,"has_ipv6":true}}`,
			expectedKeys: []string{"private.ResourceDatacenterRoutingZoneLoopbackAddresses"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterRoutingZoneLoopbackAddresses](ctx, ps, diags)
			},
		},
		"ResourceDatacenterSwitchingZone": {
			sl: &private.ResourceDatacenterSwitchingZone{
				HasRouteTarget: true,
			},
			slJSON:       `{"has_route_target":true}`,
			expectedKeys: []string{"private.ResourceDatacenterSwitchingZone"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterSwitchingZone](ctx, ps, diags)
			},
		},
		"ResourceDatacenterTag": {
			sl: &private.ResourceDatacenterTag{
				Id: "test-id",
			},
			slJSON:       `{"id":"test-id"}`,
			expectedKeys: []string{"private.ResourceDatacenterTag"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterTag](ctx, ps, diags)
			},
		},
		"ResourceFreeformAggregateLink": {
			sl: &private.ResourceFreeformAggregateLink{
				SysIDToLAGID: map[string]string{"key-id": "value-id"},
			},
			slJSON:       `{"SysIDToLAGID":{"key-id":"value-id"}}`,
			expectedKeys: []string{"private.ResourceFreeformAggregateLink"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceFreeformAggregateLink](ctx, ps, diags)
			},
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			preferredKey := reflect.TypeOf(tCase.sl).Elem().String()

			// ensure test case is constructed correctly
			b, err := json.Marshal(tCase.sl)
			require.NoError(t, err)
			require.JSONEq(t, tCase.slJSON, string(b))
			require.Contains(t, tCase.expectedKeys, preferredKey)
			require.NotNil(t, tCase.roundTrip)

			// set the struct in private state
			ps := make(testPS)
			var diags diag.Diagnostics
			tCase.sl.SetPrivateState(context.Background(), ps, &diags)
			require.False(t, diags.HasError())

			// private state should contain one or two keys (two if the struct is migrating to a new key)
			require.Equal(t, len(tCase.expectedKeys), len(ps))

			// ensure the stored bytes are exactly the expected JSON
			for _, expectedKey := range tCase.expectedKeys {
				require.JSONEq(t, tCase.slJSON, string(ps[expectedKey]))
			}

			// Test key migration by deleting the preferred key for types which have
			// a legacy key to simulate a private state which needs to be upgraded.
			if len(ps) > 1 {
				// Delete the preferred key to simulate migration from a legacy key to the preferred key.
				delete(ps, preferredKey)
			}

			// Round trip a sample struct from private state and back to ensure it writes under all of the expected keys.
			tCase.roundTrip(context.Background(), ps, &diags)
			require.False(t, diags.HasError())

			// private state should contain one or two keys (two if the struct is migrating to a new key)
			require.Equal(t, len(tCase.expectedKeys), len(ps))

			// ensure the stored bytes are exactly the expected JSON
			for _, expectedKey := range tCase.expectedKeys {
				require.JSONEq(t, tCase.slJSON, string(ps[expectedKey]))
			}

			// recover the struct from private state into a new target
			target := reflect.New(reflect.TypeOf(tCase.sl).Elem()).Interface().(private.SetterLoader)
			target.LoadPrivateState(context.Background(), ps, &diags)
			require.False(t, diags.HasError())

			// ensure the recovered struct matches the original
			require.Equal(t, tCase.sl, target)

			// Delete the private state data using a zero-value of the struct type.
			zero := reflect.New(reflect.TypeOf(tCase.sl).Elem()).Interface().(private.SetterLoader)
			zero.SetPrivateState(context.Background(), ps, &diags)
			require.False(t, diags.HasError())
			require.Empty(t, ps)

			// Round trip with empty state to ensure that state remains empty.
			tCase.roundTrip(context.Background(), ps, &diags)
			require.False(t, diags.HasError())
			require.Empty(t, ps)
		})
	}
}
