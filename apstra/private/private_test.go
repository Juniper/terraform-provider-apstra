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

type testPS map[string][]byte // testPS mocks the private.State interface for testing purposes.

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
		setterLoader private.SetterLoader // SeterLoader is the type which writes and reads itself in and out of private state.
		expectedJSON string               // expected JSON representation of the SetterLoader

		// Expected keys in private state after the SetterLoader writes itself. Note that these keys are
		// hard coded here to discourage maintainers from inadvertently changing them by renaming structs.
		// Changing keys needs to be done with care because those keys are persisted into the Terraform
		// state. Careless changes will orphan previously stored state.
		expectedKeys []string

		// A type-specific instance of the generic RoundTrip() function which works on the SetterLoader type in this test case.
		roundTrip func(context.Context, private.State, *diag.Diagnostics)
	}

	testCases := map[string]testCase{
		"EphemeralApiToken": { // Uses a legacy key in release 0.x.y
			setterLoader: &private.EphemeralApiToken{
				Token:         "token-value",
				ExpiresAt:     time.Date(2029, time.January, 20, 17, 0, 0, 0, time.UTC),
				WarnThreshold: time.Minute * 10,
				DoNotLogOut:   true,
			},
			expectedJSON: `{"token":"token-value","expires_at":"2029-01-20T17:00:00Z","warn_threshold":600000000000,"do_not_log_out":true}`,
			expectedKeys: []string{"private.EphemeralApiToken", "EphemeralApiToken"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.EphemeralApiToken](ctx, ps, diags)
			},
		},
		"ResourceDatacenterIpLinkAddressingInterfaceAddressing": { // Uses a legacy key in release 0.x.y
			setterLoader: &private.ResourceDatacenterIpLinkAddressingInterfaceAddressing{
				SwitchIpv4:  enum.InterfaceNumberingIpv4TypeNumbered,
				SwitchIpv6:  enum.InterfaceNumberingIpv6TypeLinkLocal,
				GenericIpv4: enum.InterfaceNumberingIpv4TypeNumbered,
				GenericIpv6: enum.InterfaceNumberingIpv6TypeLinkLocal,
			},
			expectedJSON: `{"generic_ipv4":"numbered", "generic_ipv6":"link_local", "switch_ipv4":"numbered", "switch_ipv6":"link_local"}`,
			expectedKeys: []string{"private.ResourceDatacenterIpLinkAddressingInterfaceAddressing", "ResourceDatacenterIpLinkAddressingInterfaceAddressing"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterIpLinkAddressingInterfaceAddressing](ctx, ps, diags)
			},
		},
		"ResourceDatacenterIpLinkAddressingInterfaceIds": { // Uses a legacy key in release 0.x.y
			setterLoader: &private.ResourceDatacenterIpLinkAddressingInterfaceIds{
				SwitchInterface:  "switch-if-id",
				GenericInterface: "generic-if-id",
			},
			expectedJSON: `{"switch_interface":"switch-if-id","generic_interface":"generic-if-id"}`,
			expectedKeys: []string{"private.ResourceDatacenterIpLinkAddressingInterfaceIds", "ResourceDatacenterIpLinkAddressingInterfaceIds"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterIpLinkAddressingInterfaceIds](ctx, ps, diags)
			},
		},
		"ResourceDatacenterRoutingZoneLoopbackAddresses": {
			setterLoader: &private.ResourceDatacenterRoutingZoneLoopbackAddresses{
				"loopback-id-one": {
					HasIpv4: true,
					HasIpv6: false,
				},
				"loopback-id-two": {
					HasIpv4: false,
					HasIpv6: true,
				},
			},
			expectedJSON: `{"loopback-id-one":{"has_ipv4":true,"has_ipv6":false},"loopback-id-two":{"has_ipv4":false,"has_ipv6":true}}`,
			expectedKeys: []string{"private.ResourceDatacenterRoutingZoneLoopbackAddresses"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterRoutingZoneLoopbackAddresses](ctx, ps, diags)
			},
		},
		"ResourceDatacenterSwitchingZone": {
			setterLoader: &private.ResourceDatacenterSwitchingZone{
				HasRouteTarget: true,
			},
			expectedJSON: `{"has_route_target":true}`,
			expectedKeys: []string{"private.ResourceDatacenterSwitchingZone"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterSwitchingZone](ctx, ps, diags)
			},
		},
		"ResourceDatacenterTag": {
			setterLoader: &private.ResourceDatacenterTag{
				Id: "test-id",
			},
			expectedJSON: `{"id":"test-id"}`,
			expectedKeys: []string{"private.ResourceDatacenterTag"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceDatacenterTag](ctx, ps, diags)
			},
		},
		"ResourceFreeformAggregateLink": {
			setterLoader: &private.ResourceFreeformAggregateLink{
				SysIDToLAGID: map[string]string{"key-id": "value-id"},
			},
			expectedJSON: `{"SysIDToLAGID":{"key-id":"value-id"}}`,
			expectedKeys: []string{"private.ResourceFreeformAggregateLink"},
			roundTrip: func(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
				private.RoundTrip[private.ResourceFreeformAggregateLink](ctx, ps, diags)
			},
		},
	}

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			// Ensure test case is constructed correctly and derive the preferred key string.
			require.NotNil(t, tCase.setterLoader)
			preferredKey := reflect.TypeOf(tCase.setterLoader).Elem().String()
			b, err := json.Marshal(tCase.setterLoader)
			require.NoError(t, err)
			require.JSONEq(t, tCase.expectedJSON, string(b))
			require.Contains(t, tCase.expectedKeys, preferredKey)
			require.NotNil(t, tCase.roundTrip)

			// Set the struct in private state.
			ps := make(testPS)
			var diags diag.Diagnostics
			tCase.setterLoader.SetPrivateState(context.Background(), ps, &diags)
			require.False(t, diags.HasError())

			// Private state should contain one or two keys (two if the struct is migrating to a new key)
			require.Equal(t, len(tCase.expectedKeys), len(ps))

			// Ensure the stored bytes are exactly the expected JSON.
			for _, expectedKey := range tCase.expectedKeys {
				require.JSONEq(t, tCase.expectedJSON, string(ps[expectedKey]))
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

			// Private state should contain one or two keys (two if the struct is migrating to a new key)
			require.Equal(t, len(tCase.expectedKeys), len(ps))

			// Ensure the stored bytes are exactly the expected JSON.
			for _, expectedKey := range tCase.expectedKeys {
				require.JSONEq(t, tCase.expectedJSON, string(ps[expectedKey]))
			}

			// Recover the struct from private state into a new target.
			target := reflect.New(reflect.TypeOf(tCase.setterLoader).Elem()).Interface().(private.SetterLoader)
			target.LoadPrivateState(context.Background(), ps, &diags)
			require.False(t, diags.HasError())

			// Ensure the recovered struct matches the original.
			require.Equal(t, tCase.setterLoader, target)

			// Delete the private state data using a zero-value of the struct type.
			zero := reflect.New(reflect.TypeOf(tCase.setterLoader).Elem()).Interface().(private.SetterLoader)
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
