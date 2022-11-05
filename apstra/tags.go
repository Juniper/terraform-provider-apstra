package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
