package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
)

const (
	ctPrimitiveIPv4AddressingTypeNone = "none"
	ctPrimitiveIPv6AddressingTypeNone = "none"

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

	resourceGroupNameSpineLeafLinkIpv6       = "spine_leaf_link_ips_ipv6"
	resourceGroupNameSpineSuperspineLinkIpv6 = "spine_superspine_link_ips_ipv6"
	resourceGroupNameToGenericLinkIpv6       = "to_generic_link_ips_ipv6"

	interfaceNumberingIpv4TypeNone = "none"
	interfaceNumberingIpv6TypeNone = "none"

	freeformResourceTypeIpv4     = "ipv4"
	freeformResourceTypeHostIpv4 = "host_ipv4"

	resourcePoolTypeIpv4 = "ipv4"
)

// StringersToFriendlyStrings accepts stringers (probably apstra-go-sdk
// string-able iota or enum types) and returns []string that better reflects
// terminology used by the Apstra web UI. This function is different from
// StringersToFriendlyString() in that here, each input fmt.Stringer represents
// an element in the output, where StringersToFriendlyString() uses all input
// elements to produce a single output element.
func StringersToFriendlyStrings[A fmt.Stringer](in []A) []string {
	result := make([]string, len(in))
	for i, s := range in {
		result[i] = StringersToFriendlyString(s)
	}
	return result
}

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
//
// In most cases, only a single input element is required, but some stringers
// require extra context.
func StringersToFriendlyString(in ...fmt.Stringer) string {
	if len(in) == 0 {
		return ""
	}

	switch in0 := in[0].(type) {
	case apstra.AsnAllocationScheme:
		return asnAllocationSchemeToFriendlyString(in0)
	case enum.ConfigletSection:
		return configletSectionToFriendlyString(in0, in[1:]...)
	case apstra.CtPrimitiveIPv4AddressingType:
		return ctPrimitiveIPv4AddressingTypeToFriendlyString(in0)
	case apstra.CtPrimitiveIPv6AddressingType:
		return ctPrimitiveIPv6AddressingTypeToFriendlyString(in0)
	case enum.DeployMode:
		return deployModeToFriendlyString(in0)
	case enum.FFResourceType:
		return ffResourceTypeToFriendlyString(in0)
	case enum.InterfaceNumberingIpv4Type:
		return interfaceNumberingIpv4TypeToFriendlyString(in0)
	case enum.InterfaceNumberingIpv6Type:
		return interfaceNumberingIpv6TypeToFriendlyString(in0)
	case apstra.OverlayControlProtocol:
		return overlayControlProtocolToFriendlyString(in0)
	case enum.PolicyRuleProtocol:
		return policyRuleProtocolToFriendlyString(in0)
	case enum.RefDesign:
		return refDesignToFriendlyString(in0)
	case apstra.ResourceGroupName:
		return resourceGroupNameToFriendlyString(in0)
	case enum.ResourcePoolType:
		return resourcePoolTypeToFriendlyString(in0)
	case enum.StorageSchemaPath:
		return storageSchemaPathToFriendlyString(in0)
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
	case *enum.ConfigletSection:
		return configletSectionFromFriendlyString(target, in...)
	case *apstra.CtPrimitiveIPv4AddressingType:
		return ctPrimitiveIPv4AddressingTypeFromFriendlyString(target, in...)
	case *apstra.CtPrimitiveIPv6AddressingType:
		return ctPrimitiveIPv6AddressingTypeFromFriendlyString(target, in...)
	case *enum.DeployMode:
		return nodeDeployModeFromFriendlyString(target, in...)
	case *enum.FFResourceType:
		return freeformResourceTypeFromFriendlyString(target, in...)
	case *enum.InterfaceNumberingIpv4Type:
		return interfaceNumberingIpv4TypeFromFriendlyString(target, in...)
	case *enum.InterfaceNumberingIpv6Type:
		return interfaceNumberingIpv6TypeFromFriendlyString(target, in...)
	case *apstra.OverlayControlProtocol:
		return overlayControlProtocolFromFriendlyString(target, in...)
	case *enum.PolicyRuleProtocol:
		return policyRuleProtocolFromFriendlyString(target, in[0])
	case *enum.RefDesign:
		return refDesignFromFriendlyString(target, in...)
	case *apstra.ResourceGroupName:
		return resourceGroupNameFromFriendlyString(target, in...)
	case *enum.ResourcePoolType:
		return resourcePoolTypeFromFriendlyString(target, in...)
	case *enum.StorageSchemaPath:
		return target.FromString("aos.sdk.telemetry.schemas." + in[0])
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

func configletSectionToFriendlyString(in enum.ConfigletSection, additionalInfo ...fmt.Stringer) string {
	if len(additionalInfo) == 0 {
		return in.String()
	}

	os, ok := additionalInfo[0].(enum.ConfigletStyle)
	if !ok {
		return in.String()
	}

	switch os {
	case enum.ConfigletStyleJunos:
		switch in {
		case enum.ConfigletSectionSystem:
			return junOSTopLevelHierarchical
		case enum.ConfigletSectionSetBasedSystem:
			return junOSTopLevelSetDelete
		case enum.ConfigletSectionSetBasedInterface:
			return junOSInterfaceLevelSet
		case enum.ConfigletSectionDeleteBasedInterface:
			return junOSInterfaceLevelDelete
		case enum.ConfigletSectionInterface:
			return junOSInterfaceLevelHierarchical
		}
	}

	return in.String()
}

func ctPrimitiveIPv4AddressingTypeToFriendlyString(in apstra.CtPrimitiveIPv4AddressingType) string {
	switch in {
	case apstra.CtPrimitiveIPv4AddressingTypeNone:
		return ctPrimitiveIPv4AddressingTypeNone
	}

	return in.String()
}

func ctPrimitiveIPv6AddressingTypeToFriendlyString(in apstra.CtPrimitiveIPv6AddressingType) string {
	switch in {
	case apstra.CtPrimitiveIPv6AddressingTypeNone:
		return ctPrimitiveIPv6AddressingTypeNone
	}

	return in.String()
}

func deployModeToFriendlyString(in enum.DeployMode) string {
	switch in {
	case enum.DeployModeNone:
		return nodeDeployModeNotSet
	}

	return in.String()
}

func ffResourceTypeToFriendlyString(in enum.FFResourceType) string {
	switch in {
	case enum.FFResourceTypeHostIpv4:
		return freeformResourceTypeHostIpv4
	case enum.FFResourceTypeIpv4:
		return freeformResourceTypeIpv4
	}

	return in.String()
}

func interfaceNumberingIpv4TypeToFriendlyString(in enum.InterfaceNumberingIpv4Type) string {
	switch in {
	case enum.InterfaceNumberingIpv4TypeNone:
		return interfaceNumberingIpv4TypeNone
	}

	return in.String()
}

func interfaceNumberingIpv6TypeToFriendlyString(in enum.InterfaceNumberingIpv6Type) string {
	switch in {
	case enum.InterfaceNumberingIpv6TypeNone:
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

func policyRuleProtocolToFriendlyString(in enum.PolicyRuleProtocol) string {
	return strings.ToLower(in.String())
}

func refDesignToFriendlyString(in enum.RefDesign) string {
	switch in {
	case enum.RefDesignDatacenter:
		return refDesignDataCenter
	}

	return in.String()
}

func storageSchemaPathToFriendlyString(in enum.StorageSchemaPath) string {
	s := strings.Split(in.String(), ".")
	return s[len(s)-1]
}

func resourceGroupNameToFriendlyString(in apstra.ResourceGroupName) string {
	switch in {
	case apstra.ResourceGroupNameLeafL3PeerLinkLinkIp4:
		return resourceGroupNameLeafL3PeerLinksIpv4
	case apstra.ResourceGroupNameLeafL3PeerLinkLinkIp6:
		return resourceGroupNameLeafL3PeerLinksIpv6
	case apstra.ResourceGroupNameVxlanVnIds:
		return resourceGroupNameVxlanVnIds
	case apstra.ResourceGroupNameSpineLeafIp6:
		return resourceGroupNameSpineLeafLinkIpv6
	case apstra.ResourceGroupNameSuperspineSpineIp6:
		return resourceGroupNameSpineSuperspineLinkIpv6
	case apstra.ResourceGroupNameToGenericLinkIpv6:
		return resourceGroupNameToGenericLinkIpv6
	}

	return in.String()
}

func resourcePoolTypeToFriendlyString(in enum.ResourcePoolType) string {
	switch in {
	case enum.ResourcePoolTypeIpv4:
		return resourcePoolTypeIpv4
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

func configletSectionFromFriendlyString(target *enum.ConfigletSection, in ...string) error {
	switch len(in) {
	case 0:
		return target.FromString("")
	case 1:
		return target.FromString(in[0])
	}

	section := in[0]
	platform := in[1]

	if platform != enum.ConfigletStyleJunos.String() {
		return target.FromString(section)
	}

	switch section {
	case junOSTopLevelHierarchical:
		*target = enum.ConfigletSectionSystem
	case junOSInterfaceLevelHierarchical:
		*target = enum.ConfigletSectionInterface
	case junOSTopLevelSetDelete:
		*target = enum.ConfigletSectionSetBasedSystem
	case junOSInterfaceLevelDelete:
		*target = enum.ConfigletSectionDeleteBasedInterface
	case junOSInterfaceLevelSet:
		*target = enum.ConfigletSectionSetBasedInterface
	default:
		return target.FromString(section)
	}

	return nil
}

func ctPrimitiveIPv4AddressingTypeFromFriendlyString(target *apstra.CtPrimitiveIPv4AddressingType, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case ctPrimitiveIPv4AddressingTypeNone:
		*target = apstra.CtPrimitiveIPv4AddressingTypeNone
	default:
		return target.FromString(in[0])
	}

	return nil
}

func ctPrimitiveIPv6AddressingTypeFromFriendlyString(target *apstra.CtPrimitiveIPv6AddressingType, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case ctPrimitiveIPv6AddressingTypeNone:
		*target = apstra.CtPrimitiveIPv6AddressingTypeNone
	default:
		return target.FromString(in[0])
	}

	return nil
}

func nodeDeployModeFromFriendlyString(target *enum.DeployMode, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case nodeDeployModeNotSet:
		*target = enum.DeployModeNone
	default:
		return target.FromString(in[0])
	}

	return nil
}

func freeformResourceTypeFromFriendlyString(target *enum.FFResourceType, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case freeformResourceTypeHostIpv4:
		*target = enum.FFResourceTypeHostIpv4
	case freeformResourceTypeIpv4:
		*target = enum.FFResourceTypeIpv4
	default:
		return target.FromString(in[0])
	}

	return nil
}

func interfaceNumberingIpv4TypeFromFriendlyString(target *enum.InterfaceNumberingIpv4Type, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case interfaceNumberingIpv4TypeNone:
		*target = enum.InterfaceNumberingIpv4TypeNone
	default:
		return target.FromString(in[0])
	}

	return nil
}

func interfaceNumberingIpv6TypeFromFriendlyString(target *enum.InterfaceNumberingIpv6Type, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case interfaceNumberingIpv6TypeNone:
		*target = enum.InterfaceNumberingIpv6TypeNone
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

func policyRuleProtocolFromFriendlyString(target *enum.PolicyRuleProtocol, s string) error {
	t := enum.PolicyRuleProtocols.Parse(strings.ToUpper(s))
	if t == nil {
		return fmt.Errorf("cannot parse PolicyRuleProtocol %q", s)
	}
	target.Value = t.Value
	return nil
}

func refDesignFromFriendlyString(target *enum.RefDesign, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case refDesignDataCenter:
		*target = enum.RefDesignDatacenter
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
	case resourceGroupNameSpineLeafLinkIpv6:
		*target = apstra.ResourceGroupNameSpineLeafIp6
	case resourceGroupNameSpineSuperspineLinkIpv6:
		*target = apstra.ResourceGroupNameSuperspineSpineIp6
	case resourceGroupNameToGenericLinkIpv6:
		*target = apstra.ResourceGroupNameToGenericLinkIpv6
	default:
		return target.FromString(in[0])
	}

	return nil
}

func resourcePoolTypeFromFriendlyString(target *enum.ResourcePoolType, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case resourcePoolTypeIpv4:
		*target = enum.ResourcePoolTypeIpv4
	default:
		return target.FromString(in[0])
	}

	return nil
}
