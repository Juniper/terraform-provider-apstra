package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"net"
	"testing"
)

const (
	resourceDataCenterDeviceAllocationRefName = "apstra_datacenter_device_allocation.test"
	resourceDataCenterDeviceAllocationHCL     = `
resource "apstra_datacenter_device_allocation" "test" {
	blueprint_id             = %q
	node_name                = %q
	initial_interface_map_id = %q
	system_attributes        = %s
}
`

	resourceDataCenterDeviceAllocationSystemAttributesHCL = `    {
		name          = %s
		hostname      = %s
		asn           = %s
		loopback_ipv4 = %s
		loopback_ipv6 = %s
		deploy_mode   = %s
		tags          = %s
    }
`
)

func TestResourceDatacenterDeviceAllocation(t *testing.T) {
	ctx := context.Background()

	bpClient := testutils.BlueprintA(t, ctx)

	// set spine ASN pool
	err := bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: apstra.ResourceTypeAsnPool,
			Name: apstra.ResourceGroupNameSpineAsn,
		},
		PoolIds: []apstra.ObjectId{"Private-64512-65534"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// set spine loopback ipv4 pool
	err = bpClient.SetResourceAllocation(ctx, &apstra.ResourceGroupAllocation{
		ResourceGroup: apstra.ResourceGroup{
			Type: apstra.ResourceTypeIp4Pool,
			Name: apstra.ResourceGroupNameSpineIp4,
		},
		PoolIds: []apstra.ObjectId{"Private-10_0_0_0-8"},
	})
	if err != nil {
		t.Fatal(err)
	}

	type systemAttributes struct {
		name         string
		hostname     string
		asn          int
		loopbackIpv4 *net.IPNet
		loopbackIpv6 *net.IPNet
		deployMode   string
		tags         []string
	}

	renderSystemAttributes := func(in *systemAttributes) string {
		if in == nil {
			return "null"
		}

		return fmt.Sprintf(resourceDataCenterDeviceAllocationSystemAttributesHCL,
			stringOrNull(in.name),
			stringOrNull(in.hostname),
			intPtrOrNull(&in.asn),
			cidrOrNull(in.loopbackIpv4),
			cidrOrNull(in.loopbackIpv6),
			stringOrNull(in.deployMode),
			stringSetOrNull(in.tags),
		)
	}

	type deviceAllocation struct {
		blueprintId           string
		nodeName              string
		initialInterfaceMapId string
		systemAttributes      *systemAttributes
	}

	renderDeviceAllocation := func(in deviceAllocation) string {
		return fmt.Sprintf(resourceDataCenterDeviceAllocationHCL,
			in.blueprintId,
			in.nodeName,
			in.initialInterfaceMapId,
			renderSystemAttributes(in.systemAttributes),
		)
	}

	type testStep struct {
		config deviceAllocation
		checks []resource.TestCheckFunc
	}

	type testCase struct {
		steps []testStep
	}

	testCases := map[string]testCase{
		"spine1": {
			steps: []testStep{
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "spine1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Spine",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Spine"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", "not_set"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "spine1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "spine1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4"),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "spine1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Spine",
						systemAttributes: &systemAttributes{
							name:     "SPINE1",
							hostname: "spine1.test",
							asn:      1,
							loopbackIpv4: &net.IPNet{
								IP:   net.IP{1, 1, 1, 1},
								Mask: net.CIDRMask(32, 32),
							},
							loopbackIpv6: nil,
							deployMode:   "ready",
							tags:         []string{"one"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Spine"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "SPINE1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "spine1.test"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", "ready"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.#", "1"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "one"),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "spine1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Spine",
						systemAttributes: &systemAttributes{
							name:     "SPINE2",
							hostname: "spine2.test",
							asn:      2,
							loopbackIpv4: &net.IPNet{
								IP:   net.IP{2, 2, 2, 2},
								Mask: net.CIDRMask(32, 32),
							},
							loopbackIpv6: nil,
							deployMode:   "drain",
							tags:         []string{"two", "2"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Spine"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "SPINE2"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "spine2.test"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "2"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "2.2.2.2/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", "drain"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "two"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "2"),
					},
				},
			},
		},
	}

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + renderDeviceAllocation(step.config),
					Check:  resource.ComposeAggregateTestCheckFunc(step.checks...),
				}
			}
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
