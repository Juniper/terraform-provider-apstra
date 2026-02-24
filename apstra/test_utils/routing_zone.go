package testutils

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// SecurityZoneA creates a minimally configured EVPN security zone
func SecurityZoneA(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient, cleanup bool) string {
	t.Helper()

	name := acctest.RandString(10)
	id, err := client.CreateSecurityZone(ctx, apstra.SecurityZone{
		Label:           name,
		Type:            enum.SecurityZoneTypeEVPN,
		VRFName:         name,
		RoutingPolicyID: "",
		RouteTarget:     nil,
		RTPolicy:        nil,
		VLAN:            nil,
		VNI:             nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cleanup {
		t.Cleanup(func() {
			err := client.DeleteSecurityZone(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

	return id
}

// SecurityZoneB creates a minimally configured EVPN security zone. If Apstra version >= 6.1.0, IPv6 will be enabled.
func SecurityZoneB(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient, cleanup bool) string {
	t.Helper()

	name := acctest.RandString(10)
	var as *enum.AddressingScheme
	if compatibility.BPDefaultRoutingZoneAddressingOK.Check(version.Must(version.NewVersion(client.Client().ApiVersion()))) {
		as = &enum.AddressingSchemeIPv6
	}
	id, err := client.CreateSecurityZone(ctx, apstra.SecurityZone{
		Label:             name,
		Type:              enum.SecurityZoneTypeEVPN,
		VRFName:           name,
		RoutingPolicyID:   "",
		RouteTarget:       nil,
		RTPolicy:          nil,
		VLAN:              nil,
		VNI:               nil,
		AddressingSupport: as,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cleanup {
		t.Cleanup(func() {
			err := client.DeleteSecurityZone(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

	return id
}
