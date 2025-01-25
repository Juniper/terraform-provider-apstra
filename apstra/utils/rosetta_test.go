package utils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
)

func TestRosetta(t *testing.T) {
	type tc struct {
		string    string
		stringers []fmt.Stringer
	}

	testCases := []tc{
		{string: "unique", stringers: []fmt.Stringer{apstra.AsnAllocationSchemeDistinct}},
		{string: "single", stringers: []fmt.Stringer{apstra.AsnAllocationSchemeSingle}},

		{string: "delete_based_interface", stringers: []fmt.Stringer{enum.ConfigletSectionDeleteBasedInterface, enum.ConfigletStyleCumulus}},
		{string: "file", stringers: []fmt.Stringer{enum.ConfigletSectionFile}},

		{string: "top_level_hierarchical", stringers: []fmt.Stringer{enum.ConfigletSectionSystem, enum.ConfigletStyleJunos}},
		{string: "top_level_set_delete", stringers: []fmt.Stringer{enum.ConfigletSectionSetBasedSystem, enum.ConfigletStyleJunos}},
		{string: "interface_level_hierarchical", stringers: []fmt.Stringer{enum.ConfigletSectionInterface, enum.ConfigletStyleJunos}},
		{string: "interface_level_set", stringers: []fmt.Stringer{enum.ConfigletSectionSetBasedInterface, enum.ConfigletStyleJunos}},
		{string: "interface_level_delete", stringers: []fmt.Stringer{enum.ConfigletSectionDeleteBasedInterface, enum.ConfigletStyleJunos}},

		{string: "static", stringers: []fmt.Stringer{apstra.OverlayControlProtocolNone}},
		{string: "evpn", stringers: []fmt.Stringer{apstra.OverlayControlProtocolEvpn}},

		{string: "icmp", stringers: []fmt.Stringer{enum.PolicyRuleProtocolIcmp}},
		{string: "ip", stringers: []fmt.Stringer{enum.PolicyRuleProtocolIp}},
		{string: "tcp", stringers: []fmt.Stringer{enum.PolicyRuleProtocolTcp}},
		{string: "udp", stringers: []fmt.Stringer{enum.PolicyRuleProtocolUdp}},

		{string: "datacenter", stringers: []fmt.Stringer{enum.RefDesignDatacenter}},
		{string: "freeform", stringers: []fmt.Stringer{enum.RefDesignFreeform}},

		{string: "vni_virtual_network_ids", stringers: []fmt.Stringer{apstra.ResourceGroupNameVxlanVnIds}},
		{string: "leaf_l3_peer_links", stringers: []fmt.Stringer{apstra.ResourceGroupNameLeafL3PeerLinkLinkIp4}},
		{string: "leaf_l3_peer_links_ipv6", stringers: []fmt.Stringer{apstra.ResourceGroupNameLeafL3PeerLinkLinkIp6}},

		{string: "spine_leaf_link_ips_ipv6", stringers: []fmt.Stringer{apstra.ResourceGroupNameSpineLeafIp6}},
		{string: "spine_superspine_link_ips_ipv6", stringers: []fmt.Stringer{apstra.ResourceGroupNameSuperspineSpineIp6}},
		{string: "to_generic_link_ips_ipv6", stringers: []fmt.Stringer{apstra.ResourceGroupNameToGenericLinkIpv6}},

		{string: "ipv4", stringers: []fmt.Stringer{enum.ResourcePoolTypeIpv4}},

		{string: "none", stringers: []fmt.Stringer{enum.InterfaceNumberingIpv4TypeNone}},
		{string: "none", stringers: []fmt.Stringer{enum.InterfaceNumberingIpv6TypeNone}},

		{string: "none", stringers: []fmt.Stringer{apstra.CtPrimitiveIPv4AddressingTypeNone}},
		{string: "none", stringers: []fmt.Stringer{apstra.CtPrimitiveIPv6AddressingTypeNone}},
	}

	for i, tc := range testCases {
		// test creating friendly string from iota/stringer type
		result := StringersToFriendlyString(tc.stringers...)
		if result != tc.string {
			t.Fatalf("testcase [%d], expected %q, got %q", i, tc.string, result)
		}

		// test creating iota/stringer type from friendly string
		var target StringerWithFromString
		switch tc.stringers[0].(type) {
		case enum.ConfigletSection:
			x := enum.ConfigletSection{}
			target = &x
		case apstra.CtPrimitiveIPv4AddressingType:
			x := apstra.CtPrimitiveIPv4AddressingType(-1)
			target = &x
		case apstra.CtPrimitiveIPv6AddressingType:
			x := apstra.CtPrimitiveIPv6AddressingType(-1)
			target = &x
		case apstra.AsnAllocationScheme:
			x := apstra.AsnAllocationScheme(-1)
			target = &x
		case enum.InterfaceNumberingIpv4Type:
			x := enum.InterfaceNumberingIpv4Type{}
			target = &x
		case enum.InterfaceNumberingIpv6Type:
			x := enum.InterfaceNumberingIpv6Type{}
			target = &x
		case apstra.OverlayControlProtocol:
			x := apstra.OverlayControlProtocol(-1)
			target = &x
		case enum.PolicyRuleProtocol:
			x := enum.PolicyRuleProtocol{}
			target = &x
		case enum.RefDesign:
			x := enum.RefDesign{}
			target = &x
		case apstra.ResourceGroupName:
			x := apstra.ResourceGroupName(-1)
			target = &x
		case enum.ResourcePoolType:
			x := enum.ResourcePoolType{}
			target = &x
		}

		if target == nil {
			t.Fatalf("missing case above - %q target is nil", reflect.TypeOf(tc.stringers[0]))
		}

		// stringsWithContext is the []string sent to the rosetta function to populate target
		stringsWithContext := []string{tc.string}
		for _, s := range tc.stringers[1:] {
			stringsWithContext = append(stringsWithContext, s.String())
		}

		// populate the target
		err := ApiStringerFromFriendlyString(target, stringsWithContext...)
		if err != nil {
			t.Fatalf("[%d] produced error: %s", i, err.Error())
		}

		// invoke the un-translated String() method to compare against the original input
		if target.String() != tc.stringers[0].String() {
			t.Fatalf("[%d] got %s expected %s", i, tc.stringers[0], target.String())
		}
	}
}
