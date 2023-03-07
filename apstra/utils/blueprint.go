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

func DeviceProfileIdFromInterfaceMapId(ctx context.Context, bpId goapstra.ObjectId, iMapId goapstra.ObjectId, client *goapstra.Client, diags *diag.Diagnostics) goapstra.ObjectId {
	var response struct {
		Items []struct {
			DeviceProfile struct {
				Id goapstra.ObjectId `json:"id"`
			} `json:"n_device_profile"`
		} `json:"items"`
	}

	query := client.NewQuery(bpId).SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"id", goapstra.QEStringVal(iMapId.String())},
			//{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("device_profile")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("device_profile")},
			{"name", goapstra.QEStringVal("n_device_profile")},
		})

	err := query.Do(&response)

	if err != nil {
		diags.AddError("error querying graphDB for device profile", err.Error())
		return ""
	}

	switch len(response.Items) {
	case 0:
		diags.AddError("no results when querying for Device Profile", fmt.Sprintf("query string %q", query.String()))
		return ""
	case 1:
		return response.Items[0].DeviceProfile.Id
	default:
		diags.AddError("multiple matches when querying for Device Profile", fmt.Sprintf("query string %q", query.String()))
		return ""
	}
}

func DeviceProfileIdFromDeviceKey(ctx context.Context, deviceKey string, client *goapstra.Client, diags *diag.Diagnostics) goapstra.ObjectId {
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
