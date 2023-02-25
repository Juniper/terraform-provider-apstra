package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type templateRackInfo struct {
	Count    types.Int64  `tfsdk:"count"`
	RackType types.Object `tfsdk:"rack_type"`
}

func (o templateRackInfo) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of instances of this Rack Type.",
			Computed:            true,
		},
		"rack_type": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Rack Type details.",
			Computed:            true,
			Attributes:          rackType{}.dataSourceAttributes(),
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

func newRackTypeMap(ctx context.Context, in *goapstra.TemplateRackBasedData, diags *diag.Diagnostics) types.Map {
	rackTypeMap := make(map[string]templateRackInfo, len(in.RackInfo))
	for i := range in.RackInfo {
		var tri templateRackInfo
		tri.loadApiData(ctx, &in.RackInfo[i], diags)
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: templateRackInfo{}.attrTypes()})
		}
		rackTypeMap[string(in.RackInfo[i].Id)] = tri
	}

	result, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: templateRackInfo{}.attrTypes()}, rackTypeMap)
	diags.Append(d...)
	return result
}
