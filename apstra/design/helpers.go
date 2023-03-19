package design

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func translateAsnAllocationSchemeFromWebUi(in string) string {
	switch in {
	case AsnAllocationUnique:
		return goapstra.AsnAllocationSchemeDistinct.String()
	}
	return in
}

func asnAllocationSchemeToString(in goapstra.AsnAllocationScheme, diags *diag.Diagnostics) string {
	switch in {
	case goapstra.AsnAllocationSchemeSingle:
		return AsnAllocationSingle
	case goapstra.AsnAllocationSchemeDistinct:
		return AsnAllocationUnique
	default:
		diags.AddError(errProviderBug, fmt.Sprintf("unknown ASN allocation scheme: %d", in))
		return ""
	}
}

func overlayControlProtocolToString(in goapstra.OverlayControlProtocol, diags *diag.Diagnostics) string {
	switch in {
	case goapstra.OverlayControlProtocolEvpn:
		return OverlayControlProtocolEvpn
	case goapstra.OverlayControlProtocolNone:
		return OverlayControlProtocolStatic
	default:
		diags.AddError(errProviderBug, fmt.Sprintf("unknown Overlay Control Protocol: %d", in))
		return ""
	}
}

const (
	JunOSTopLevelHierarchical       = "top_level_hierarchical"
	JunOSTopLevelSetDelete          = "top_level_set_delete"
	JunOSInterfaceLevelHierarchical = "interface_level_hierarchical"
	JunOSInterfaceLevelSet          = "interface_level_set"
	JunOSInterfaceLevelDelete       = "interface_level_delete"
)

func ConfigletSectionIotaToFriendlyString(diags *diag.Diagnostics, in goapstra.ConfigletSection, ctx ...fmt.Stringer) string {
	// We know that the first entry is going to be the platform OS
	var os goapstra.PlatformOS
	err := os.FromString(ctx[0].String())
	if err != nil {
		diags.AddError("Unknown network OS", "Unknown network OS %s"ctx[0].String())
		return ""
	}
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
			diags.AddError("Unknown Section for JunOS", "Unknown section %s for JunOS"os.String())
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
func IotaToFriendlyString(diags *diag.Diagnostics, in fmt.Stringer, ctx ...fmt.Stringer) string {
	switch in.(type) {
	case goapstra.ConfigletSection:
		var i goapstra.ConfigletSection
		err := i.FromString(in)
		if err != nil {
			diags.AddError(errProviderBug, fmt.Sprintf("Unknown Configlet Section %s", in.String()))
			return ""
		}
		return ConfigletSectionIotaToFriendlyString(diags, in, ctx)
	default:
		x := types.StringNull()
		x.ValueString()
		return in.String()
	}
}

func FriendlyStringToIota(diags *diag.Diagnostics, in string, ctx ...string) fmt.Stringer {
}

//switch on typeof (in[0])
//return typeioatatoui(in...)
//
//return return in[0].string()
//}
