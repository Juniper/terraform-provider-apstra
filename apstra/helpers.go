package tfapstra

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
)

func newKeyLogWriter(fileName string) (*os.File, error) {
	absPath, err := homedir.Expand(fileName)
	if err != nil {
		return nil, fmt.Errorf("error expanding home directory '%s' - %w", fileName, err)
	}

	err = os.MkdirAll(filepath.Dir(absPath), os.FileMode(0600))
	if err != nil {
		return nil, err
	}
	return os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
}

// sliceWithoutString returns a copy of in with all occurrences of t removed.
// the returned int indicates the number of occurrences removed.
func sliceWithoutString(in []string, t string) ([]string, int) {
	result := make([]string, len(in))
	var resultIdx int
	for inIdx := range in {
		if in[inIdx] == t {
			continue
		}
		result[resultIdx] = in[inIdx]
		resultIdx++
	}
	return result[:resultIdx], len(in) - resultIdx
}

// sliceWithoutInt returns a copy of in with all instances of t removed.
// the returned int indicates the number of instances removed.
func sliceWithoutInt(in []int, t int) ([]int, int) {
	result := make([]int, len(in))
	var resultIdx int
	for inIdx := range in {
		if in[inIdx] == t {
			continue
		}
		result[resultIdx] = in[inIdx]
		resultIdx++
	}
	return result[:resultIdx], len(in) - resultIdx
}

//// getAllSystemsInfo returns map[string]apstra.ManagedSystemInfo keyed by
//// device_key (switch serial number)
//func getAllSystemsInfo(ctx context.Context, client *apstra.client, diags *diag.Diagnostics) map[string]apstra.ManagedSystemInfo {
//	// pull SystemInfo for all switches managed by apstra
//	asi, err := client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
//	if err != nil {
//		diags.AddError(errApiData, fmt.Sprintf("GetAllSystemsInfo error - %s", err.Error()))
//		return nil
//	}
//
//	// organize the []ManagedSystemInfo into a map by device key (serial number)
//	deviceKeyToSystemInfo := make(map[string]apstra.ManagedSystemInfo, len(asi))
//	for i := range asi {
//		deviceKeyToSystemInfo[asi[i].DeviceKey] = asi[i]
//	}
//	return deviceKeyToSystemInfo
//}
//
//func sliceAttrValueToSliceObjectId(in []attr.Value) []apstra.ObjectId {
//	result := make([]apstra.ObjectId, len(in))
//	stringSlice := sliceAttrValueToSliceString(in)
//	for i, s := range stringSlice {
//		result[i] = apstra.ObjectId(s)
//	}
//	return result
//}
//
