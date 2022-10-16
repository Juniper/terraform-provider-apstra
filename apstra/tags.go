package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func tagAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"label":       types.StringType,
		"description": types.StringType,
	}
}

func newTagSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: tagAttrTypes()},
	}
}

func newTagSetFromSliceDesignTagData(tags []goapstra.DesignTagData) types.Set {
	result := newTagSet(len(tags))
	for i, tag := range tags {
		result.Elems[i] = types.Object{
			AttrTypes: tagAttrTypes(),
			Attrs: map[string]attr.Value{
				"label":       types.String{Value: tag.Label},
				"description": types.String{Value: tag.Description},
			},
		}
	}
	return result
}
