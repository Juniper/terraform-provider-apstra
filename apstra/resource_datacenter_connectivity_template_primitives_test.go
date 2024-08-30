//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
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
  name            = %q
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
			o.name,
			o.routingZoneId,
			o.network.String(),
			o.nextHop.String(),
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name":            o.name,
		"routing_zone_id": o.routingZoneId,
		"network":         o.network.String(),
		"next_hop":        o.nextHop.String(),
	}

	return result
}

func randomCustomStaticRoutes(t testing.TB, ctx context.Context, ipv4Count, ipv6Count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute {
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
		result[i] = resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute{
			name:          acctest.RandString(6),
			routingZoneId: rzId.String(),
			network:       randomSlash31(t, "10.0.0.0/8"),
			nextHop:       randIpvAddressMust(t, "10.0.0.0/8"),
		}
	}

	// add IPv6 routes
	for i := range ipv6Count {
		result[ipv4Count+i] = resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute{
			name:          acctest.RandString(6),
			routingZoneId: rzId.String(),
			network:       randomSlash127(t, "2001:db8::/32"),
			nextHop:       randIpvAddressMust(t, "2001:db8::/32"),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicyHCL = `{
  name              = %q
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
			o.name,
			o.routingPolicyId,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name": o.name,
	}

	return result
}

func randomRoutingPolicies(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy, count)
	for i := range result {
		policyId, err := client.CreateRoutingPolicy(ctx, &apstra.DcRoutingPolicyData{
			Label:        acctest.RandString(6),
			PolicyType:   apstra.DcRoutingPolicyTypeUser,
			ImportPolicy: oneOf(apstra.DcRoutingPolicyImportPolicyAll, apstra.DcRoutingPolicyImportPolicyDefaultOnly, apstra.DcRoutingPolicyImportPolicyExtraOnly),
		})
		require.NoError(t, err)

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy{
			name:            acctest.RandString(6),
			routingPolicyId: policyId.String(),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpointHCL = `{
  name             = %q
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

type resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint struct {
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

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint) render(indent int) string {
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
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpointHCL,
			o.name,
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

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name": o.name,
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

func randomBgpPeeringIpEndpointPrimitives(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint, count)
	for i := range result {
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

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint{
			name:            acctest.RandString(6),
			neighborAsn:     oneOf(utils.ToPtr(rand.IntN(constants.AsnMax+constants.AsnMin)), (*int)(nil)),
			ttl:             utils.ToPtr(rand.IntN(constants.TtlMax-constants.TtlMin) + constants.TtlMin),
			bfdEnabled:      oneOf(true, false),
			password:        oneOf(acctest.RandString(6), ""),
			keepaliveTime:   keepaliveTime,
			holdTime:        holdTime,
			localAsn:        oneOf(utils.ToPtr(rand.IntN(constants.AsnMax+constants.AsnMin)), (*int)(nil)),
			ipv4Address:     ipv4Address,
			ipv6Address:     ipv6Address,
			routingPolicies: randomRoutingPolicies(t, ctx, rand.IntN(count), client, cleanup),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringHCL = `{
  name             = %q
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

type resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering struct {
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

func (o resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering) render(indent int) string {
	routingPolicies := "null"
	if len(o.routingPolicies) > 0 {
		sb := new(strings.Builder)
		for _, routingPolicy := range o.routingPolicies {
			sb.WriteString(routingPolicy.render(indent))
		}

		routingPolicies = "[\n" + sb.String() + "  ]"
	}

	return tfapstra.Indent(indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringHCL,
			o.name,
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

func (o resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name":         o.name,
		"bfd_enabled":  strconv.FormatBool(o.bfdEnabled),
		"ipv4_enabled": strconv.FormatBool(o.ipv4Enabled),
		"ipv6_enabled": strconv.FormatBool(o.ipv6Enabled),
	}
	if o.ttl != nil {
		result["ttl"] = strconv.Itoa(*o.ttl)
	}
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
	if o.ipv4PeerPrefix.String() != "<nil>" {
		result["ipv4_peer_prefix"] = o.ipv4PeerPrefix.String()
	}
	if o.ipv6PeerPrefix.String() != "<nil>" {
		result["ipv6_peer_prefix"] = o.ipv6PeerPrefix.String()
	}

	// todo: --------------- add routing policies to map ... somehow?

	return result
}

func randomDynamicBgpPeeringPrimitives(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering, count)
	for i := range result {
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

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering{
			name:            acctest.RandString(6),
			ttl:             utils.ToPtr(rand.IntN(constants.TtlMax-constants.TtlMin) + constants.TtlMin),
			bfdEnabled:      oneOf(true, false),
			password:        oneOf(acctest.RandString(6), ""),
			keepaliveTime:   keepaliveTime,
			holdTime:        holdTime,
			ipv4Enabled:     ipv4Enabled,
			ipv6Enabled:     ipv6Enabled,
			localAsn:        oneOf(utils.ToPtr(rand.IntN(constants.AsnMax+constants.AsnMin)), (*int)(nil)),
			ipv4PeerPrefix:  ipv4PeerPrefix,
			ipv6PeerPrefix:  ipv6PeerPrefix,
			routingPolicies: randomRoutingPolicies(t, ctx, rand.IntN(count), client, cleanup),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystemHCL = `{
  name                 = %q
  ttl                  = %s
  bfd_enabled          = %q
  password             = %s
  keepalive_time       = %s
  hold_time            = %s
  ipv4_addressing_type = %s
  ipv6_addressing_type = %s
  local_asn            = %s
  neighbor_asn_dynamic = %s
  peer_from_loopback   = %s
  peer_to              = %q
  routing_policies     = %s
},
`

type resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem struct {
	name               string
	ttl                *int
	bfdEnabled         bool
	password           string
	keepaliveTime      *int
	holdTime           *int
	ipv4Addressing     apstra.CtPrimitiveIPv4ProtocolSessionAddressing
	ipv6Addressing     apstra.CtPrimitiveIPv6ProtocolSessionAddressing
	localAsn           *int
	neighborAsnDynamic bool
	peerFromLoopback   bool
	peerTo             apstra.CtPrimitiveBgpPeerTo
	routingPolicies    []resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy
}

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem) render(indent int) string {
	routingPolicies := "null"
	if len(o.routingPolicies) > 0 {
		sb := new(strings.Builder)
		for _, routingPolicy := range o.routingPolicies {
			sb.WriteString(routingPolicy.render(indent))
		}

		routingPolicies = "[\n" + sb.String() + "  ]"
	}

	return tfapstra.Indent(indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystemHCL,
			o.name,
			intPtrOrNull(o.ttl),
			strconv.FormatBool(o.bfdEnabled),
			stringOrNull(o.password),
			intPtrOrNull(o.keepaliveTime),
			intPtrOrNull(o.holdTime),
			stringOrNull(o.ipv4Addressing.String()),
			stringOrNull(o.ipv6Addressing.String()),
			intPtrOrNull(o.localAsn),
			strconv.FormatBool(o.neighborAsnDynamic),
			strconv.FormatBool(o.peerFromLoopback),
			utils.StringersToFriendlyString(o.peerTo),
			routingPolicies,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name":                 o.name,
		"bfd_enabled":          strconv.FormatBool(o.bfdEnabled),
		"neighbor_asn_dynamic": strconv.FormatBool(o.neighborAsnDynamic),
		"peer_from_loopback":   strconv.FormatBool(o.peerFromLoopback),
		"peer_to":              utils.StringersToFriendlyString(o.peerTo),
	}
	if o.ttl != nil {
		result["ttl"] = strconv.Itoa(*o.ttl)
	}
	if o.password != "" {
		result["password"] = o.password
	}
	if o.keepaliveTime != nil {
		result["keepalive_time"] = strconv.Itoa(*o.keepaliveTime)
	}
	if o.holdTime != nil {
		result["hold_time"] = strconv.Itoa(*o.holdTime)
	}
	if o.ipv4Addressing.String() != "" {
		result["ipv4_addressing"] = o.ipv4Addressing.String()
	}
	if o.ipv6Addressing.String() != "" {
		result["ipv6_addressing"] = o.ipv6Addressing.String()
	}
	if o.localAsn != nil {
		result["local_asn"] = strconv.Itoa(*o.localAsn)
	}

	// todo: --------------- add routing policies to map ... somehow?

	return result
}

func randomBgpPeeringGenericSystemPrimitives(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem, count)
	for i := range result {
		var holdTime, keepaliveTime *int
		if rand.Int()%2 == 0 {
			keepaliveTime = utils.ToPtr(rand.IntN(constants.KeepaliveTimeMax-constants.KeepaliveTimeMin) + constants.KeepaliveTimeMin)
			holdMin := *keepaliveTime * 3
			holdTime = utils.ToPtr(rand.IntN(constants.HoldTimeMax-holdMin) + holdMin)
		}

		var ipv4Addressing apstra.CtPrimitiveIPv4ProtocolSessionAddressing
		var ipv6Addressing apstra.CtPrimitiveIPv6ProtocolSessionAddressing
		switch rand.IntN(3) {
		case 0:
			ipv4Addressing = apstra.CtPrimitiveIPv4ProtocolSessionAddressingAddressed
			ipv6Addressing = apstra.CtPrimitiveIPv6ProtocolSessionAddressingNone
		case 1:
			ipv4Addressing = apstra.CtPrimitiveIPv4ProtocolSessionAddressingNone
			ipv6Addressing = oneOf(apstra.CtPrimitiveIPv6ProtocolSessionAddressingAddressed, apstra.CtPrimitiveIPv6ProtocolSessionAddressingLinkLocal)
		case 2:
			ipv4Addressing = apstra.CtPrimitiveIPv4ProtocolSessionAddressingAddressed
			ipv6Addressing = oneOf(apstra.CtPrimitiveIPv6ProtocolSessionAddressingAddressed, apstra.CtPrimitiveIPv6ProtocolSessionAddressingLinkLocal)
		}

		var peerFromLoopback bool
		var peerTo apstra.CtPrimitiveBgpPeerTo
		if ipv6Addressing == apstra.CtPrimitiveIPv6ProtocolSessionAddressingLinkLocal {
			peerFromLoopback = false
			peerTo = oneOf(
				apstra.CtPrimitiveBgpPeerToInterfaceOrIpEndpoint,
				apstra.CtPrimitiveBgpPeerToInterfaceOrSharedIpEndpoint,
			)
		} else {
			peerFromLoopback = oneOf(true, false)
			peerTo = oneOf(apstra.CtPrimitiveBgpPeerToInterfaceOrIpEndpoint,
				apstra.CtPrimitiveBgpPeerToInterfaceOrSharedIpEndpoint,
				apstra.CtPrimitiveBgpPeerToLoopback,
			)
		}

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem{
			name:               acctest.RandString(6),
			ttl:                utils.ToPtr(rand.IntN(constants.TtlMax-constants.TtlMin) + constants.TtlMin),
			bfdEnabled:         oneOf(true, false),
			password:           oneOf(acctest.RandString(6), ""),
			keepaliveTime:      keepaliveTime,
			holdTime:           holdTime,
			ipv4Addressing:     ipv4Addressing,
			ipv6Addressing:     ipv6Addressing,
			localAsn:           oneOf(utils.ToPtr(rand.IntN(constants.AsnMax+constants.AsnMin)), (*int)(nil)),
			neighborAsnDynamic: oneOf(true, false),
			peerFromLoopback:   peerFromLoopback,
			peerTo:             peerTo,
			routingPolicies:    randomRoutingPolicies(t, ctx, rand.IntN(count), client, cleanup),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveStaticRouteHCL = `{
  name              = %q
  network           = %q
  share_ip_endpoint = %s
},
`

type resourceDataCenterConnectivityTemplatePrimitiveStaticRoute struct {
	name            string
	network         net.IPNet
	shareIpEndpoint bool
}

func (o resourceDataCenterConnectivityTemplatePrimitiveStaticRoute) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveStaticRouteHCL,
			o.name,
			o.network.String(),
			strconv.FormatBool(o.shareIpEndpoint),
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveStaticRoute) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name":              o.name,
		"network":           o.network.String(),
		"share_ip_endpoint": strconv.FormatBool(o.shareIpEndpoint),
	}

	return result
}

func randomStaticRoutePrimitives(t testing.TB, _ context.Context, ipv4Count, ipv6Count int, _ *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveStaticRoute {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveStaticRoute, ipv4Count+ipv6Count)

	// add IPv4 routes
	for i := range ipv4Count {
		result[i] = resourceDataCenterConnectivityTemplatePrimitiveStaticRoute{
			name:            acctest.RandString(6),
			network:         randomPrefix(t, "10.0.0.0/8", 24),
			shareIpEndpoint: oneOf(true, false),
		}
	}

	// add IPv6 routes
	for i := range ipv6Count {
		result[ipv4Count+i] = resourceDataCenterConnectivityTemplatePrimitiveStaticRoute{
			name:            acctest.RandString(6),
			network:         randomPrefix(t, "2001:db8::/32", 64),
			shareIpEndpoint: oneOf(true, false),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingleHCL = `{
  name               = %q
  virtual_network_id = %q
  tagged             = %q
},
`

type resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle struct {
	name             string
	virtualNetworkId string
	tagged           bool
}

func (o resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingleHCL,
			o.name,
			o.virtualNetworkId,
			strconv.FormatBool(o.tagged),
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name":               o.name,
		"virtual_network_id": o.virtualNetworkId,
		"tagged":             strconv.FormatBool(o.tagged),
	}

	// todo: --------------- add static routes to map ... somehow?
	// todo: --------------- add routing policies to map ... somehow?

	return result
}

func randomVirtualNetworkSingles(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle, count)
	for i := range result {
		result[i] = resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle{
			name:             acctest.RandString(6),
			virtualNetworkId: testutils.VirtualNetworkVxlan(t, ctx, client, cleanup).String(),
			tagged:           oneOf(true, false),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultipleHCL = `{
  name           = %q
  untagged_vn_id = %s
  tagged_vn_ids  = %s
},
`

type resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple struct {
	name         string
	untaggedVnId string
	taggedVnIds  []string
}

func (o resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultipleHCL,
			o.name,
			stringOrNull(o.untaggedVnId),
			stringSliceOrNull(o.taggedVnIds),
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name":           o.name,
		"untagged_vn_id": o.untaggedVnId,
	}

	//if len(o.taggedVnIds) > 0 {
	//	todo how to add a set to this map?
	//}

	return result
}

func randomVirtualNetworkMultiples(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple, count)
	for i := range result {
		result[i] = resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple{
			name: acctest.RandString(6),
		}
		if rand.Int()%2 == 0 {
			for range rand.IntN(3) {
				result[i].taggedVnIds = append(result[i].taggedVnIds, testutils.VirtualNetworkVxlan(t, ctx, client, cleanup).String())
			}
		}
		if rand.Int()%2 == 0 || len(result[i].taggedVnIds) == 0 {
			result[i].untaggedVnId = testutils.VirtualNetworkVxlan(t, ctx, client, cleanup).String()
		}

	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraintHCL = `{
  name                       = %q
  routing_zone_constraint_id = %q
},
`

type resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint struct {
	name                    string
	routingZoneConstraintId string
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraintHCL,
			o.name,
			o.routingZoneConstraintId,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name":                       o.name,
		"routing_zone_constraint_id": o.routingZoneConstraintId,
	}

	return result
}

func randomRoutingZoneConstraints(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint {
	t.Helper()

	var routingZoneIds []apstra.ObjectId
	for i := range rand.IntN(4) {
		switch i {
		case 0: // first loop does nothing, so routingZoneIds stays nil
			continue
		case 1: // second loop changes routingZoneIds nil -> {}
			routingZoneIds = []apstra.ObjectId{}
		default: // third and subsequent loops add routing zones
			routingZoneIds = append(routingZoneIds, testutils.SecurityZoneA(t, ctx, client, cleanup))
		}
	}

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint, count)
	for i := range result {
		policyId, err := client.CreateRoutingZoneConstraint(ctx, &apstra.RoutingZoneConstraintData{
			Label:           acctest.RandString(6),
			Mode:            oneOf(apstra.RoutingZoneConstraintModeAllow, apstra.RoutingZoneConstraintModeDeny, apstra.RoutingZoneConstraintModeNone),
			MaxRoutingZones: oneOf(nil, utils.ToPtr(0), utils.ToPtr(1), utils.ToPtr(2)),
			RoutingZoneIds:  nil,
		})
		require.NoError(t, err)

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint{
			name:                    acctest.RandString(6),
			routingZoneConstraintId: policyId.String(),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveIpLinkHCL = `{
  name                        = %q
  routing_zone_id             = %q
  vlan_id                     = %s
  l3_mtu                      = %s
  ipv4_addressing_type        = %q
  ipv6_addressing_type        = %q
  bgp_peering_generic_systems = %s
  bgp_peering_ip_endpoints    = %s
  dynamic_bgp_peerings        = %s
  static_routes               = %s
},
`

type resourceDataCenterConnectivityTemplatePrimitiveIpLink struct {
	name                     string
	routingZoneId            string
	vlanId                   *int
	l3Mtu                    *int
	ipv4AddressingType       apstra.CtPrimitiveIPv4AddressingType
	ipv6AddressingType       apstra.CtPrimitiveIPv6AddressingType
	bgpPeeringGenericSystems []resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem
	bgpPeeringIpEndpoints    []resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint
	dynamicBgpPeerings       []resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering
	staticRoutes             []resourceDataCenterConnectivityTemplatePrimitiveStaticRoute
}

func (o resourceDataCenterConnectivityTemplatePrimitiveIpLink) render(indent int) string {
	bgpPeeringGenericSystems := "null"
	if len(o.bgpPeeringGenericSystems) > 0 {
		sb := new(strings.Builder)
		for _, bgpPeeringGenericSystem := range o.bgpPeeringGenericSystems {
			sb.WriteString(bgpPeeringGenericSystem.render(indent))
		}

		bgpPeeringGenericSystems = "[\n" + sb.String() + "  ]"
	}

	bgpPeeringIpEndpoints := "null"
	if len(o.bgpPeeringIpEndpoints) > 0 {
		sb := new(strings.Builder)
		for _, bgpPeeringIpEndpoint := range o.bgpPeeringIpEndpoints {
			sb.WriteString(bgpPeeringIpEndpoint.render(indent))
		}

		bgpPeeringIpEndpoints = "[\n" + sb.String() + "  ]"
	}

	dynamicBgpPeerings := "null"
	if len(o.dynamicBgpPeerings) > 0 {
		sb := new(strings.Builder)
		for _, dynamicBgpPeering := range o.dynamicBgpPeerings {
			sb.WriteString(dynamicBgpPeering.render(indent))
		}

		dynamicBgpPeerings = "[\n" + sb.String() + "  ]"
	}

	staticRoutes := "null"
	if len(o.staticRoutes) > 0 {
		sb := new(strings.Builder)
		for _, staticRoute := range o.staticRoutes {
			sb.WriteString(staticRoute.render(indent))
		}

		staticRoutes = "[\n" + sb.String() + "  ]"
	}

	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveIpLinkHCL,
			o.name,
			o.routingZoneId,
			intPtrOrNull(o.vlanId),
			intPtrOrNull(o.l3Mtu),
			utils.StringersToFriendlyString(o.ipv4AddressingType),
			utils.StringersToFriendlyString(o.ipv6AddressingType),
			bgpPeeringGenericSystems,
			bgpPeeringIpEndpoints,
			dynamicBgpPeerings,
			staticRoutes,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveIpLink) valueAsMapForChecks() map[string]string {
	result := map[string]string{
		"name":                 o.name,
		"routing_zone_id":      o.routingZoneId,
		"ipv4_addressing_type": utils.StringersToFriendlyString(o.ipv4AddressingType),
		"ipv6_addressing_type": utils.StringersToFriendlyString(o.ipv6AddressingType),
	}
	if o.vlanId != nil {
		result["vlan_id"] = strconv.Itoa(*o.vlanId)
	}
	if o.l3Mtu != nil {
		result["l3_mtu"] = strconv.Itoa(*o.l3Mtu)
	}

	return result
}

func randomIpLinks(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) []resourceDataCenterConnectivityTemplatePrimitiveIpLink {
	t.Helper()

	result := make([]resourceDataCenterConnectivityTemplatePrimitiveIpLink, count)
	for i := range result {
		var ipv4AddressingType apstra.CtPrimitiveIPv4AddressingType
		var ipv6AddressingType apstra.CtPrimitiveIPv6AddressingType
		switch rand.IntN(3) {
		case 0:
			ipv4AddressingType = apstra.CtPrimitiveIPv4AddressingTypeNumbered
			ipv6AddressingType = apstra.CtPrimitiveIPv6AddressingTypeNone
		case 1:
			ipv4AddressingType = apstra.CtPrimitiveIPv4AddressingTypeNone
			ipv6AddressingType = oneOf(apstra.CtPrimitiveIPv6AddressingTypeLinkLocal, apstra.CtPrimitiveIPv6AddressingTypeNumbered)
		case 2:
			ipv4AddressingType = apstra.CtPrimitiveIPv4AddressingTypeNumbered
			ipv6AddressingType = oneOf(apstra.CtPrimitiveIPv6AddressingTypeLinkLocal, apstra.CtPrimitiveIPv6AddressingTypeNumbered)
		}

		result[i] = resourceDataCenterConnectivityTemplatePrimitiveIpLink{
			name:                     acctest.RandString(6),
			routingZoneId:            testutils.SecurityZoneA(t, ctx, client, cleanup).String(),
			vlanId:                   oneOf(nil, utils.ToPtr(rand.IntN(4000)+100)),
			l3Mtu:                    oneOf(nil, utils.ToPtr((rand.IntN((constants.L3MtuMax-constants.L3MtuMin)/2)*2)+constants.L3MtuMin)),
			ipv4AddressingType:       ipv4AddressingType,
			ipv6AddressingType:       ipv6AddressingType,
			bgpPeeringGenericSystems: randomBgpPeeringGenericSystemPrimitives(t, ctx, rand.IntN(3), client, cleanup),
			bgpPeeringIpEndpoints:    randomBgpPeeringIpEndpointPrimitives(t, ctx, rand.IntN(3), client, cleanup),
			dynamicBgpPeerings:       randomDynamicBgpPeeringPrimitives(t, ctx, rand.IntN(3), client, cleanup),
			staticRoutes:             randomStaticRoutePrimitives(t, ctx, rand.IntN(3), rand.IntN(3), client, cleanup),
		}
	}

	return result
}
