package utils

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetAllSystemsInfo returns map[string]apstra.ManagedSystemInfo keyed by
// device_key (switch serial number)
func GetAllSystemsInfo(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) map[string]apstra.ManagedSystemInfo {
	// pull SystemInfo for all switches managed by apstra
	asi, err := client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
	if err != nil {
		diags.AddError("get managed system info", err.Error())
		return nil
	}

	// organize the []ManagedSystemInfo into a map by device key (serial number)
	deviceKeyToSystemInfo := make(map[string]apstra.ManagedSystemInfo, len(asi)) // map-ify the Apstra output
	for _, si := range asi {
		deviceKeyToSystemInfo[si.DeviceKey] = si
	}
	return deviceKeyToSystemInfo
}

func BlueprintExists(ctx context.Context, client *apstra.Client, id apstra.ObjectId, diags *diag.Diagnostics) bool {
	ids, err := client.ListAllBlueprintIds(ctx)
	if err != nil {
		diags.AddError("error listing blueprints", err.Error())
		return false
	}

	for i := range ids {
		if ids[i] == id {
			return true
		}
	}
	return false
}
