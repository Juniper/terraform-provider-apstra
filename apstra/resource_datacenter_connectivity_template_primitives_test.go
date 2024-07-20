//go:build integration

package tfapstra_test

import (
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"math/rand/v2"
	"net"
	"testing"
)

const resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRouteHCL = `
    {
      name            = %s
      routing_zone_id = %q
      network         = %q
      next_hop        = %q
    },`

type resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute struct {
	name          string
	routingZoneId string
	network       net.IPNet
	nextHop       net.IP
}

func (o resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute) render() string {
	return fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRouteHCL,
		stringOrNull(o.name),
		o.routingZoneId,
		o.network.String(),
		o.nextHop.String(),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"routing_zone_id": o.routingZoneId,
		"network":         o.network.String(),
		"next_hop":        o.nextHop.String(),
	}
	if o.name != "" {
		result["name"] = o.name
	}
	return result
}

func randomCustomStaticRoutes(t testing.TB, ipv4Count, ipv6Count int, withLabel bool, routingZoneIds []apstra.ObjectId) []resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute {
	result := make([]resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute, ipv4Count+ipv6Count)

	// add IPv4 routes
	for i := range ipv4Count {
		var name string
		if withLabel {
			name = acctest.RandString(6)
		}
		result[i] = resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute{
			name:          name,
			routingZoneId: routingZoneIds[rand.Int()%len(routingZoneIds)].String(),
			network:       randomSlash31(t, "10.0.0.0/8"),
			nextHop:       randIpvAddressMust(t, "10.0.0.0/8"),
		}
	}

	// add IPv6 routes
	for i := range ipv6Count {
		var name string
		if withLabel {
			name = acctest.RandString(6)
		}
		result[ipv4Count+i] = resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute{
			name:          name,
			routingZoneId: routingZoneIds[rand.Int()%len(routingZoneIds)].String(),
			network:       randomSlash127(t, "2001:db8::/32"),
			nextHop:       randIpvAddressMust(t, "2001:db8::/32"),
		}
	}

	return result
}
