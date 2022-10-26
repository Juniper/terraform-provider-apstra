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
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func tagDataAttrType() attr.Type {
	return types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: tagDataAttrTypes()}}
}

func tagIdsAttrType() attr.Type {
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

func tagIdsAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "IDs of tags from the global catalog to be applied to this element upon creation.",
		Optional:            true,
		Type:                types.SetType{ElemType: types.StringType},
		Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
		//PlanModifiers:       []tfsdk.AttributePlanModifier{useStateForUnknownNull()},
		//PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
	}
}

func tagsDataAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Details any tags applied to the element.",
		Computed:            true,
		PlanModifiers: tfsdk.AttributePlanModifiers{
			useStateForUnknownNull(),
			tagDataTrackTagIds(),
		},
		//PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
			"name": {
				MarkdownDescription: "Tag name (label) field.",
				Computed:            true,
				Type:                types.StringType,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"description": {
				MarkdownDescription: "Tag description field.",
				Computed:            true,
				Type:                types.StringType,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
		}),
	}
}

func parseApiSliceTagDataToTypesSetString(in []goapstra.DesignTagData) types.Set {
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

func setTagIdStringToSetTagDataObj(ctx context.Context, tagIdSetString types.Set, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) types.Set {
	// short circuit empty result
	if tagIdSetString.IsNull() || len(tagIdSetString.Elems) == 0 {
		return newTagDataSet(0)
	}

	// get a []DesignTagData based on the supplied tag ID set
	tagData := make([]goapstra.DesignTagData, len(tagIdSetString.Elems))
	for i, tagId := range tagIdSetString.Elems {
		tag, err := client.GetTag(ctx, goapstra.ObjectId(tagId.(types.String).Value))
		if err != nil {
			var ace goapstra.ApstraClientErr
			if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
				diags.AddAttributeError(errPath, "tag not found", fmt.Sprintf("tag '%s' not found", tagId))
			}
			diags.AddError("error requesting tag data", err.Error())
			return newTagDataSet(0)
		}
		tagData[i] = *tag.Data
	}

	return parseApiSliceTagDataToTypesSetObject(tagData)
}
