//go:build integration

package dctestobj

import (
	"context"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// SwitchingZoneA creates a minimally configured Switching Zone
func SwitchingZoneA(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient, cleanup bool) string {
	t.Helper()

	name := acctest.RandString(10)
	id, err := client.CreateSwitchingZone(ctx, datacenter.SwitchingZone{
		Label:             &name,
		MACVRFName:        &name,
		MACVRFServiceType: random.OneOf(&enum.SwitchingZoneMACVRFServiceTypeVLANBundle, &enum.SwitchingZoneMACVRFServiceTypeVLANAware),
	})
	if err != nil {
		t.Fatal(err)
	}

	if cleanup {
		t.Cleanup(func() {
			err := client.DeleteSwitchingZone(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

	return id
}
