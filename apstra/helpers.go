package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// stringValueOrNull returns a types.String based on the supplied string. If the
// supplied string is empty, the returned types.String will be flagged as null.
func stringValueOrNull(_ context.Context, in string, _ *diag.Diagnostics) types.String {
	if in == "" {
		return types.StringNull()
	}

	return types.StringValue(in)
}

// stringValueWithNull returns a types.String based on the supplied inStr. If
// inStr matches nullStr or is empty, the returned types.String will be flagged
// as null.
func stringValueWithNull(ctx context.Context, inStr string, nullStr string, diags *diag.Diagnostics) types.String {
	if inStr == nullStr {
		return types.StringNull()
	}
	return stringValueOrNull(ctx, inStr, diags)
}

// mapValueOrNull returns a types.Map based on the supplied elements. If the
// supplied elements is empty, the returned types.Map will be flagged as null.
func mapValueOrNull[T any](ctx context.Context, elementType attr.Type, elements map[string]T, diags *diag.Diagnostics) types.Map {
	if len(elements) == 0 {
		return types.MapNull(elementType)
	}

	result, d := types.MapValueFrom(ctx, elementType, elements)
	diags.Append(d...)
	return result
}

// listValueOrNull returns a types.List based on the supplied elements. If the
// supplied elements is empty, the returned types.List will be flagged as null.
func listValueOrNull[T any](ctx context.Context, elementType attr.Type, elements []T, diags *diag.Diagnostics) types.List {
	if len(elements) == 0 {
		return types.ListNull(elementType)
	}

	result, d := types.ListValueFrom(ctx, elementType, elements)
	diags.Append(d...)
	return result
}

// setValueOrNull returns a types.Set based on the supplied elements. If the
// supplied elements is empty, the returned types.Set will be flagged as null.
func setValueOrNull[T any](ctx context.Context, elementType attr.Type, elements []T, diags *diag.Diagnostics) types.Set {
	if len(elements) == 0 {
		return types.SetNull(elementType)
	}

	result, d := types.SetValueFrom(ctx, elementType, elements)
	diags.Append(d...)
	return result
}

func asnAllocationSchemeFromString(in string, diags *diag.Diagnostics) goapstra.AsnAllocationScheme {
	switch in {
	case asnAllocationSingle:
		return goapstra.AsnAllocationSchemeSingle
	case asnAllocationUnique:
		return goapstra.AsnAllocationSchemeDistinct
	default:
		diags.AddError(errProviderBug, fmt.Sprintf("unknown ASN allocation scheme: %q", in))
		return -1
	}
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

func overlayControlProtocolFromString(in string, diags *diag.Diagnostics) goapstra.OverlayControlProtocol {
	switch in {
	case overlayControlProtocolEvpn:
		return goapstra.OverlayControlProtocolEvpn
	case overlayControlProtocolStatic:
		return goapstra.OverlayControlProtocolNone
	default:
		diags.AddError(errProviderBug, fmt.Sprintf("unknown ASN Allocation Scheme: %q", in))
		return -1
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

func translateAsnAllocationSchemeForWebUi(in string) string {
	switch in {
	case asnAllocationUnique:
		return goapstra.AsnAllocationSchemeDistinct.String()
	}
	return in
}
