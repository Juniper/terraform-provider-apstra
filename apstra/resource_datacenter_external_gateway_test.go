package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"net"
	"strconv"
	"strings"
	"testing"
)

const (
	resourceDataCenterExternalGateway = `
resource "apstra_datacenter_external_gateway" "test" {
  blueprint_id        = "%s"
  name                = "%s"
  ip_address          = "%s"
  asn                 = %d
  evpn_route_types    = "%s"
  local_gateway_nodes = ["%s"]
  ttl                 = %s
  keepalive_time      = %s
  hold_time           = %s
  password            = %s
}
`
)

type testCaseResourceExternalGateway struct {
	name          string
	ipAddress     net.IP
	asn           uint32
	routeTypes    apstra.RemoteGatewayRouteTypes
	nodes         string
	ttl           *uint8
	keepaliveTime *uint16
	holdTime      *uint16
	password      string
	testCheckFunc resource.TestCheckFunc
}

func renderResourceDataCenterExternalGateway(tc testCaseResourceExternalGateway, bp *apstra.TwoStageL3ClosClient) string {
	return fmt.Sprintf(resourceDataCenterExternalGateway,
		bp.Id(),
		tc.name,
		tc.ipAddress,
		tc.asn,
		tc.routeTypes.Value,
		tc.nodes,
		intPtrOrNull(tc.ttl),
		intPtrOrNull(tc.keepaliveTime),
		intPtrOrNull(tc.holdTime),
		stringOrNull(tc.password),
	)
}

func TestResourceDatacenterExternalGateway(t *testing.T) {
	ctx := context.Background()

	bp := testutils.BlueprintC(t, ctx)

	leafIds := systemIds(ctx, t, bp, "leaf")
	uint8Val3 := uint8(3)
	uint16Val1 := uint16(1)
	uint16Val3 := uint16(3)

	testCases := []testCaseResourceExternalGateway{
		{
			name:       "name1",
			ipAddress:  net.IP{1, 1, 1, 1},
			asn:        1,
			routeTypes: apstra.RemoteGatewayRouteTypesAll,
			nodes:      leafIds[0],
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_external_gateway.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "name", "name1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ip_address", "1.1.1.1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "asn", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "evpn_route_types", apstra.RemoteGatewayRouteTypesAll.Value),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.0", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ttl", "30"),            // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "keepalive_time", "10"), // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "hold_time", "30"),      // default
			}...),
		},
		{
			name:       "name2",
			ipAddress:  net.IP{1, 1, 1, 2},
			asn:        2,
			routeTypes: apstra.RemoteGatewayRouteTypesFiveOnly,
			nodes:      strings.Join(leafIds[1:], `","`),
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_external_gateway.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "name", "name2"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ip_address", "1.1.1.2"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "asn", "2"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "evpn_route_types", apstra.RemoteGatewayRouteTypesFiveOnly.Value),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.#", strconv.Itoa(len(leafIds)-1)),
				resource.TestCheckTypeSetElemAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.*", leafIds[1]),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ttl", "30"),            // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "keepalive_time", "10"), // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "hold_time", "30"),      // default
			}...),
		},
		{
			name:          "name3",
			ipAddress:     net.IP{1, 1, 1, 3},
			asn:           3,
			routeTypes:    apstra.RemoteGatewayRouteTypesAll,
			nodes:         leafIds[0],
			ttl:           &uint8Val3,
			keepaliveTime: &uint16Val1,
			holdTime:      &uint16Val3,
			password:      "big secret1",
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_external_gateway.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "name", "name3"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ip_address", "1.1.1.3"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "asn", "3"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "evpn_route_types", apstra.RemoteGatewayRouteTypesAll.Value),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.0", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ttl", "3"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "keepalive_time", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "hold_time", "3"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "password", "big secret1"),
			}...),
		},
		{
			name:       "name1",
			ipAddress:  net.IP{1, 1, 1, 1},
			asn:        1,
			routeTypes: apstra.RemoteGatewayRouteTypesAll,
			nodes:      leafIds[0],
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_external_gateway.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "name", "name1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ip_address", "1.1.1.1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "asn", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "evpn_route_types", apstra.RemoteGatewayRouteTypesAll.Value),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.#", "1"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.0", leafIds[0]),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ttl", "30"),            // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "keepalive_time", "10"), // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "hold_time", "30"),      // default
			}...),
		},
		{
			name:       "name2",
			ipAddress:  net.IP{1, 1, 1, 2},
			asn:        2,
			routeTypes: apstra.RemoteGatewayRouteTypesFiveOnly,
			nodes:      strings.Join(leafIds[1:], `","`),
			password:   "big secret2",
			testCheckFunc: resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet("apstra_datacenter_external_gateway.test", "id"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "blueprint_id", bp.Id().String()),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "name", "name2"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ip_address", "1.1.1.2"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "asn", "2"),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "evpn_route_types", apstra.RemoteGatewayRouteTypesFiveOnly.Value),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.#", strconv.Itoa(len(leafIds)-1)),
				resource.TestCheckTypeSetElemAttr("apstra_datacenter_external_gateway.test", "local_gateway_nodes.*", leafIds[1]),
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "ttl", "30"),            // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "keepalive_time", "10"), // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "hold_time", "30"),      // default
				resource.TestCheckResourceAttr("apstra_datacenter_external_gateway.test", "password", "big secret2"),
			}...),
		},
	}

	steps := make([]resource.TestStep, len(testCases))
	for i, tc := range testCases {
		steps[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + renderResourceDataCenterExternalGateway(tc, bp),
			Check:  tc.testCheckFunc,
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}
