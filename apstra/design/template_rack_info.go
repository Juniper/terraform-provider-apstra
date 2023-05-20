package design

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TemplateRackInfo struct {
	Count    types.Int64  `tfsdk:"count"`
	RackType types.Object `tfsdk:"rack_type"`
}

func (o TemplateRackInfo) DataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of instances of this Rack Type.",
			Computed:            true,
		},
		"rack_type": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Rack Type attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          RackType{}.DataSourceAttributesNested(),
		},
	}
}

func (o TemplateRackInfo) ResourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of instances of this Rack Type.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"rack_type": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Rack Type attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          RackType{}.ResourceAttributesNested(),
		},
	}
}

func (o TemplateRackInfo) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"count":     types.Int64Type,
		"rack_type": types.ObjectType{AttrTypes: RackType{}.AttrTypes()},
	}
}

func (o *TemplateRackInfo) LoadApiData(ctx context.Context, in *apstra.TemplateRackBasedRackInfo, diags *diag.Diagnostics) {
	if in.Count == 0 {
		diags.AddError(errProviderBug, "attempt to load TemplateRackInfo with 0 instances of rack type")
		return
	}

	o.Count = types.Int64Value(int64(in.Count))
	o.RackType = NewRackTypeObject(ctx, in.RackTypeData, diags)
}

func NewRackInfoMap(ctx context.Context, in *apstra.TemplateRackBasedData, diags *diag.Diagnostics) types.Map {
	rackTypeMap := make(map[apstra.ObjectId]TemplateRackInfo, len(in.RackInfo))
	for key, apiData := range in.RackInfo {
		var tri TemplateRackInfo
		tri.LoadApiData(ctx, &apiData, diags)
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: TemplateRackInfo{}.AttrTypes()})
		}
		rackTypeMap[key] = tri
	}

	result, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: TemplateRackInfo{}.AttrTypes()}, rackTypeMap)
	diags.Append(d...)
	return result
}
