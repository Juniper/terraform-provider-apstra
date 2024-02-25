package tfapstra_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

	resourceDataCenterDeviceAllocationSystemAttributesHCL = ` {
		name          = %s
		hostname      = %s
		asn           = %s
		loopback_ipv4 = %s
		loopback_ipv6 = %s
		deploy_mode   = %s
		tags          = %s
    }`
)

func TestResourceDatacenterDeviceAllocation(t *testing.T) {
	ctx := context.Background()

	bpClient := testutils.BlueprintC(t, ctx)

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
		asn          *int
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
			intPtrOrNull(in.asn),
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
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", "not_set"),
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
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
						systemAttributes: &systemAttributes{
							name:     "SPINE1",
							hostname: "spine1.test",
							asn:      utils.ToPtr(1),
							loopbackIpv4: &net.IPNet{
								IP:   net.IP{1, 1, 1, 1},
								Mask: net.CIDRMask(32, 32),
							},
							deployMode: utils.StringersToFriendlyString(apstra.NodeDeployModeReady),
							tags:       []string{"one"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeReady)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "SPINE1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "spine1.test"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeReady)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.#", "1"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "one"),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "spine1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
						systemAttributes: &systemAttributes{
							name:     "SPINE2",
							hostname: "spine2.test",
							asn:      utils.ToPtr(2),
							loopbackIpv4: &net.IPNet{
								IP:   net.IP{2, 2, 2, 2},
								Mask: net.CIDRMask(32, 32),
							},
							deployMode: utils.StringersToFriendlyString(apstra.NodeDeployModeDrain),
							tags:       []string{"two", "2"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "SPINE2"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "spine2.test"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "2"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "2.2.2.2/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "two"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "2"),
					},
				},
			},
		},
		"leaf_start_minimal": {
			steps: []testStep{
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_001_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeNone)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeNone)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_001_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							name:         "leaf_start_minimal_name",
							hostname:     "leafstartminimalhostname.com",
							asn:          utils.ToPtr(1),
							loopbackIpv4: &net.IPNet{IP: net.IP{1, 1, 1, 1}, Mask: net.CIDRMask(32, 32)},
							deployMode:   utils.StringersToFriendlyString(apstra.NodeDeployModeDrain),
							tags:         []string{"one", "1"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "leaf_start_minimal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "leafstartminimalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "one"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "1"),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_001_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "leaf_start_minimal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "leafstartminimalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_001_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							name: "l2_one_access_001_leaf1",
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "l2_one_access_001_leaf1"),
					},
				},
			},
		},
		"leaf_start_maximal": {
			steps: []testStep{
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							name:         "leaf_start_maximal_name",
							hostname:     "leafstartmaximalhostname.com",
							asn:          utils.ToPtr(1),
							loopbackIpv4: &net.IPNet{IP: net.IP{1, 1, 1, 1}, Mask: net.CIDRMask(32, 32)},
							deployMode:   utils.StringersToFriendlyString(apstra.NodeDeployModeDrain),
							tags:         []string{"one", "1"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "leaf_start_maximal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "leafstartmaximalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "one"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "1"),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "leaf_start_maximal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "leafstartmaximalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							name: "l2_one_access_002_leaf1",
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "l2_one_access_002_leaf1"),
					},
				},
			},
		},
		"access_start_minimal": {
			steps: []testStep{
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_001_access1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeNone)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeNone)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_001_access1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
						systemAttributes: &systemAttributes{
							name:       "access_start_minimal_name",
							hostname:   "accessstartminimalhostname.com",
							deployMode: utils.StringersToFriendlyString(apstra.NodeDeployModeDrain),
							tags:       []string{"one", "1"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "access_start_minimal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "accessstartminimalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "one"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "1"),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_001_access1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "access_start_minimal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "accessstartminimalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_001_access1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
						systemAttributes: &systemAttributes{
							name: "l2_one_access_001_access1",
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "l2_one_access_001_access1"),
					},
				},
			},
		},
		"access_start_maximal": {
			steps: []testStep{
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_access1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
						systemAttributes: &systemAttributes{
							name:       "access_start_maximal_name",
							hostname:   "accessstartmaximalhostname.com",
							deployMode: utils.StringersToFriendlyString(apstra.NodeDeployModeDrain),
							tags:       []string{"one", "1"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "access_start_maximal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "accessstartmaximalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "one"),
						resource.TestCheckTypeSetElemAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.tags.*", "1"),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_access1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "access_start_maximal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "accessstartmaximalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_access1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-8x10-1",
						systemAttributes: &systemAttributes{
							name: "l2_one_access_002_access1",
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "l2_one_access_002_access1"),
					},
				},
			},
		},
		"bug_584": {
			steps: []testStep{
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							deployMode: utils.StringersToFriendlyString(apstra.NodeDeployModeNone),
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeNone)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeNone)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							deployMode: utils.StringersToFriendlyString(apstra.NodeDeployModeDrain),
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(apstra.NodeDeployModeDrain)),
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
