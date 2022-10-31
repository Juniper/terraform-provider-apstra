package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func logicalDeviceAttrType() attr.Type {
	return types.ObjectType{
		AttrTypes: logicalDeviceDataAttrTypes()}
}

type logicalDeviceData struct {
	Name   string               `tfsdk:"name"`
	Panels []logicalDevicePanel `tfsdk:"panels"'`
}

func (o *logicalDeviceData) parseApi(in *goapstra.LogicalDeviceData) {
	o.Name = in.DisplayName
	o.Panels = make([]logicalDevicePanel, len(in.Panels))

	for i := range o.Panels {
		o.Panels[i].parseApi(&in.Panels[i])
	}
}

func (o logicalDeviceData) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"panels": types.ListType{
				ElemType: logicalDevicePanel{}.attrType(),
			},
		},
	}
}

func parseApiLogicalDeviceData(in *goapstra.LogicalDeviceData) *logicalDeviceData {
	panels := make([]logicalDevicePanel, len(in.Panels))
	for i := range in.Panels {
		panels[i].parseApi(&in.Panels[i])
	}
	return &logicalDeviceData{
		Name:   in.DisplayName,
		Panels: panels,
	}
}
func logicalDeviceDataAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":   types.StringType,
		"panels": types.ListType{ElemType: types.ObjectType{AttrTypes: panelAttrTypes()}},
	}
}

func logicalDeviceDataAttributeSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
		Computed:            true,
		PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"panels": dPanelsAttributeSchema(),
			"name": {
				MarkdownDescription: "Logical device display name.",
				Computed:            true,
				Type:                types.StringType,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
		}),
	}
}

func parseApiLogicalDeviceToTypesObject(ctx context.Context, in *goapstra.LogicalDeviceData, diags *diag.Diagnostics) types.Object {
	structLogicalDeviceData := parseApiLogicalDeviceData(in)
	result, d := types.ObjectValueFrom(ctx, logicalDeviceDataAttrTypes(), structLogicalDeviceData)
	diags.Append(d...)
	return result
}
