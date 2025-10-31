//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceDataCenterDeviceAllocationRefName = "apstra_datacenter_device_allocation.test"
	resourceDataCenterDeviceAllocationHCL     = `
resource "apstra_datacenter_device_allocation" "test" {
	blueprint_id             = %q
	node_name                = %q
	initial_interface_map_id = %q
    deploy_mode              = %s
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

	unspecifiedDeployModeByVersion := func(bpClient *apstra.TwoStageL3ClosClient) string {
		t.Helper()

		apiVersion, err := version.NewVersion(bpClient.Client().ApiVersion())
		require.NoError(t, err)

		unspecifiedAsNone := version.MustConstraints(version.NewConstraint("<6.0.0"))
		unspecifiedAsUndeploy := version.MustConstraints(version.NewConstraint(">=6.0.0"))

		switch {
		case unspecifiedAsNone.Check(apiVersion):
			t.Logf("case unspecifiedAsNone: returning %s for version %s", utils.StringersToFriendlyString(enum.DeployModeNone), apiVersion)
			return utils.StringersToFriendlyString(enum.DeployModeNone)
		case unspecifiedAsUndeploy.Check(apiVersion):
			t.Logf("case unspecifiedAsUndeploy: returning %s for version %s", utils.StringersToFriendlyString(enum.DeployModeNone), apiVersion)
			return utils.StringersToFriendlyString(enum.DeployModeUndeploy)
		default:
			panic("unable to determine behavior for unspecified deploy mode")
		}
	}

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
			stringSliceOrNull(in.tags),
		)
	}

	type deviceAllocation struct {
		blueprintId           string
		nodeName              string
		initialInterfaceMapId string
		systemAttributes      *systemAttributes
		deployMode            string
	}

	renderDeviceAllocation := func(in deviceAllocation) string {
		return fmt.Sprintf(resourceDataCenterDeviceAllocationHCL,
			in.blueprintId,
			in.nodeName,
			in.initialInterfaceMapId,
			stringOrNull(in.deployMode),
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
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", unspecifiedDeployModeByVersion(bpClient)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", unspecifiedDeployModeByVersion(bpClient)),
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
							asn:      pointer.To(1),
							loopbackIpv4: &net.IPNet{
								IP:   net.IP{1, 1, 1, 1},
								Mask: net.CIDRMask(32, 32),
							},
							deployMode: utils.StringersToFriendlyString(enum.DeployModeReady),
							tags:       []string{"one"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeReady)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "SPINE1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "spine1.test"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeReady)),
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
							asn:      pointer.To(2),
							loopbackIpv4: &net.IPNet{
								IP:   net.IP{2, 2, 2, 2},
								Mask: net.CIDRMask(32, 32),
							},
							deployMode: utils.StringersToFriendlyString(enum.DeployModeDrain),
							tags:       []string{"two", "2"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "SPINE2"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "spine2.test"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "2"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "2.2.2.2/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", unspecifiedDeployModeByVersion(bpClient)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", unspecifiedDeployModeByVersion(bpClient)),
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
							asn:          pointer.To(1),
							loopbackIpv4: &net.IPNet{IP: net.IP{1, 1, 1, 1}, Mask: net.CIDRMask(32, 32)},
							deployMode:   utils.StringersToFriendlyString(enum.DeployModeDrain),
							tags:         []string{"one", "1"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "leaf_start_minimal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "leafstartminimalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "leaf_start_minimal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "leafstartminimalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
							asn:          pointer.To(1),
							loopbackIpv4: &net.IPNet{IP: net.IP{1, 1, 1, 1}, Mask: net.CIDRMask(32, 32)},
							deployMode:   utils.StringersToFriendlyString(enum.DeployModeDrain),
							tags:         []string{"one", "1"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-7x10-Leaf"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "leaf_start_maximal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "leafstartmaximalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "leaf_start_maximal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "leafstartmaximalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.asn", "1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.loopback_ipv4", "1.1.1.1/32"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", unspecifiedDeployModeByVersion(bpClient)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", unspecifiedDeployModeByVersion(bpClient)),
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
							deployMode: utils.StringersToFriendlyString(enum.DeployModeDrain),
							tags:       []string{"one", "1"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "access_start_minimal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "accessstartminimalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "access_start_minimal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "accessstartminimalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
							deployMode: utils.StringersToFriendlyString(enum.DeployModeDrain),
							tags:       []string{"one", "1"},
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "initial_interface_map_id", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "interface_map_name", "Juniper_vQFX__AOS-8x10-1"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "device_profile_node_id"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "access_start_maximal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "accessstartmaximalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.name", "access_start_maximal_name"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.hostname", "accessstartmaximalhostname.com"),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
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
							deployMode: utils.StringersToFriendlyString(enum.DeployModeNone),
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeNone)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeNone)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_002_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							deployMode: utils.StringersToFriendlyString(enum.DeployModeDrain),
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeDrain)),
					},
				},
			},
		},
		"bug_609_deprecated_option": {
			steps: []testStep{
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_one_access_003_leaf1",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						deployMode:            "drain",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(resourceDataCenterDeviceAllocationRefName, "node_id"),
					},
				},
			},
		},
		"deploy_to_not_set": {
			steps: []testStep{
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_esi_acs_dual_001_leaf2",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							deployMode: utils.StringersToFriendlyString(enum.DeployModeReady),
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeReady)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeReady)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_esi_acs_dual_001_leaf2",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							deployMode: utils.StringersToFriendlyString(enum.DeployModeNone),
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeNone)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeNone)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_esi_acs_dual_001_leaf2",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
						systemAttributes: &systemAttributes{
							deployMode: utils.StringersToFriendlyString(enum.DeployModeReady),
						},
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeReady)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeReady)),
					},
				},
				{
					config: deviceAllocation{
						blueprintId:           bpClient.Id().String(),
						nodeName:              "l2_esi_acs_dual_001_leaf2",
						initialInterfaceMapId: "Juniper_vQFX__AOS-7x10-Leaf",
					},
					checks: []resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "blueprint_id", bpClient.Id().String()),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "deploy_mode", utils.StringersToFriendlyString(enum.DeployModeReady)),
						resource.TestCheckResourceAttr(resourceDataCenterDeviceAllocationRefName, "system_attributes.deploy_mode", utils.StringersToFriendlyString(enum.DeployModeReady)),
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
