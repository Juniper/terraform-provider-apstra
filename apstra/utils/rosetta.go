package utils

import (
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
)

const (
	JunOSTopLevelHierarchical       = "top_level_hierarchical"
	JunOSTopLevelSetDelete          = "top_level_set_delete"
	JunOSInterfaceLevelHierarchical = "interface_level_hierarchical"
	JunOSInterfaceLevelSet          = "interface_level_set"
	JunOSInterfaceLevelDelete       = "interface_level_delete"

	AsnAllocationSingle = "single"
	AsnAllocationUnique = "unique"

	OverlayControlProtocolEvpn   = "evpn"
	OverlayControlProtocolStatic = "static"
)

type StringerWithFromString interface {
	String() string
	FromString(string) error
}

// StringersToFriendlyString accepts stringers (probably apstra-go-sdk
// string-able iota types) and returns a string that better reflects terminology
// used by the Apstra web UI. For example, the API uses "distinct" where the web
// UI uses "unique". This function turns apstra.AsnAllocationSchemeDistinct into
// "unique".
func StringersToFriendlyString(in ...fmt.Stringer) string {
	if len(in) == 0 {
		return ""
	}

	switch in[0].(type) {
	case apstra.AsnAllocationScheme:
		return asnAllocationSchemeToFriendlyString(in[0].(apstra.AsnAllocationScheme))
	case apstra.ConfigletSection:
		return configletSectionToFriendlyString(in[0].(apstra.ConfigletSection), in[1:]...)
	case apstra.OverlayControlProtocol:
		return overlayControlProtocolToFriendlyString(in[0].(apstra.OverlayControlProtocol))
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

	switch target.(type) {
	case *apstra.AsnAllocationScheme:
		return asnAllocationSchemeFromFriendlyString(target.(*apstra.AsnAllocationScheme), in...)
	case *apstra.ConfigletSection:
		return configletSectionFromFriendlyString(target.(*apstra.ConfigletSection), in...)
	case *apstra.OverlayControlProtocol:
		return overlayControlProtocolFromFriendlyString(target.(*apstra.OverlayControlProtocol), in...)
	}

	return target.FromString(in[0])
}

func asnAllocationSchemeToFriendlyString(in apstra.AsnAllocationScheme) string {
	switch in {
	case apstra.AsnAllocationSchemeSingle:
		return AsnAllocationSingle
	case apstra.AsnAllocationSchemeDistinct:
		return AsnAllocationUnique
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
			return JunOSTopLevelHierarchical
		case apstra.ConfigletSectionSetBasedSystem:
			return JunOSTopLevelSetDelete
		case apstra.ConfigletSectionSetBasedInterface:
			return JunOSInterfaceLevelSet
		case apstra.ConfigletSectionDeleteBasedInterface:
			return JunOSInterfaceLevelDelete
		case apstra.ConfigletSectionInterface:
			return JunOSInterfaceLevelHierarchical
		}
	}

	return in.String()
}

func overlayControlProtocolToFriendlyString(in apstra.OverlayControlProtocol) string {
	switch in {
	case apstra.OverlayControlProtocolEvpn:
		return OverlayControlProtocolEvpn
	case apstra.OverlayControlProtocolNone:
		return OverlayControlProtocolStatic
	}

	return in.String()
}

func asnAllocationSchemeFromFriendlyString(target *apstra.AsnAllocationScheme, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case AsnAllocationUnique:
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
	case JunOSTopLevelHierarchical:
		*target = apstra.ConfigletSectionSystem
	case JunOSInterfaceLevelHierarchical:
		*target = apstra.ConfigletSectionInterface
	case JunOSTopLevelSetDelete:
		*target = apstra.ConfigletSectionSetBasedSystem
	case JunOSInterfaceLevelDelete:
		*target = apstra.ConfigletSectionDeleteBasedInterface
	case JunOSInterfaceLevelSet:
		*target = apstra.ConfigletSectionSetBasedInterface
	default:
		return target.FromString(section)
	}

	return nil
}

func overlayControlProtocolFromFriendlyString(target *apstra.OverlayControlProtocol, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case OverlayControlProtocolEvpn:
		*target = apstra.OverlayControlProtocolEvpn
	case OverlayControlProtocolStatic:
		*target = apstra.OverlayControlProtocolNone
	default:
		return target.FromString(in[0])
	}

	return nil
}
