package utils

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

func DeviceProfileFromDeviceKey(ctx context.Context, deviceKey string, client *goapstra.Client, diags *diag.Diagnostics) goapstra.ObjectId {
	gasi := GetAllSystemsInfo(ctx, client, diags)
	if diags.HasError() {
		return ""
	}

	si, ok := gasi[deviceKey]
	if !ok {
		diags.AddAttributeError(
			path.Root("device_key"),
			"Device Key not found",
			fmt.Sprintf("Device Key %q not found", deviceKey),
		)
		return ""
	}
	return si.Facts.AosHclModel
}
