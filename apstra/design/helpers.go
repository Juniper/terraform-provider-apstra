package design

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
