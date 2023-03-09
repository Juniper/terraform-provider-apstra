package utils

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetAllSystemsInfo returns map[string]goapstra.ManagedSystemInfo keyed by
// device_key (switch serial number)
func GetAllSystemsInfo(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) map[string]goapstra.ManagedSystemInfo {
	// pull SystemInfo for all switches managed by apstra
	asi, err := client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
	if err != nil {
		diags.AddError("get managed system info", err.Error())
		return nil
	}

	// organize the []ManagedSystemInfo into a map by device key (serial number)
	deviceKeyToSystemInfo := make(map[string]goapstra.ManagedSystemInfo, len(asi)) // map-ify the Apstra output
	for _, si := range asi {
		deviceKeyToSystemInfo[si.DeviceKey] = si
	}
	return deviceKeyToSystemInfo
}
