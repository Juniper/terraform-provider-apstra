package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
)

const (
	junOSTopLevelHierarchical       = "top_level_hierarchical"
	junOSTopLevelSetDelete          = "top_level_set_delete"
	junOSInterfaceLevelHierarchical = "interface_level_hierarchical"
	junOSInterfaceLevelSet          = "interface_level_set"
	junOSInterfaceLevelDelete       = "interface_level_delete"

	asnAllocationUnique = "unique"

	overlayControlProtocolStatic = "static"

	refDesignDataCenter = "datacenter"

	nodeDeployModeNotSet = "not_set"

	resourceGroupNameVxlanVnIds          = "vni_virtual_network_ids"
	resourceGroupNameLeafL3PeerLinksIpv4 = "leaf_l3_peer_links"
	resourceGroupNameLeafL3PeerLinksIpv6 = "leaf_l3_peer_links_ipv6"

	// search for todos with 'enable_rosetta_for_pools_with_leading_ipv6' to enable rosetta for these items.
	// total 18 occurences between this file and the test file
	// resourceGroupNameSpineLeafLinkIpv6       = "spine_leaf_link_ips_ipv6"       // todo: enable_rosetta_for_pools_with_leading_ipv6
	// resourceGroupNameSpineSuperspineLinkIpv6 = "spine_superspine_link_ips_ipv6" // todo: enable_rosetta_for_pools_with_leading_ipv6
	// resourceGroupNameToGenericLinkIpv6       = "to_generic_link_ips_ipv6"       // todo: enable_rosetta_for_pools_with_leading_ipv6

	interfaceNumberingIpv4TypeNone = "none"
	interfaceNumberingIpv6TypeNone = "none"
)

type StringerWithFromString interface {
	String() string
	FromString(string) error
}

// StringersToFriendlyString accepts stringers (probably apstra-go-sdk
// string-able iota or enum types) and returns a string that better reflects
// terminology used by the Apstra web UI.
//
// For example, the API uses "distinct" where the web UI uses "unique".
// This function turns apstra.AsnAllocationSchemeDistinct into "unique".
func StringersToFriendlyString(in ...fmt.Stringer) string {
	if len(in) == 0 {
		return ""
	}

	switch in0 := in[0].(type) {
	case apstra.AsnAllocationScheme:
		return asnAllocationSchemeToFriendlyString(in0)
	case apstra.ConfigletSection:
		return configletSectionToFriendlyString(in0, in[1:]...)
	case apstra.DeployMode:
		return deployModeToFriendlyString(in0)
	case apstra.InterfaceNumberingIpv4Type:
		return interfaceNumberingIpv4TypeToFriendlyString(in0)
	case apstra.InterfaceNumberingIpv6Type:
		return interfaceNumberingIpv6TypeToFriendlyString(in0)
	case apstra.OverlayControlProtocol:
		return overlayControlProtocolToFriendlyString(in0)
	case apstra.PolicyRuleProtocol:
		return policyRuleProtocolToFriendlyString(in0)
	case apstra.RefDesign:
		return refDesignToFriendlyString(in0)
	case apstra.ResourceGroupName:
		return resourceGroupNameToFriendlyString(in0)
	}

	return in[0].String()
}

// ApiStringerFromFriendlyString attempts to populate a StringerWithFromString
// using one or more friendly 'in' strings. It is used to turn friendly strings
// used in the web UI into types used by the SDK and ultimately the API. For
// example, we can get apstra.AsnAllocationSchemeDistinct directly from a string
// by invoking apstra.AsnAllocationScheme.FromString("distinct"). But the web UI
// uses "unique", rather than "distinct". This method will be able to translate
// "unique" into an apstra.AsnAllocationScheme value.
func ApiStringerFromFriendlyString(target StringerWithFromString, in ...string) error {
	if len(in) == 0 {
		return errors.New("ApiStringerFromFriendlyString called with no string input")
	}

	switch target := target.(type) {
	case *apstra.AsnAllocationScheme:
		return asnAllocationSchemeFromFriendlyString(target, in...)
	case *apstra.ConfigletSection:
		return configletSectionFromFriendlyString(target, in...)
	case *apstra.DeployMode:
		return nodeDeployModeFromFriendlyString(target, in...)
	case *apstra.InterfaceNumberingIpv4Type:
		return interfaceNumberingIpv4TypeFromFriendlyString(target, in...)
	case *apstra.InterfaceNumberingIpv6Type:
		return interfaceNumberingIpv6TypeFromFriendlyString(target, in...)
	case *apstra.OverlayControlProtocol:
		return overlayControlProtocolFromFriendlyString(target, in...)
	case *apstra.PolicyRuleProtocol:
		return policyRuleProtocolFromFriendlyString(target, in[0])
	case *apstra.RefDesign:
		return refDesignFromFriendlyString(target, in...)
	case *apstra.ResourceGroupName:
		return resourceGroupNameFromFriendlyString(target, in...)
	}

	return target.FromString(in[0])
}

