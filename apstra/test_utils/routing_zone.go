package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func SecurityZoneA(ctx context.Context, client *apstra.TwoStageL3ClosClient) (apstra.ObjectId, func(context.Context) error, error) {
	deleteFunc := func(_ context.Context) error { return nil }
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
		return "", deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeleteSecurityZone(ctx, id)
	}

	return id, deleteFunc, nil
}
