package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func validateAccessSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	as := rt.Data.AccessSwitches[i]
	if as.RedundancyProtocol == goapstra.AccessRedundancyProtocolEsi && as.EsiLagInfo == nil {
		diags.AddError("access switch ESI LAG Info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, as.Label, as.RedundancyProtocol.String()))
	}
	if as.LogicalDevice == nil {
		diags.AddError("access switch logical device info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' logical device is nil",
				rt.Id, as.Label))
	}
}
