package tfapstra_test

import (
	"context"
	"errors"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	resourceDataCenterRack = `
resource "apstra_datacenter_rack" "test" {
  blueprint_id = %q
  rack_type_id = %q
  rack_name    = %s
}`
)

type stringList struct {
	list []string
}

func (o *stringList) append(s string) string {
	o.list = append(o.list, s)
	return s
}
func (o stringList) last() string {
	return o.list[len(o.list)-1]
}

func TestResourceDatacenterRack(t *testing.T) {
	ctx := context.Background()

	bp, bpDelete, err := testutils.BlueprintC(ctx)
	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}
	defer func() {
		err = bpDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	type step struct {
		config string
		check  resource.TestCheckFunc
	}

	type testCase struct {
		steps []step
	}

	var names stringList

	testCases := map[string]testCase{
		"start_with_name": {
			steps: []step{
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", names.append(fmt.Sprintf("%q", acctest.RandString(5)))),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", names.append(fmt.Sprintf("%q", acctest.RandString(5)))),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", names.last()),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", "null"),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", names.append(fmt.Sprintf("%q", acctest.RandString(5)))),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
			},
		},
		"start_without_name": {
			steps: []step{
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", "null"),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", names.append(fmt.Sprintf("%q", acctest.RandString(5)))),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", names.append(fmt.Sprintf("%q", acctest.RandString(5)))),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
				{
					config: insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterRack, bp.Id(), "access_switch", names.last()),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("apstra_datacenter_rack.test", "id"),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "blueprint_id", bp.Id().String()),
						resource.TestCheckResourceAttr("apstra_datacenter_rack.test", "rack_type_id", "access_switch"),
					),
				},
			},
		},
	}

	for _, tc := range testCases {
		steps := make([]resource.TestStep, len(tc.steps))
		for i := 0; i < len(tc.steps); i++ {
			steps[i] = resource.TestStep{
				Config: tc.steps[i].config,
				Check:  tc.steps[i].check,
			}
		}

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps:                    steps,
		})
	}
}
