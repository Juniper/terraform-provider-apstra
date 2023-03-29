package utils

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"strings"
)

func asnAllocationSchemeToFriendlyString(in goapstra.AsnAllocationScheme) string {
	switch in {
	case goapstra.AsnAllocationSchemeSingle:
		return AsnAllocationSingle
	case goapstra.AsnAllocationSchemeDistinct:
		return AsnAllocationUnique
	default:
		return ""
	}
}

func overlayControlProtocolToFriendlyString(in goapstra.OverlayControlProtocol) string {
	switch in {
	case goapstra.OverlayControlProtocolEvpn:
		return OverlayControlProtocolEvpn
	case goapstra.OverlayControlProtocolNone:
		return OverlayControlProtocolStatic
	default:
		return ""
	}
}

const (
	JunOSTopLevelHierarchical       = "top_level_hierarchical"
	JunOSTopLevelSetDelete          = "top_level_set_delete"
	JunOSInterfaceLevelHierarchical = "interface_level_hierarchical"
	JunOSInterfaceLevelSet          = "interface_level_set"
	JunOSInterfaceLevelDelete       = "interface_level_delete"
	JunOSUnknown                    = "unknown_section"

	AsnAllocationSingle = "single"
	AsnAllocationUnique = "unique"

	OverlayControlProtocolEvpn   = "evpn"
	OverlayControlProtocolStatic = "static"
)

func configletSectionIotaToFriendlyString(in goapstra.ConfigletSection, ctx ...fmt.Stringer) string {
	if len(ctx) == 0 {
		return ""
	}
	os := ctx[0].(goapstra.PlatformOS)
	switch os {
	case goapstra.PlatformOSJunos:
		switch in {
		case goapstra.ConfigletSectionSystem:
			return JunOSTopLevelHierarchical
		case goapstra.ConfigletSectionInterface:
			return JunOSInterfaceLevelHierarchical
		case goapstra.ConfigletSectionSetBasedSystem:
			return JunOSTopLevelSetDelete
		case goapstra.ConfigletSectionDeleteBasedInterface:
			return JunOSInterfaceLevelDelete
		case goapstra.ConfigletSectionSetBasedInterface:
			return JunOSInterfaceLevelSet
		default:
			return JunOSUnknown
		}
	default:
		return in.String()
	}
	return ""
}

/*
	This accepts a Iota, potential context strings and returns a string that is what the customer would see on the UI
    For example, for Junos, the configletsection Iota
*/
func StringersToFriendlyString(in ...fmt.Stringer) string {
	if len(in) == 0 {
		return ""
	}
	switch in[0].(type) {
	case goapstra.ConfigletSection:
		return configletSectionIotaToFriendlyString(in[0].(goapstra.ConfigletSection), in[1:]...)
	case goapstra.OverlayControlProtocol:
		return overlayControlProtocolToFriendlyString(in[0].(goapstra.OverlayControlProtocol))
	case goapstra.AsnAllocationScheme:
		return asnAllocationSchemeToFriendlyString(in[0].(goapstra.AsnAllocationScheme))
	default:
		return in[0].String()
	}
}

type StringerWithFromString interface {
	fmt.Stringer
	FromString(string) error
}

func FriendlyStringToAPIStringer(target StringerWithFromString, in ...string) error {
	switch target.(type) {
	case *goapstra.ConfigletSection:
		return configletSectionFriendlyStringToIota(target.(*goapstra.ConfigletSection), in...)
	case *goapstra.AsnAllocationScheme:
		return asnAllocationSchemeFriendlyStringtoIota(target.(*goapstra.AsnAllocationScheme), in...)
	default:
		return target.FromString(in[0])
	}
}

func configletSectionFriendlyStringToIota(out *goapstra.ConfigletSection, in ...string) error {
	if len(in) < 1 {
		return out.FromString("")
	}
	if len(in) < 2 {
		return out.FromString(in[0])
	}
	cs := in[0]
	os := in[1]
	if strings.ToUpper(os) != strings.ToUpper(goapstra.PlatformOSJunos.String()) {
		return out.FromString(cs)
	}
	switch cs {
	case JunOSTopLevelHierarchical:
		*out = goapstra.ConfigletSectionSystem
		return nil
	case JunOSInterfaceLevelHierarchical:
		*out = goapstra.ConfigletSectionInterface
		return nil
	case JunOSTopLevelSetDelete:
		*out = goapstra.ConfigletSectionSetBasedSystem
		return nil
	case JunOSInterfaceLevelDelete:
		*out = goapstra.ConfigletSectionDeleteBasedInterface
		return nil
	case JunOSInterfaceLevelSet:
		*out = goapstra.ConfigletSectionSetBasedInterface
		return nil
	default:
		return out.FromString(cs)
	}
	return nil
}

func asnAllocationSchemeFriendlyStringtoIota(out *goapstra.AsnAllocationScheme, in ...string) error {
	if len(in) < 1 {
		return out.FromString("")
	}
	switch in[0] {
	case AsnAllocationUnique:
		*out = goapstra.AsnAllocationSchemeDistinct
	default:
		return out.FromString(in[0])
	}
	return nil
}
