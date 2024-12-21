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
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
)

const resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRouteHCL = `{
  routing_zone_id = %q
  network         = %q
  next_hop        = %q
},
`

type resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute struct {
	routingZoneId string
	network       net.IPNet
	nextHop       net.IP
}

func (o resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRouteHCL,
			o.routingZoneId,
			o.network.String(),
			o.nextHop.String(),
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute) testChecks(path string) [][]string {
	var result [][]string
	result = append(result, []string{"TestCheckResourceAttr", path + ".routing_zone_id", o.routingZoneId})
	result = append(result, []string{"TestCheckResourceAttr", path + ".network", o.network.String()})
	result = append(result, []string{"TestCheckResourceAttr", path + ".next_hop", o.nextHop.String()})
	return result
}

func randomCustomStaticRoutes(t testing.TB, ctx context.Context, ipv4Count, ipv6Count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute, ipv4Count+ipv6Count)

	rzName := acctest.RandString(6)
	rzId, err := client.CreateSecurityZone(ctx, &apstra.SecurityZoneData{
		Label:   rzName,
		SzType:  apstra.SecurityZoneTypeEVPN,
		VrfName: rzName,
	})
	require.NoError(t, err)

	// add IPv4 routes
	for range ipv4Count {
		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute{
			routingZoneId: rzId.String(),
			network:       randomSlash31(t, "10.0.0.0/8"),
			nextHop:       randIpvAddressMust(t, "10.0.0.0/8"),
		}
	}

	// add IPv6 routes
	for range ipv6Count {
		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveCustomStaticRoute{
			routingZoneId: rzId.String(),
			network:       randomSlash127(t, "2001:db8::/32"),
			nextHop:       randIpvAddressMust(t, "2001:db8::/32"),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicyHCL = `{
  routing_policy_id = %q
},
`

type resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy struct {
	routingPolicyId string
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicyHCL,
			o.routingPolicyId,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy) testChecks(path string) [][]string {
	var result [][]string
	result = append(result, []string{"TestCheckResourceAttr", path + ".routing_policy_id", o.routingPolicyId})
	return result
}

func randomRoutingPolicies(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy, count)
	for range count {
		policyId, err := client.CreateRoutingPolicy(ctx, &apstra.DcRoutingPolicyData{
			Label:        acctest.RandString(6),
			PolicyType:   apstra.DcRoutingPolicyTypeUser,
			ImportPolicy: oneOf(apstra.DcRoutingPolicyImportPolicyAll, apstra.DcRoutingPolicyImportPolicyDefaultOnly, apstra.DcRoutingPolicyImportPolicyExtraOnly),
		})
		require.NoError(t, err)

		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy{
			routingPolicyId: policyId.String(),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpointHCL = `{
  neighbor_asn     = %s
  ttl              = %s
  bfd_enabled      = %s
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
	neighborAsn     *int
	ttl             *int
	bfdEnabled      bool
	password        string
	keepaliveTime   *int
	holdTime        *int
	localAsn        *int
	ipv4Address     net.IP
	ipv6Address     net.IP
	routingPolicies map[string]resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy
}

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint) render(indent int) string {
	routingPolicies := "null"
	if len(o.routingPolicies) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.routingPolicies {
			sb.WriteString(tfapstra.Indent(indent, k+" = "+v.render(indent)))
		}

		routingPolicies = "{\n" + sb.String() + "  }"
	}

	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpointHCL,
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

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint) testChecks(path string) [][]string {
	var result [][]string
	if o.neighborAsn == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".neighbor_asn"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".neighbor_asn", strconv.Itoa(*o.neighborAsn)})
	}
	if o.ttl == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".ttl"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".ttl", strconv.Itoa(*o.ttl)})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".bfd_enabled", strconv.FormatBool(o.bfdEnabled)})
	if o.password == "" {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".password"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".password", o.password})
	}
	if o.keepaliveTime == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".keepalive_time"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".keepalive_time", strconv.Itoa(*o.keepaliveTime)})
	}
	if o.holdTime == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".hold_time"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".hold_time", strconv.Itoa(*o.holdTime)})
	}
	if o.localAsn == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".local_asn"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".local_asn", strconv.Itoa(*o.localAsn)})
	}
	if o.ipv4Address == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".ipv4_address"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".ipv4_address", o.ipv4Address.String()})
	}
	if o.ipv6Address == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".ipv6_address"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".ipv6_address", o.ipv6Address.String()})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".routing_policies.%", strconv.Itoa(len(o.routingPolicies))})
	for k, v := range o.routingPolicies {
		result = append(result, v.testChecks(path+".routing_policies."+k)...)
	}

	return result
}

func randomBgpPeeringIpEndpointPrimitives(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint, count)
	for range count {
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

		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint{
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
  ttl              = %s
  bfd_enabled      = %s
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
	routingPolicies map[string]resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy
}

func (o resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering) render(indent int) string {
	routingPolicies := "null"
	if len(o.routingPolicies) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.routingPolicies {
			sb.WriteString(tfapstra.Indent(indent, k+" = "+v.render(indent)))
		}

		routingPolicies = "{\n" + sb.String() + "  }"
	}

	return tfapstra.Indent(indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeeringHCL,
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

func (o resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering) testChecks(path string) [][]string {
	var result [][]string
	if o.ttl == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".ttl"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".ttl", strconv.Itoa(*o.ttl)})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".bfd_enabled", strconv.FormatBool(o.bfdEnabled)})
	if o.password == "" {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".password"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".password", o.password})
	}
	if o.keepaliveTime == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".keepalive_time"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".keepalive_time", strconv.Itoa(*o.keepaliveTime)})
	}
	if o.holdTime == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".hold_time"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".hold_time", strconv.Itoa(*o.holdTime)})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".ipv4_enabled", strconv.FormatBool(o.ipv4Enabled)})
	result = append(result, []string{"TestCheckResourceAttr", path + ".ipv6_enabled", strconv.FormatBool(o.ipv6Enabled)})
	if o.localAsn == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".local_asn"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".local_asn", strconv.Itoa(*o.localAsn)})
	}
	if o.ipv4PeerPrefix.IP == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".ipv4_peer_prefix"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".ipv4_peer_prefix", o.ipv4PeerPrefix.String()})
	}
	if o.ipv6PeerPrefix.IP == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".ipv6_peer_prefix"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".ipv6_peer_prefix", o.ipv6PeerPrefix.String()})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".routing_policies.%", strconv.Itoa(len(o.routingPolicies))})
	for k, v := range o.routingPolicies {
		result = append(result, v.testChecks(path+".routing_policies."+k)...)
	}

	return result
}

func randomDynamicBgpPeeringPrimitives(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering, count)
	for range count {
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

		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering{
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
  ttl                  = %s
  bfd_enabled          = %s
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
	routingPolicies    map[string]resourceDataCenterConnectivityTemplatePrimitiveRoutingPolicy
}

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem) render(indent int) string {
	routingPolicies := "null"
	if len(o.routingPolicies) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.routingPolicies {
			sb.WriteString(tfapstra.Indent(indent, k+" = "+v.render(indent)))
		}

		routingPolicies = "{\n" + sb.String() + "  }"
	}

	return tfapstra.Indent(indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystemHCL,
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

func (o resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem) testChecks(path string) [][]string {
	var result [][]string
	if o.ttl == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".ttl"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".ttl", strconv.Itoa(*o.ttl)})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".bfd_enabled", strconv.FormatBool(o.bfdEnabled)})
	if o.password == "" {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".password"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".password", o.password})
	}
	if o.keepaliveTime == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".keepalive_time"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".keepalive_time", strconv.Itoa(*o.keepaliveTime)})
	}
	if o.holdTime == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".hold_time"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".hold_time", strconv.Itoa(*o.holdTime)})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".ipv4_addressing_type", o.ipv4Addressing.String()})
	result = append(result, []string{"TestCheckResourceAttr", path + ".ipv6_addressing_type", o.ipv6Addressing.String()})
	if o.localAsn == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".local_asn"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".local_asn", strconv.Itoa(*o.localAsn)})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".neighbor_asn_dynamic", strconv.FormatBool(o.neighborAsnDynamic)})
	result = append(result, []string{"TestCheckResourceAttr", path + ".peer_from_loopback", strconv.FormatBool(o.peerFromLoopback)})
	result = append(result, []string{"TestCheckResourceAttr", path + ".peer_to", o.peerTo.String()})
	result = append(result, []string{"TestCheckResourceAttr", path + ".routing_policies.%", strconv.Itoa(len(o.routingPolicies))})
	for k, v := range o.routingPolicies {
		result = append(result, v.testChecks(path+".routing_policies."+k)...)
	}

	return result
}

func randomBgpPeeringGenericSystemPrimitives(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem, count)
	for range count {
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

		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem{
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
  network           = %q
  share_ip_endpoint = %s
},
`

type resourceDataCenterConnectivityTemplatePrimitiveStaticRoute struct {
	network         net.IPNet
	shareIpEndpoint bool
}

func (o resourceDataCenterConnectivityTemplatePrimitiveStaticRoute) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveStaticRouteHCL,
			o.network.String(),
			strconv.FormatBool(o.shareIpEndpoint),
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveStaticRoute) testChecks(path string) [][]string {
	var result [][]string
	result = append(result, []string{"TestCheckResourceAttr", path + ".network", o.network.String()})
	result = append(result, []string{"TestCheckResourceAttr", path + ".share_ip_endpoint", strconv.FormatBool(o.shareIpEndpoint)})
	return result
}

func randomStaticRoutePrimitives(t testing.TB, _ context.Context, ipv4Count, ipv6Count int, _ *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveStaticRoute {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveStaticRoute, ipv4Count+ipv6Count)

	// add IPv4 routes
	for range ipv4Count {
		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveStaticRoute{
			network:         randomPrefix(t, "10.0.0.0/8", 24),
			shareIpEndpoint: oneOf(true, false),
		}
	}

	// add IPv6 routes
	for range ipv6Count {
		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveStaticRoute{
			network:         randomPrefix(t, "2001:db8::/32", 64),
			shareIpEndpoint: oneOf(true, false),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingleHCL = `{
  virtual_network_id = %q
  tagged             = %q
},
`

type resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle struct {
	virtualNetworkId string
	tagged           bool
}

func (o resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingleHCL,
			o.virtualNetworkId,
			strconv.FormatBool(o.tagged),
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle) testChecks(path string) [][]string {
	var result [][]string
	result = append(result, []string{"TestCheckResourceAttr", path + ".virtual_network_id", o.virtualNetworkId})
	result = append(result, []string{"TestCheckResourceAttr", path + ".tagged", strconv.FormatBool(o.tagged)})
	return result
}

func randomVirtualNetworkSingles(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle, count)
	for range count {
		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkSingle{
			virtualNetworkId: testutils.VirtualNetworkVxlan(t, ctx, client, cleanup).String(),
			tagged:           oneOf(true, false),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultipleHCL = `{
  untagged_vn_id = %s
  tagged_vn_ids  = %s
},
`

type resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple struct {
	untaggedVnId string
	taggedVnIds  []string
}

func (o resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultipleHCL,
			stringOrNull(o.untaggedVnId),
			stringSliceOrNull(o.taggedVnIds),
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple) testChecks(path string) [][]string {
	var result [][]string
	if o.untaggedVnId == "" {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".untagged_vn_id"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".untagged_vn_id", o.untaggedVnId})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".tagged_vn_ids.#", strconv.Itoa(len(o.taggedVnIds))})
	for _, taggedVnId := range o.taggedVnIds {
		result = append(result, []string{"TestCheckTypeSetElemAttr", path + ".tagged_vn_ids.*", taggedVnId})
	}
	return result
}

func randomVirtualNetworkMultiples(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple, count)
	for range count {
		var taggedVnIds []string
		var untaggedVnId string
		if rand.Int()%2 == 0 {
			for range rand.IntN(3) {
				taggedVnIds = append(taggedVnIds, testutils.VirtualNetworkVxlan(t, ctx, client, cleanup).String())
			}
		}
		if rand.Int()%2 == 0 || len(taggedVnIds) == 0 {
			untaggedVnId = testutils.VirtualNetworkVxlan(t, ctx, client, cleanup).String()
		}
		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveVirtualNetworkMultiple{
			taggedVnIds:  taggedVnIds,
			untaggedVnId: untaggedVnId,
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraintHCL = `{
  routing_zone_constraint_id = %q
},
`

type resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint struct {
	routingZoneConstraintId string
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint) render(indent int) string {
	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraintHCL,
			o.routingZoneConstraintId,
		),
	)
}

func (o resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint) testChecks(path string) [][]string {
	var result [][]string
	result = append(result, []string{"TestCheckResourceAttr", path + ".routing_zone_constraint_id", o.routingZoneConstraintId})
	return result
}

func randomRoutingZoneConstraints(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint {
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

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint, count)
	for range count {
		policyId, err := client.CreateRoutingZoneConstraint(ctx, &apstra.RoutingZoneConstraintData{
			Label:           acctest.RandString(6),
			Mode:            oneOf(enum.RoutingZoneConstraintModeAllow, enum.RoutingZoneConstraintModeDeny, enum.RoutingZoneConstraintModeNone),
			MaxRoutingZones: oneOf(nil, utils.ToPtr(0), utils.ToPtr(1), utils.ToPtr(2)),
			RoutingZoneIds:  nil,
		})
		require.NoError(t, err)

		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveRoutingZoneConstraint{
			routingZoneConstraintId: policyId.String(),
		}
	}

	return result
}

const resourceDataCenterConnectivityTemplatePrimitiveIpLinkHCL = `{
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
	routingZoneId            string
	vlanId                   *int
	l3Mtu                    *int
	ipv4AddressingType       apstra.CtPrimitiveIPv4AddressingType
	ipv6AddressingType       apstra.CtPrimitiveIPv6AddressingType
	bgpPeeringGenericSystems map[string]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringGenericSystem
	bgpPeeringIpEndpoints    map[string]resourceDataCenterConnectivityTemplatePrimitiveBgpPeeringIpEndpoint
	dynamicBgpPeerings       map[string]resourceDataCenterConnectivityTemplatePrimitiveDynamicBgpPeering
	staticRoutes             map[string]resourceDataCenterConnectivityTemplatePrimitiveStaticRoute
}

func (o resourceDataCenterConnectivityTemplatePrimitiveIpLink) render(indent int) string {
	bgpPeeringGenericSystems := "null"
	if len(o.bgpPeeringGenericSystems) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.bgpPeeringGenericSystems {
			sb.WriteString(tfapstra.Indent(indent, k+" = "+v.render(indent)))
		}

		bgpPeeringGenericSystems = "{\n" + sb.String() + "  }"
	}

	bgpPeeringIpEndpoints := "null"
	if len(o.bgpPeeringIpEndpoints) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.bgpPeeringIpEndpoints {
			sb.WriteString(tfapstra.Indent(indent, k+" = "+v.render(indent)))
		}

		bgpPeeringIpEndpoints = "{\n" + sb.String() + "  }"
	}

	dynamicBgpPeerings := "null"
	if len(o.dynamicBgpPeerings) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.dynamicBgpPeerings {
			sb.WriteString(tfapstra.Indent(indent, k+" = "+v.render(indent)))
		}

		dynamicBgpPeerings = "{\n" + sb.String() + "  }"
	}

	staticRoutes := "null"
	if len(o.staticRoutes) > 0 {
		sb := new(strings.Builder)
		for k, v := range o.staticRoutes {
			sb.WriteString(tfapstra.Indent(indent, k+" = "+v.render(indent)))
		}

		staticRoutes = "{\n" + sb.String() + "  }"
	}

	return tfapstra.Indent(
		indent,
		fmt.Sprintf(resourceDataCenterConnectivityTemplatePrimitiveIpLinkHCL,
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

func (o resourceDataCenterConnectivityTemplatePrimitiveIpLink) testChecks(path string) [][]string {
	var result [][]string
	result = append(result, []string{"TestCheckResourceAttr", path + ".routing_zone_id", o.routingZoneId})
	if o.vlanId == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".vlan_id"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".vlan_id", strconv.Itoa(*o.vlanId)})
	}
	if o.l3Mtu == nil {
		result = append(result, []string{"TestCheckNoResourceAttr", path + ".l3_mtu"})
	} else {
		result = append(result, []string{"TestCheckResourceAttr", path + ".l3_mtu", strconv.Itoa(*o.l3Mtu)})
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".ipv4_addressing_type", o.ipv4AddressingType.String()})
	result = append(result, []string{"TestCheckResourceAttr", path + ".ipv6_addressing_type", o.ipv6AddressingType.String()})
	result = append(result, []string{"TestCheckResourceAttr", path + ".bgp_peering_generic_systems.%", strconv.Itoa(len(o.bgpPeeringGenericSystems))})
	for k, v := range o.bgpPeeringGenericSystems {
		result = append(result, v.testChecks(path+".bgp_peering_generic_systems."+k)...)
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".bgp_peering_ip_endpoints.%", strconv.Itoa(len(o.bgpPeeringIpEndpoints))})
	for k, v := range o.bgpPeeringIpEndpoints {
		result = append(result, v.testChecks(path+".bgp_peering_ip_endpoints."+k)...)
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".dynamic_bgp_peerings.%", strconv.Itoa(len(o.dynamicBgpPeerings))})
	for k, v := range o.dynamicBgpPeerings {
		result = append(result, v.testChecks(path+".dynamic_bgp_peerings."+k)...)
	}
	result = append(result, []string{"TestCheckResourceAttr", path + ".static_routes.%", strconv.Itoa(len(o.staticRoutes))})
	for k, v := range o.staticRoutes {
		result = append(result, v.testChecks(path+".static_routes."+k)...)
	}
	return result
}

func randomIpLinks(t testing.TB, ctx context.Context, count int, client *apstra.TwoStageL3ClosClient, cleanup bool) map[string]resourceDataCenterConnectivityTemplatePrimitiveIpLink {
	t.Helper()

	result := make(map[string]resourceDataCenterConnectivityTemplatePrimitiveIpLink, count)
	for range count {
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

		result[acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)] = resourceDataCenterConnectivityTemplatePrimitiveIpLink{
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
