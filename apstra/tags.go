package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

type tagData struct {
	Name        string `tfsdk:"name"`
	Description string `tfsdk:"description"`
}

func (o tagData) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":        types.StringType,
			"description": types.StringType}}
}

func (o *tagData) parseApi(in *goapstra.DesignTagData) {
	o.Name = in.Label
	o.Description = in.Description
}
