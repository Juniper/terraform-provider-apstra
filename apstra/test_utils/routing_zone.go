package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"testing"
)

func SecurityZoneA(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient) apstra.ObjectId {
	t.Helper()

	name := acctest.RandString(10)
	id, err := client.CreateSecurityZone(ctx, &apstra.SecurityZoneData{
		Label:           name,
		SzType:          apstra.SecurityZoneTypeEVPN,
		VrfName:         name,
		RoutingPolicyId: "",
		RouteTarget:     nil,
		RtPolicy:        nil,
		VlanId:          nil,
		VniId:           nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := client.DeleteSecurityZone(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
	})

	return id
}
