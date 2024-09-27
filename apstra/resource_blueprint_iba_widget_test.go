package tfapstra

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceBlueprintIbaWidgetHCL = `
resource "apstra_blueprint_iba_probe" "p_device_health" {
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
resource "apstra_blueprint_iba_widget" "w_device_health_high_cpu" {
  blueprint_id = "%s"
  name = "%s"
  probe_id = apstra_blueprint_iba_probe.p_device_health.id
  stage = "Systems with high CPU utilization"
  description = "made from terraform"
}
`
)

func TestAccResourceWidget(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)
	clientVersion := version.Must(version.NewVersion(client.ApiVersion()))
	if !compatibility.BpIbaWidgetOk.Check(clientVersion) {
		t.Skipf("skipping due to version constraint %s", compatibility.BpIbaWidgetOk)
	}

	bpClient := testutils.MakeOrFindBlueprint(t, ctx, "widget", testutils.BlueprintA)

	n1 := "Widget1"
	n2 := "Widget2"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceBlueprintIbaWidgetHCL, bpClient.Id(), bpClient.Id(), n1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_blueprint_iba_widget.w_device_health_high_cpu", "id"),
					// Check the name
					resource.TestCheckResourceAttr("apstra_blueprint_iba_widget.w_device_health_high_cpu", "name", n1),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(resourceBlueprintIbaWidgetHCL, bpClient.Id(), bpClient.Id(), n2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_blueprint_iba_widget.w_device_health_high_cpu", "id"),
					// Check the name
					resource.TestCheckResourceAttr("apstra_blueprint_iba_widget.w_device_health_high_cpu", "name", n2),
				),
			},
		},
	})
}
