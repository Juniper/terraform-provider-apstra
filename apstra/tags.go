package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func tagAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
	}
}

func newTagLabelSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.StringType,
	}
}

func newTagDataSet(size int) types.Set {
	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: tagAttrTypes()},
	}
}

func sdkTagsDataToTagDataObj(tags []goapstra.DesignTagData) types.Set {
	result := newTagDataSet(len(tags))
	for i, tag := range tags {
		result.Elems[i] = types.Object{
			AttrTypes: tagAttrTypes(),
			Attrs: map[string]attr.Value{
				"name":        types.String{Value: tag.Label},
				"description": types.String{Value: tag.Description},
			},
		}
	}
	return result
}

func tagLabelsAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Labels of tags from the global catalog to be applied to this element upon creation.",
		Optional:            true,
		Type:                types.SetType{ElemType: types.StringType},
	}
}

func tagsDataAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Details any tags applied to the element.",
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"name": {
				MarkdownDescription: "Tag name (label) field.",
				Computed:            true,
				Type:                types.StringType,
			},
			"description": {
				MarkdownDescription: "Tag description field.",
				Computed:            true,
				Type:                types.StringType,
			},
		}),
		Computed: true,
	}
}

func sliceTagDataToSetString(in []goapstra.DesignTagData) types.Set {
	result := newTagLabelSet(len(in))
	for i, tag := range in {
		result.Elems[i] = types.String{Value: tag.Label}
	}
	return result
}
