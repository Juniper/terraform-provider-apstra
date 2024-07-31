//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

const resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRouteHCL = `{
  name            = %s
  routing_zone_id = %q
  network         = %q
  next_hop        = %q
},
`

type resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute struct {
	name          string
	routingZoneId string
	network       net.IPNet
	nextHop       net.IP
}

func (o resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRouteHCL,
			stringOrNull(o.name),
			o.routingZoneId,
			o.network.String(),
			o.nextHop.String(),
		),
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

func randomCustomStaticRoutes(t testing.TB, ctx context.Context, ipv4Count, ipv6Count int, withLabel bool, client *apstra.TwoStageL3ClosClient) []resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute, ipv4Count+ipv6Count)

	rzName := acctest.RandString(6)
	rzId, err := client.CreateSecurityZone(ctx, &apstra.SecurityZoneData{
		Label:   rzName,
		SzType:  apstra.SecurityZoneTypeEVPN,
		VrfName: rzName,
	})
	require.NoError(t, err)

	// add IPv4 routes
	for i := range ipv4Count {
		var name string
		if withLabel {
			name = acctest.RandString(6)
		}
		result[i] = resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute{
			name:          name,
			routingZoneId: rzId.String(),
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
			routingZoneId: rzId.String(),
			network:       randomSlash127(t, "2001:db8::/32"),
			nextHop:       randIpvAddressMust(t, "2001:db8::/32"),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicyHCL = `{
  name              = %s
  routing_policy_id = %q
},
`

type resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy struct {
	name            string
	routingPolicyId string
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicyHCL,
			stringOrNull(o.name),
			o.routingPolicyId,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy) valueAsMapForChecks() map[string]string {
	result := make(map[string]string)
	if o.name != "" {
		result["name"] = o.name
	}
	return result
}

func randomRoutingPolicies(t testing.TB, ctx context.Context, count int, withLabel bool, client *apstra.TwoStageL3ClosClient) []resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy, count)
	for i := range result {
		policyId, err := client.CreateRoutingPolicy(ctx, &apstra.DcRoutingPolicyData{
			Label:        acctest.RandString(6),
			PolicyType:   apstra.DcRoutingPolicyTypeUser,
			ImportPolicy: apstra.DcRoutingPolicyImportPolicyAll,
		})
		require.NoError(t, err)

		var name string
		if withLabel {
			name = acctest.RandString(6)
		}

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy{
			name:            name,
			routingPolicyId: policyId.String(),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpPrimitiveHCL = `{
  name             = %s
  neighbor_asn     = %s
  ttl              = %s
  bfd_enabled      = %q
  password         = %s
  keepalive_time   = %s
  hold_time        = %s
  local_asn        = %s
  ipv4_address     = %s
  ipv6_address     = %s
  routing_policies = %s
},
`

type resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpPrimitive struct {
	name            string
	neighborAsn     *int
	ttl             *int
	bfdEnabled      bool
	password        string
	keepaliveTime   *int
	holdTime        *int
	localAsn        *int
	ipv4Address     net.IP
	ipv6Address     net.IP
	routingPolicies []resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy
}

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpPrimitive) render(indent int) string {
	routingPolicies := "null"

	if len(o.routingPolicies) > 0 {
		sb := new(strings.Builder)
		for _, routingPolicy := range o.routingPolicies {
			sb.WriteString(routingPolicy.render(indent))
		}

		routingPolicies = "[\n" + sb.String() + "  ]"
	}

	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpPrimitiveHCL,
			stringOrNull(o.name),
			intPtrOrNull(o.neighborAsn),
			intPtrOrNull(o.ttl),
			strconv.FormatBool(o.bfdEnabled),
			stringOrNull(o.password),
			intPtrOrNull(o.keepaliveTime),
			intPtrOrNull(o.holdTime),
			intPtrOrNull(o.localAsn),
			ipOrNull(o.ipv4Address),
			ipOrNull(o.ipv6Address),
			routingPolicies,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpPrimitive) valueAsMapForChecks() map[string]string {
	result := make(map[string]string)
	if o.name != "" {
		result["name"] = o.name
	}
	if o.neighborAsn != nil {
		result["neighbor_asn"] = strconv.Itoa(*o.neighborAsn)
	}
	if o.ttl != nil {
		result["ttl"] = strconv.Itoa(*o.ttl)
	}
	result["bfd_enabled"] = strconv.FormatBool(o.bfdEnabled)
	if o.password != "" {
		result["password"] = o.password
	}
	if o.keepaliveTime != nil {
		result["keepalive_time"] = strconv.Itoa(*o.keepaliveTime)
	}
	if o.holdTime != nil {
		result["hold_time"] = strconv.Itoa(*o.holdTime)
	}
	if o.localAsn != nil {
		result["local_asn"] = strconv.Itoa(*o.localAsn)
	}
	if o.ipv4Address.String() != "<nil>" {
		result["ipv4_address"] = o.ipv4Address.String()
	}
	if o.ipv6Address.String() != "<nil>" {
		result["ipv6_address"] = o.ipv6Address.String()
	}

	// todo: --------------- add routing policies to map ... somehow?

	return result
}

func randomBgpPeeringIpPrimitives(t testing.TB, ctx context.Context, count int, withLabel bool, client *apstra.TwoStageL3ClosClient) []resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpPrimitive {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpPrimitive, count)
	for i := range result {
		var name string
		if withLabel {
			name = acctest.RandString(6)
		}
		var holdTime, keepaliveTime *int
		if rand.Int()%2 == 0 {
			keepaliveTime = utils.ToPtr(rand.IntN(constants.KeepaliveTimeMax-constants.KeepaliveTimeMin) + constants.KeepaliveTimeMin)
			holdMin := *keepaliveTime * 3
			holdTime = utils.ToPtr(rand.IntN(constants.HoldTimeMax-holdMin) + holdMin)
		}

		var ipv4Address, ipv6Address net.IP
		if rand.Int()%2 == 0 {
			ipv4Address = randIpvAddressMust(t, "192.0.2.0/24")
		} else {
			ipv6Address = randIpvAddressMust(t, "2001:db8::/32")
		}

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpPrimitive{
			name:            name,
			neighborAsn:     oneOf(utils.ToPtr(rand.IntN(constants.AsnMax+constants.AsnMin)), (*int)(nil)),
			ttl:             utils.ToPtr(rand.IntN(constants.TtlMax) + constants.TtlMin), // always send TTL so whole object isn't null
			bfdEnabled:      oneOf(true, false),
			password:        oneOf(acctest.RandString(6), ""),
			keepaliveTime:   keepaliveTime,
			holdTime:        holdTime,
			localAsn:        oneOf(utils.ToPtr(rand.IntN(constants.AsnMax+constants.AsnMin)), (*int)(nil)),
			ipv4Address:     ipv4Address,
			ipv6Address:     ipv6Address,
			routingPolicies: randomRoutingPolicies(t, ctx, rand.IntN(count), withLabel, client),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringPrimitiveHCL = `{
  name             = %s
  ttl              = %s
  bfd_enabled      = %q
  password         = %s
  keepalive_time   = %s
  hold_time        = %s
  ipv4_enabled     = %q
  ipv6_enabled     = %q
  local_asn        = %s
  ipv4_peer_prefix = %s
  ipv6_peer_prefix = %s
  routing_policies = %s
},
`

type resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringPrimitive struct {
	name            string
	ttl             *int
	bfdEnabled      bool
	password        string
	keepaliveTime   *int
	holdTime        *int
	ipv4Enabled     bool
	ipv6Enabled     bool
	localAsn        *int
	ipv4PeerPrefix  net.IPNet
	ipv6PeerPrefix  net.IPNet
	routingPolicies []resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy
}

func (o resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringPrimitive) render(indent int) string {
	routingPolicies := "null"

	if len(o.routingPolicies) > 0 {
		sb := new(strings.Builder)
		for _, routingPolicy := range o.routingPolicies {
			sb.WriteString(routingPolicy.render(indent))
		}

		routingPolicies = "[\n" + sb.String() + "  ]"
	}

	return tfapstra.Indent(indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringPrimitiveHCL,
			stringOrNull(o.name),
			intPtrOrNull(o.ttl),
			strconv.FormatBool(o.bfdEnabled),
			stringOrNull(o.password),
			intPtrOrNull(o.keepaliveTime),
			intPtrOrNull(o.holdTime),
			strconv.FormatBool(o.ipv4Enabled),
			strconv.FormatBool(o.ipv6Enabled),
			intPtrOrNull(o.localAsn),
			ipNetOrNull(o.ipv4PeerPrefix),
			ipNetOrNull(o.ipv6PeerPrefix),
			routingPolicies,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringPrimitive) valueAsMapForChecks() map[string]string {
	result := make(map[string]string)
	if o.name != "" {
		result["name"] = o.name
	}
	if o.ttl != nil {
		result["ttl"] = strconv.Itoa(*o.ttl)
	}
	result["bfd_enabled"] = strconv.FormatBool(o.bfdEnabled)
	if o.password != "" {
		result["password"] = o.password
	}
	if o.keepaliveTime != nil {
		result["keepalive_time"] = strconv.Itoa(*o.keepaliveTime)
	}
	if o.holdTime != nil {
		result["hold_time"] = strconv.Itoa(*o.holdTime)
	}
	result["ipv4_enabled"] = strconv.FormatBool(o.ipv4Enabled)
	result["ipv6_enabled"] = strconv.FormatBool(o.ipv6Enabled)
	if o.localAsn != nil {
		result["local_asn"] = strconv.Itoa(*o.localAsn)
	}
	if o.ipv4PeerPrefix.String() != "<nil>" {
		result["ipv4_peer_prefix"] = o.ipv4PeerPrefix.String()
	}
	if o.ipv6PeerPrefix.String() != "<nil>" {
		result["ipv6_peer_prefix"] = o.ipv6PeerPrefix.String()
	}

	// todo: --------------- add routing policies to map ... somehow?

	return result
}

func randomDynamicBgpPeeringPrimitives(t testing.TB, ctx context.Context, count int, withLabel bool, client *apstra.TwoStageL3ClosClient) []resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringPrimitive {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringPrimitive, count)
	for i := range result {
		var name string
		if withLabel {
			name = acctest.RandString(6)
		}
		var holdTime, keepaliveTime *int
		if rand.Int()%2 == 0 {
			keepaliveTime = utils.ToPtr(rand.IntN(constants.KeepaliveTimeMax-constants.KeepaliveTimeMin) + constants.KeepaliveTimeMin)
			holdMin := *keepaliveTime * 3
			holdTime = utils.ToPtr(rand.IntN(constants.HoldTimeMax-holdMin) + holdMin)
		}

		var ipv4Enabled, ipv6Enabled bool
		switch rand.IntN(3) {
		case 0:
			ipv4Enabled = true
		case 1:
			ipv6Enabled = true
		case 2:
			ipv4Enabled = true
			ipv6Enabled = true
		}

		var ipv4PeerPrefix, ipv6PeerPrefix net.IPNet
		if ipv4Enabled && (rand.Int()%2) == 0 {
			ipv4PeerPrefix = randomPrefix(t, "192.0.2.0/24", 27)
		}
		if ipv6Enabled && ipv4PeerPrefix.IP != nil {
			ipv6PeerPrefix = randomPrefix(t, "3fff::/20", 64)
		}

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringPrimitive{
			name:            name,
			ttl:             utils.ToPtr(rand.IntN(constants.TtlMax) + constants.TtlMin), // always send TTL so whole object isn't null
			bfdEnabled:      oneOf(true, false),
			password:        oneOf(acctest.RandString(6), ""),
			keepaliveTime:   keepaliveTime,
			holdTime:        holdTime,
			ipv4Enabled:     ipv4Enabled,
			ipv6Enabled:     ipv6Enabled,
			localAsn:        oneOf(utils.ToPtr(rand.IntN(constants.AsnMax+constants.AsnMin)), (*int)(nil)),
			ipv4PeerPrefix:  ipv4PeerPrefix,
			ipv6PeerPrefix:  ipv6PeerPrefix,
			routingPolicies: randomRoutingPolicies(t, ctx, rand.IntN(count), withLabel, client),
		}
	}

	return result
}
