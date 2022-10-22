package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

func tagDataAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
	}
}

func newTagLabelSet(size int) types.Set {
	if size == 0 {
		return types.Set{
			Null:     true,
			ElemType: types.StringType,
		}
	}

	return types.Set{
		Elems:    make([]attr.Value, size),
		ElemType: types.StringType,
	}
}

func tagDataElemType() attr.Type {
	return types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: tagDataAttrTypes()}}
}

func tagNameElemType() attr.Type {
	return types.SetType{
		ElemType: types.StringType}
}

func newTagDataSet(size int) types.Set {
	if size == 0 {
		return types.Set{
			Null:     true,
			ElemType: types.ObjectType{AttrTypes: tagDataAttrTypes()},
		}
	}

	return types.Set{
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
		Computed:            true,
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
	}
}

func parseSliceApiTagDataToTypesSetString(in []goapstra.DesignTagData) types.Set {
	result := newTagLabelSet(len(in))
	if result.IsNull() {
		return result
	}

	for i, tag := range in {
		result.Elems[i] = types.String{Value: tag.Label}
	}
	return result
}

func parseApiSliceTagDataToTypesSetObject(in []goapstra.DesignTagData) types.Set {
	result := newTagDataSet(len(in))
	if result.IsNull() {
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

func sliceApiTagToSetTagDataObj(in []goapstra.DesignTag) types.Set {
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

func setTagNameStringToSetTagDataObj(ctx context.Context, nameSetString types.Set, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) types.Set {
	// short circuit empty result
	if nameSetString.IsNull() || len(nameSetString.Elems) == 0 {
		return newTagDataSet(0)
	}

	// get a list of tag labels -- this could possibly use nameSetString.As(...)
	tagLabels := make([]string, len(nameSetString.Elems))
	for i, tagName := range nameSetString.Elems {
		tagLabels[i] = tagName.(types.String).Value
	}

	// get the tag slice from Apstra
	tags, err := client.GetTagsByLabels(ctx, tagLabels)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			diags.AddAttributeError(errPath, "named tag not found",
				fmt.Sprintf("at least one of the requested tags does not exist: '%s'",
					strings.Join(tagLabels, "', '")),
			)
			return newTagDataSet(0)
		}
		diags.AddError("error requesting tag data", err.Error())
	}
	return sliceApiTagToSetTagDataObj(tags)
}
