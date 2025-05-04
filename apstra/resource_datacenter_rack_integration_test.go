package tfapstra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceDataCenterRackHCL = `
resource "apstra_datacenter_rack" "test" {
  blueprint_id = %q
  rack_type_id = %q
  name         = %q
}`
)

func TestResourceDatacenterRack(t *testing.T) {
	ctx := context.Background()

	testutils.TestCfgFileToEnv(t)

	bp := testutils.BlueprintC(t, ctx)

	type config struct {
		rackTypeId apstra.ObjectId
		name       string
	}

	renderConfig := func(in config) string {
		return fmt.Sprintf(resourceDataCenterRackHCL,
			bp.Id(),
			in.rackTypeId,
			in.name,
		)
	}

	type step struct {
		config config
		checks []resource.TestCheckFunc
	}

	type testCase struct {
		steps []step
	}

	testCases := map[string]testCase{
		"start_with_name": {
			steps: []step{
				{
					config: config{
						rackTypeId: "access_switch",
						name:       acctest.RandString(5),
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					},
				},
				{
					config: config{
						rackTypeId: "access_switch",
						name:       acctest.RandString(5),
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					},
				},
				{
					config: config{
						rackTypeId: "L2_Virtual",
						name:       acctest.RandString(5),
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "L2_Virtual"),
					},
				},
				{
					config: config{
						rackTypeId: "L2_Virtual",
						name:       acctest.RandString(5),
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "L2_Virtual"),
					},
				},
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase

		steps := make([]resource.TestStep, len(tCase.steps))
		for i, step := range tCase.steps {
			check := resource.ComposeAggregateTestCheckFunc(append(step.checks, resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "name", step.config.name))...)
			steps[i] = resource.TestStep{
				Config: renderConfig(tCase.steps[i].config),
				Check:  check,
			}
		}

		t.Run(tName, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
