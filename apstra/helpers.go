package apstra

import "github.com/hashicorp/terraform-plugin-framework/types"

func sliceTfStringToSliceString(in []types.String) []string {
	//goland:noinspection GoPreferNilSlice
	out := []string{}
	for _, t := range in {
		out = append(out, t.Value)
	}
	return out
}

func sliceStringToSliceTfString(in []string) []types.String {
	var out []types.String
	for _, t := range in {
		out = append(out, types.String{Value: t})
	}
	return out
}
