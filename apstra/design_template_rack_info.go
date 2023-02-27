package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type templateRackInfo struct {
	Count    types.Int64  `tfsdk:"count"`
	RackType types.Object `tfsdk:"rack_type"`
}

func (o templateRackInfo) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	panic("templateRackInfo.dataSourceAttributes() should never be used")
	return map[string]dataSourceSchema.Attribute{}
}

func (o templateRackInfo) dataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of instances of this Rack Type.",
			Computed:            true,
		},
		"rack_type": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Rack Type attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          rackType{}.dataSourceAttributesNested(),
		},
	}
}

func (o templateRackInfo) resourceAttributes() map[string]resourceSchema.Attribute {
	panic("templateRackInfo.resourceAttributes() should never be used")
	return map[string]resourceSchema.Attribute{}
}

func (o templateRackInfo) resourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of instances of this Rack Type.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"rack_type": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Rack Type attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          rackType{}.resourceAttributesNested(),
		},
	}
}

func (o templateRackInfo) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"count":     types.Int64Type,
		"rack_type": types.ObjectType{AttrTypes: rackType{}.attrTypes()},
	}
}

func (o *templateRackInfo) loadApiData(ctx context.Context, in *goapstra.TemplateRackBasedRackInfo, diags *diag.Diagnostics) {
	if in.Count == 0 {
		diags.AddError(errProviderBug, "attempt to load templateRackInfo with 0 instances of rack type")
		return
	}

	o.Count = types.Int64Value(int64(in.Count))
	o.RackType = newRackTypeObject(ctx, in.RackTypeData, diags)
}

func newRackInfoMap(ctx context.Context, in *goapstra.TemplateRackBasedData, diags *diag.Diagnostics) types.Map {
	rackTypeMap := make(map[goapstra.ObjectId]templateRackInfo, len(in.RackInfo))
	for key, apiData := range in.RackInfo {
		var tri templateRackInfo
		tri.loadApiData(ctx, &apiData, diags)
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: templateRackInfo{}.attrTypes()})
		}
		rackTypeMap[key] = tri
	}

	result, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: templateRackInfo{}.attrTypes()}, rackTypeMap)
	diags.Append(d...)
	return result
}
