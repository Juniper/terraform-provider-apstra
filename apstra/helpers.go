package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mitchellh/go-homedir"
	"math/big"
	"os"
	"path/filepath"
)

const (
	asnAllocationSingle          = "single"
	asnAllocationUnique          = "unique"
	overlayControlProtocolEvpn   = "evpn"
	overlayControlProtocolStatic = "static"
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

func bigIntToBigFloat(in *big.Int) *big.Float {
	bigval := new(big.Float)
	bigval.SetInt(in)
	return bigval
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

func asnAllocationSchemeToString(in goapstra.AsnAllocationScheme, diags *diag.Diagnostics) string {
	switch in {
	case goapstra.AsnAllocationSchemeSingle:
		return asnAllocationSingle
	case goapstra.AsnAllocationSchemeDistinct:
		return asnAllocationUnique
	default:
		diags.AddError(errProviderBug, fmt.Sprintf("unknown ASN allocation scheme: %d", in))
		return ""
	}
}

func overlayControlProtocolToString(in goapstra.OverlayControlProtocol, diags *diag.Diagnostics) string {
	switch in {
	case goapstra.OverlayControlProtocolEvpn:
		return overlayControlProtocolEvpn
	case goapstra.OverlayControlProtocolNone:
		return overlayControlProtocolStatic
	default:
		diags.AddError(errProviderBug, fmt.Sprintf("unknown Overlay Control Protocol: %d", in))
		return ""
	}
}

func translateAsnAllocationSchemeFromWebUi(in string) string {
	switch in {
	case asnAllocationUnique:
		return goapstra.AsnAllocationSchemeDistinct.String()
	}
	return in
}

//// getAllSystemsInfo returns map[string]goapstra.ManagedSystemInfo keyed by
//// device_key (switch serial number)
//func getAllSystemsInfo(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) map[string]goapstra.ManagedSystemInfo {
//	// pull SystemInfo for all switches managed by apstra
//	asi, err := client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
//	if err != nil {
//		diags.AddError(errApiData, fmt.Sprintf("GetAllSystemsInfo error - %s", err.Error()))
//		return nil
//	}
//
//	// organize the []ManagedSystemInfo into a map by device key (serial number)
//	deviceKeyToSystemInfo := make(map[string]goapstra.ManagedSystemInfo, len(asi))
//	for i := range asi {
//		deviceKeyToSystemInfo[asi[i].DeviceKey] = asi[i]
//	}
//	return deviceKeyToSystemInfo
//}
//
//func sliceAttrValueToSliceObjectId(in []attr.Value) []goapstra.ObjectId {
//	result := make([]goapstra.ObjectId, len(in))
//	stringSlice := sliceAttrValueToSliceString(in)
//	for i, s := range stringSlice {
//		result[i] = goapstra.ObjectId(s)
//	}
//	return result
//}
//
//func newRga(name goapstra.ResourceGroupName, poolIds []goapstra.ObjectId, diags *diag.Diagnostics) *goapstra.ResourceGroupAllocation {
//	poolIds := make([]goapstra.ObjectId, 0)
//	if !set.IsNull() {
//		poolIds = sliceAttrValueToSliceObjectId(set.Elems)
//	}
//	return &goapstra.ResourceGroupAllocation{
//		ResourceGroup: goapstra.ResourceGroup{
//			Type: resourceTypeNameFromResourceGroupName(name, diags),
//			Name: name,
//		},
//		PoolIds: poolIds,
//	}
//}
