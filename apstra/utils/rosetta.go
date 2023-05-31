package utils

import (
	"errors"
	"fmt"
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
	case apstra.NodeDeployMode:
		return nodeDeployModeToFriendlyString(in[0].(apstra.NodeDeployMode))
	case apstra.OverlayControlProtocol:
		return overlayControlProtocolToFriendlyString(in[0].(apstra.OverlayControlProtocol))
	case apstra.RefDesign:
		return refDesignToFriendlyString(in[0].(apstra.RefDesign))
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

	//lint:ignore S1034 see issue #127
	switch target.(type) {
	case *apstra.AsnAllocationScheme:
		return asnAllocationSchemeFromFriendlyString(target.(*apstra.AsnAllocationScheme), in...)
	case *apstra.ConfigletSection:
		return configletSectionFromFriendlyString(target.(*apstra.ConfigletSection), in...)
	case *apstra.NodeDeployMode:
		return nodeDeployModeFromFriendlyString(target.(*apstra.NodeDeployMode), in...)
	case *apstra.OverlayControlProtocol:
		return overlayControlProtocolFromFriendlyString(target.(*apstra.OverlayControlProtocol), in...)
	case *apstra.RefDesign:
		return refDesignFromFriendlyString(target.(*apstra.RefDesign), in...)
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

func nodeDeployModeToFriendlyString(in apstra.NodeDeployMode) string {
	switch in {
	case apstra.NodeDeployModeNone:
		return nodeDeployModeNotSet
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

func refDesignToFriendlyString(in apstra.RefDesign) string {
	switch in {
	case apstra.RefDesignDatacenter:
		return refDesignDataCenter
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

func nodeDeployModeFromFriendlyString(target *apstra.NodeDeployMode, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case nodeDeployModeNotSet:
		*target = apstra.NodeDeployModeNone
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

func refDesignFromFriendlyString(target *apstra.RefDesign, in ...string) error {
	if len(in) == 0 {
		return target.FromString("")
	}

	switch in[0] {
	case refDesignDataCenter:
		*target = apstra.RefDesignDatacenter
	default:
		return target.FromString(in[0])
	}

	return nil
}
