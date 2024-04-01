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

type TemplatePodInfo struct {
	Count   types.Int64  `tfsdk:"count"`
	PodType types.Object `tfsdk:"pod_type"`
}

func (o TemplatePodInfo) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"count":    types.Int64Type,
		"pod_type": types.ObjectType{AttrTypes: TemplateRackBased{}.AttrTypes()},
	}
}
func (o TemplatePodInfo) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of instances of this Pod Type.",
			Computed:            true,
		},
		"pod_type": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Pod Type attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          TemplateRackBased{}.DataSourceAttributesNested(),
		},
	}
}

func (o TemplatePodInfo) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of instances of this Pod Type.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"pod_type": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Pod Type attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          TemplateRackBased{}.ResourceAttributesNested(),
		},
	}
}

func (o *TemplatePodInfo) LoadApiData(ctx context.Context, in *apstra.TemplateRackBasedRackInfo, diags *diag.Diagnostics) {
	if in.Count == 0 {
		diags.AddError(errProviderBug, "attempt to load TemplatePodInfo with 0 instances of rack type")
		return
	}

	o.Count = types.Int64Value(int64(in.Count))
	o.PodType = NewPodTypeObject(ctx, in.RackTypeData, diags)
}

func NewPodInfoMap(ctx context.Context, in *apstra.TemplatePodBasedData, diags *diag.Diagnostics) types.Map {
	rackTypeMap := make(map[apstra.ObjectId]TemplatePodInfo, len(in.RackBasedTemplates))
	for i, apiData := range in.RackBasedTemplates {

		podIdToCount := make(map[apstra.ObjectId]int, len(in.RackBasedTemplateCounts))
		for _, rbtc := range in.RackBasedTemplateCounts {
			podIdToCount[rbtc.RackBasedTemplateId] = rbtc.Count
		}

		podIdToRtd := make(map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo, len(in.RackBasedTemplates))
		for _, rtd := range in.RackBasedTemplates {
			podIdToRtd[rtd.Id] = rtd.Data.RackInfo
		}

		trbri := apstra.TemplateRackBasedRackInfo{
			Count:        podidtocount,
			RackTypeData: podidtortd,
		}

		var tpi TemplatePodInfo
		tpi.LoadApiData(ctx, &apiData, diags)
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: TemplatePodInfo{}.AttrTypes()})
		}
		rackTypeMap[i] = tpi
	}

	result, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: TemplatePodInfo{}.AttrTypes()}, rackTypeMap)
	diags.Append(d...)
	return result
}

func NewPodTypeObject(ctx context.Context, in *apstra.RackTypeData, diags *diag.Diagnostics) types.Object {
	var pt RackType
	pt.Id = types.StringNull()
	pt.LoadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(RackType{}.AttrTypes())
	}

	rtdObj, d := types.ObjectValueFrom(ctx, RackType{}.AttrTypes(), &pt)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(RackType{}.AttrTypes())
	}

	return rtdObj
}
