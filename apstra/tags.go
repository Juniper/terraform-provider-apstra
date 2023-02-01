package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//func tagIdsAttributeSchema() tfsdk.Attribute {
//	return tfsdk.Attribute{
//		MarkdownDescription: "IDs of tags from the global catalog to be applied to this element upon creation.",
//		Optional:            true,
//		Type:                types.SetType{ElemType: types.StringType},
//		Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
//		//PlanModifiers:       []tfsdk.AttributePlanModifier{useStateForUnknownNull()},
//		//PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
//	}
//}

func tagsDataAttributeSchema() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		MarkdownDescription: "Details any tags applied to the element.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Tag name (label) field.",
					Computed:            true,
				},
				"description": schema.StringAttribute{
					MarkdownDescription: "Tag description field.",
					Computed:            true,
				},
			},
		},
	}
}

type tagData struct {
	Name        string `tfsdk:"name"`
	Description string `tfsdk:"description"`
}

func (o tagData) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
	}
}

func (o tagData) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *tagData) parseApi(in *goapstra.DesignTagData) {
	o.Name = in.Label
	o.Description = in.Description
}

func newTagSet(ctx context.Context, in []goapstra.DesignTagData, diags *diag.Diagnostics) types.Set {
	tags := make([]tagData, len(in))
	for i, tag := range in {
		tags[i] = tagData{
			Name:        tag.Label,
			Description: tag.Description,
		}
	}

	result, d := types.SetValueFrom(ctx, tagData{}.attrType(), &tags)
	diags.Append(d...)

	return result
}