func asnAllocationSchemeToFriendlyString(in apstra.AsnAllocationScheme) string {
	switch in {
	case apstra.AsnAllocationSchemeDistinct:
		return asnAllocationUnique
	}

	return in.String()
}

func configletSectionToFriendlyString(in apstra.ConfigletSection, additionalInfo ...fmt.Stringer) string {
	if len(additionalInfo) == 0 {
		return in.String()
	}

	os, ok := additionalInfo[0].(apstra.PlatformOS)
	if !ok {
		return in.String()
	}

	switch os {
	case apstra.PlatformOSJunos:
		switch in {
		case apstra.ConfigletSectionSystem:
			return junOSTopLevelHierarchical
		case apstra.ConfigletSectionSetBasedSystem:
			return junOSTopLevelSetDelete
		case apstra.ConfigletSectionSetBasedInterface:
			return junOSInterfaceLevelSet
		case apstra.ConfigletSectionDeleteBasedInterface:
			return junOSInterfaceLevelDelete
		case apstra.ConfigletSectionInterface:
			return junOSInterfaceLevelHierarchical
		}
	}

	return in.String()
}

func deployModeToFriendlyString(in apstra.DeployMode) string {
	switch in {
	case apstra.DeployModeNone:
		return nodeDeployModeNotSet
	}

	return in.String()
}

func interfaceNumberingIpv4TypeToFriendlyString(in apstra.InterfaceNumberingIpv4Type) string {
	switch in {
	case apstra.InterfaceNumberingIpv4TypeNone:
		return interfaceNumberingIpv4TypeNone
	}

	return in.String()
}

func interfaceNumberingIpv6TypeToFriendlyString(in apstra.InterfaceNumberingIpv6Type) string {
	switch in {
	case apstra.InterfaceNumberingIpv6TypeNone:
		return interfaceNumberingIpv6TypeNone
	}

	return in.String()
}

func overlayControlProtocolToFriendlyString(in apstra.OverlayControlProtocol) string {
	switch in {
	case apstra.OverlayControlProtocolNone:
		return overlayControlProtocolStatic
	}

	return in.String()
}

func policyRuleProtocolToFriendlyString(in apstra.PolicyRuleProtocol) string {
	return strings.ToLower(in.String())
}

func refDesignToFriendlyString(in apstra.RefDesign) string {
	switch in {
	case apstra.RefDesignTwoStageL3Clos:
		return refDesignDataCenter
	}

	return in.String()
}

func resourceGroupNameToFriendlyString(in apstra.ResourceGroupName) string {
	switch in {
	case apstra.ResourceGroupNameLeafL3PeerLinkLinkIp4:
		return resourceGroupNameLeafL3PeerLinksIpv4
	case apstra.ResourceGroupNameLeafL3PeerLinkLinkIp6:
		return resourceGroupNameLeafL3PeerLinksIpv6
	case apstra.ResourceGroupNameVxlanVnIds:
		return resourceGroupNameVxlanVnIds
		// case apstra.ResourceGroupNameSpineLeafIp6: //       todo: enable_rosetta_for_pools_with_leading_ipv6
		//	return resourceGroupNameSpineLeafLinkIpv6 //       todo: enable_rosetta_for_pools_with_leading_ipv6
		// case apstra.ResourceGroupNameSuperspineSpineIp6: // todo: enable_rosetta_for_pools_with_leading_ipv6
		//	return resourceGroupNameSpineSuperspineLinkIpv6 // todo: enable_rosetta_for_pools_with_leading_ipv6
		// case apstra.ResourceGroupNameToGenericLinkIpv6: //  todo: enable_rosetta_for_pools_with_leading_ipv6
		//	return resourceGroupNameToGenericLinkIpv6 //       todo: enable_rosetta_for_pools_with_leading_ipv6
	}

	return in.String()
}

