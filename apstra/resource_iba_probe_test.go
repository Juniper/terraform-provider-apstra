package tfapstra

import (
	"context"
	"errors"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	resourceIbaProbeHCL = `
resource "apstra_iba_probe" "p_device_health" {
  blueprint_id = "%s"
  predefined_probe_id = "device_health"
  probe_config = jsonencode(
    {
      "max_cpu_utilization": 80,
      "max_memory_utilization": 80,
      "max_disk_utilization": 80,
      "duration": 660,
      "threshold_duration": 360,
      "history_duration": 604800
    }
  )
}
`
)

func TestAccResourceProbe(t *testing.T) {
	ctx := context.Background()
	bpClient, bpDelete, err := testutils.MakeOrFindBlueprint(ctx, "1bb57", testutils.BlueprintA)
	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}
	defer func() {
		err = bpDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceIbaProbeHCL, bpClient.Id()),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_iba_probe.p_device_health", "id"), resource.TestCheckResourceAttr("apstra_iba_probe.p_device_health", "name", "Device System Health")),
			},
		},
	})
}
