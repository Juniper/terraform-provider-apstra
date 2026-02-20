package testutils

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

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