func asnAllocationSchemeFromFriendlyString(target *apstra.AsnAllocationScheme, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case asnAllocationUnique:
		*target = apstra.AsnAllocationSchemeDistinct
	default:
		return target.FromString(in[0])
	}

	return nil
}

func configletSectionFromFriendlyString(target *apstra.ConfigletSection, in ...string) error {
	switch len(in) {
	case 0:
		return target.FromString("")
	case 1:
		return target.FromString(in[0])
	}

	section := in[0]
	platform := in[1]

	if platform != apstra.PlatformOSJunos.String() {
		return target.FromString(section)
	}

	switch section {
	case junOSTopLevelHierarchical:
		*target = apstra.ConfigletSectionSystem
	case junOSInterfaceLevelHierarchical:
		*target = apstra.ConfigletSectionInterface
	case junOSTopLevelSetDelete:
		*target = apstra.ConfigletSectionSetBasedSystem
	case junOSInterfaceLevelDelete:
		*target = apstra.ConfigletSectionDeleteBasedInterface
	case junOSInterfaceLevelSet:
		*target = apstra.ConfigletSectionSetBasedInterface
	default:
		return target.FromString(section)
	}

	return nil
}

func nodeDeployModeFromFriendlyString(target *apstra.DeployMode, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case nodeDeployModeNotSet:
		*target = apstra.DeployModeNone
	default:
		return target.FromString(in[0])
	}

	return nil
}

func interfaceNumberingIpv4TypeFromFriendlyString(target *apstra.InterfaceNumberingIpv4Type, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case interfaceNumberingIpv4TypeNone:
		*target = apstra.InterfaceNumberingIpv4TypeNone
	default:
		return target.FromString(in[0])
	}

	return nil
}

func interfaceNumberingIpv6TypeFromFriendlyString(target *apstra.InterfaceNumberingIpv6Type, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case interfaceNumberingIpv6TypeNone:
		*target = apstra.InterfaceNumberingIpv6TypeNone
	default:
		return target.FromString(in[0])
	}

	return nil
}

func overlayControlProtocolFromFriendlyString(target *apstra.OverlayControlProtocol, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case overlayControlProtocolStatic:
		*target = apstra.OverlayControlProtocolNone
	default:
		return target.FromString(in[0])
	}

	return nil
}

func policyRuleProtocolFromFriendlyString(target *apstra.PolicyRuleProtocol, s string) error {
	t := apstra.PolicyRuleProtocols.Parse(strings.ToUpper(s))
	if t == nil {
		return fmt.Errorf("cannot parse PolicyRuleProtocol %q", s)
	}
	target.Value = t.Value
	return nil
}

func refDesignFromFriendlyString(target *apstra.RefDesign, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case refDesignDataCenter:
		*target = apstra.RefDesignTwoStageL3Clos
	default:
		return target.FromString(in[0])
	}

	return nil
}

func resourceGroupNameFromFriendlyString(target *apstra.ResourceGroupName, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case resourceGroupNameLeafL3PeerLinksIpv4:
		*target = apstra.ResourceGroupNameLeafL3PeerLinkLinkIp4
	case resourceGroupNameLeafL3PeerLinksIpv6:
		*target = apstra.ResourceGroupNameLeafL3PeerLinkLinkIp6
	case resourceGroupNameVxlanVnIds:
		*target = apstra.ResourceGroupNameVxlanVnIds
	// case resourceGroupNameSpineLeafLinkIpv6: //              todo: enable_rosetta_for_pools_with_leading_ipv6
	//	*target = apstra.ResourceGroupNameSpineLeafIp6 //       todo: enable_rosetta_for_pools_with_leading_ipv6
	// case resourceGroupNameSpineSuperspineLinkIpv6: //        todo: enable_rosetta_for_pools_with_leading_ipv6
	//	*target = apstra.ResourceGroupNameSuperspineSpineIp6 // todo: enable_rosetta_for_pools_with_leading_ipv6
	// case resourceGroupNameToGenericLinkIpv6: //              todo: enable_rosetta_for_pools_with_leading_ipv6
	//	*target = apstra.ResourceGroupNameToGenericLinkIpv6 //  todo: enable_rosetta_for_pools_with_leading_ipv6
	default:
		return target.FromString(in[0])
	}

	return nil
}
