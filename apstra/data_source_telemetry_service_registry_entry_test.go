package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	dataSourceTelemetryServiceRegistryByServiceNameHCL = `
data "apstra_telemetry_service_registry_entry" "test" {
  name = "%s"
}
`
)

func TestAccDataSourceTelemetryServiceRegistryEntry(t *testing.T) {
	ctx := context.Background()

	ts := testutils.TelemetryServiceRegistryEntryA(t, ctx)

	TestAppSchema := func(state string) error {
		var diags diag.Diagnostics
		if !utils.JSONEqual(types.StringValue(state), types.StringValue(string(ts.ApplicationSchema)), &diags) {
			return fmt.Errorf("input schema does not match output Input %v. Output %v", ts.ApplicationSchema, state)
		}
		if diags.HasError() {
			var sb strings.Builder
			for _, d := range diags.Errors() {
				sb.WriteString(d.Summary() + "\n")
				sb.WriteString(d.Detail() + "\n")
			}
			return errors.New(sb.String())
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		// PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceTelemetryServiceRegistryByServiceNameHCL, ts.ServiceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_telemetry_service_registry_entry.test", "name", ts.ServiceName),
					resource.TestCheckResourceAttr("data.apstra_telemetry_service_registry_entry.test", "storage_schema_path", rosetta.StringersToFriendlyString(ts.StorageSchemaPath)),
					resource.TestCheckResourceAttrWith("data.apstra_telemetry_service_registry_entry.test", "application_schema", TestAppSchema),
				),
			},
		},
	})
}
