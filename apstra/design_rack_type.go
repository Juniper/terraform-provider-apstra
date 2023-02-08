package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func validateFcdSupport(_ context.Context, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	switch fcd {
	case goapstra.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case goapstra.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("Unsupported Fabric Connectivity Design '%s'",
				fcd.String()))
	}
}

func validateRackType(ctx context.Context, in *goapstra.RackType, diags *diag.Diagnostics) {
	if in.Data == nil {
		diags.AddError("rack type has no data", fmt.Sprintf("rack type '%s' data object is nil", in.Id))
		return
	}

	validateFcdSupport(ctx, in.Data.FabricConnectivityDesign, diags)
	if diags.HasError() {
		return
	}

	for i := range in.Data.LeafSwitches {
		validateLeafSwitch(in, i, diags)
	}

	for i := range in.Data.AccessSwitches {
		validateAccessSwitch(in, i, diags)
	}

	for i := range in.Data.GenericSystems {
		validateGenericSystem(in, i, diags)
	}
}
