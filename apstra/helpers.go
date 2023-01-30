package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
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

//func typesMapToMapStringString(in types.Map) map[string]string {
//	var out map[string]string
//	if len(in.Elems) > 0 {
//		out = make(map[string]string)
//	}
//	for k, v := range in.Elems {
//		out[k] = v.(types.String).Value
//	}
//	return out
//}

//func mapStringStringToTypesMap(in map[string]string) types.Map {
//	if len(in) == 0 {
//		return types.MapNull(types.StringType)
//	}
//
//	types.MapValue()
//	types.MapValueMust()
//	types.MapValueFrom()
//
//	out := types.Map{
//		Null:     len(in) == 0,
//		ElemType: types.StringType,
//		Elems:    make(map[string]attr.Value),
//	}
//	for k, v := range in {
//		out.Elems[k] = types.String{Value: v}
//	}
//	return out
//}

//func setOfAttrValuesMatch(a types.Set, b types.Set) bool {
//	return sliceOfAttrValuesMatch(a.Elems, b.Elems)
//}

//func sliceOfAttrValuesMatch(a []attr.Value, b []attr.Value) bool {
//	if len(a) != len(b) {
//		// obvious match failure if set size differs
//		return false
//	}
//
//loopA:
//	for _, ta := range a { // check every element of 'a' (test-a - 'ta')...
//		for bi, tb := range b { // against every element of 'b' (test-b - 'tb')...
//			if ta.Equal(tb) {
//				// match found. drop 'tb' from 'b' to speed search for next 'ta'
//				b[bi] = tb
//				b = b[:len(b)-1]
//				continue loopA
//			}
//		}
//		// if we got here, then no 'b' element matched 'ta'
//		return false
//	}
//	return true
//}

//func findMissingAsnPools(ctx context.Context, in []attr.Value, client *goapstra.Client, diags *diag.Diagnostics) []attr.Value {
//	return findMissingResourcePools(ctx, in, goapstra.ResourceTypeAsnPool, client, diags)
//}

//func findMissingIp4Pools(ctx context.Context, in []attr.Value, client *goapstra.Client, diags *diag.Diagnostics) []attr.Value {
//	return findMissingResourcePools(ctx, in, goapstra.ResourceTypeIp4Pool, client, diags)
//}

//func findMissingIp6Pools(ctx context.Context, in []attr.Value, client *goapstra.Client, diags *diag.Diagnostics) []attr.Value {
//	return findMissingResourcePools(ctx, in, goapstra.ResourceTypeIp6Pool, client, diags)
//}

//func findMissingResourcePools(ctx context.Context, in []attr.Value, poolType goapstra.ResourceType, client *goapstra.Client, diags *diag.Diagnostics) []attr.Value {
//	var poolsPerApi []goapstra.ObjectId
//	var err error
//	switch poolType {
//	case goapstra.ResourceTypeAsnPool:
//		poolsPerApi, err = client.ListAsnPoolIds(ctx)
//	case goapstra.ResourceTypeIp4Pool:
//		poolsPerApi, err = client.ListIp4PoolIds(ctx)
//	case goapstra.ResourceTypeIp6Pool:
//		poolsPerApi, err = client.ListIp6PoolIds(ctx)
//	default:
//		err = fmt.Errorf("cannot find missing pools - unsupported pool type '%s'", poolType.String())
//	}
//	if err != nil {
//		diags.AddError("error listing available resource pool IDs", err.Error())
//		return nil
//	}
//
//	var missing []attr.Value
//testPool:
//	for _, testPool := range in {
//		if testPool.IsNull() || testPool.IsUnknown() {
//			diags.AddWarning("request to validate existence of null or unknown pool",
//				fmt.Sprintf("refusing to check on pool %s (unknown: %t; null: %t",
//					testPool.String(), testPool.IsUnknown(), testPool.IsNull()))
//		}
//		for _, apiPool := range poolsPerApi {
//			if testPool.String() == fmt.Sprintf("%q", apiPool) {
//				continue testPool // this one's good, check the next testPool
//			}
//		}
//		missing = append(missing, testPool)
//	}
//	return missing
//}

//func sliceAttrValueToSliceString(in []attr.Value) []string {
//	result := make([]string, len(in))
//	for i, v := range in {
//		s := v.String()
//		switch v.(type) {
//		case types.String:
//			s = s[1:(len(s) - 1)] //String() on types.String objects uses '%q' formatting
//		}
//		result[i] = s
//	}
//	return result
//}

//func sliceAttrValueToSliceObjectId(in []attr.Value) []goapstra.ObjectId {
//	result := make([]goapstra.ObjectId, len(in))
//	stringSlice := sliceAttrValueToSliceString(in)
//	for i, s := range stringSlice {
//		result[i] = goapstra.ObjectId(s)
//	}
//	return result
//}

func bigIntToBigFloat(in *big.Int) *big.Float {
	bigval := new(big.Float)
	bigval.SetInt(in)
	return bigval
}

func allPortRoleStrings() []string {
	var allOnes goapstra.LogicalDevicePortRoleFlags

	// turn on every bit flag
	for i := 0; i < int(reflect.TypeOf(allOnes).Size()); i++ {
		allOnes = allOnes<<8 | math.MaxUint8
	}

	return allOnes.Strings()
}

//func getTfsdkTag(i interface{}, f string, diags *diag.Diagnostics) string {
//	field, ok := reflect.TypeOf(i).Elem().FieldByName(f)
//	if !ok {
//		diags.AddError(errProviderBug, fmt.Sprintf("attempt to look up nonexistent element '%s' of type '%s'", f, reflect.TypeOf(i)))
//	}
//	return field.Tag.Get("tfsdk")
//}

//func newRga(name goapstra.ResourceGroupName, set *types.Set, diags *diag.Diagnostics) *goapstra.ResourceGroupAllocation {
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

//func setStringEqual(a, b []string) bool {
//	sort.Strings(a)
//	sort.Strings(b)
//	return sliceStringEqual(a, b)
//}

//func sliceStringEqual(a, b []string) bool {
//	if len(a) != len(b) {
//		return false
//	}
//	for i := range a {
//		if a[i] != b[i] {
//			return false
//		}
//	}
//	return true
//}

//// sliceWithoutString returns a copy of in with all occurrences of t removed.
//// the returned int indicates the number of occurrences removed.
//func sliceWithoutString(in []string, t string) ([]string, int) {
//	result := make([]string, len(in))
//	var resultIdx int
//	for inIdx := range in {
//		if in[inIdx] == t {
//			continue
//		}
//		result[resultIdx] = in[inIdx]
//		resultIdx++
//	}
//	return result[:resultIdx], len(in) - resultIdx
//}

//// sliceWithoutInt returns a copy of in with all occurences of t removed.
//// the returned int indicates the number of occurences removed.
//func sliceWithoutInt(in []int, t int) ([]int, int) {
//	result := make([]int, len(in))
//	var resultIdx int
//	for inIdx := range in {
//		if in[inIdx] == t {
//			continue
//		}
//		result[resultIdx] = in[inIdx]
//		resultIdx++
//	}
//	return result[:resultIdx], len(in) - resultIdx
//}
