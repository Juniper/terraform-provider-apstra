package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func tagDataAttrTypes() map[string]attr.Type {
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
		Null:     size == 0,
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: tagDataAttrTypes()},
	}
}

func tagLabelsAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Labels of tags from the global catalog to be applied to this element upon creation.",
		Optional:            true,
		Type:                types.SetType{ElemType: types.StringType},
		Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
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

func parseSliceApiTagDataToTypesSetString(in []goapstra.DesignTagData) types.Set {
	result := newTagLabelSet(len(in))
	if len(in) == 0 {
		result.Null = true
		return result
	}

	for i, tag := range in {
		result.Elems[i] = types.String{Value: tag.Label}
	}
	return result
}

func parseApiSliceTagDataToTypesSetObject(in []goapstra.DesignTagData) types.Set {
	result := newTagDataSet(len(in))
	if len(in) == 0 {
		result.Null = true
		return result
	}

	for i, tagData := range in {
		result.Elems[i] = types.Object{
			AttrTypes: tagDataAttrTypes(),
			Attrs: map[string]attr.Value{
				"name":        types.String{Value: tagData.Label},
				"description": types.String{Value: tagData.Description},
			},
		}
	}
	return result
}

func sliceTagToSetObject(in []goapstra.DesignTag) types.Set {
	result := newTagDataSet(len(in))
	if len(in) == 0 {
		result.Null = true
		return result
	}

	for i, tag := range in {
		result.Elems[i] = types.Object{
			AttrTypes: tagDataAttrTypes(),
			Attrs: map[string]attr.Value{
				"name":        types.String{Value: tag.Data.Label},
				"description": types.String{Value: tag.Data.Description},
			},
		}
	}
	return result
}
