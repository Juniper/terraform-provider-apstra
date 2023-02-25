package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type templateRackType struct {
	Count        types.Int64  `tfsdk:"count"`
	RackTypeData types.Object `tfsdk:"rack_type_data"`
}

func (o templateRackType) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"count": schema.Int64Attribute{
			MarkdownDescription: "Number of instances of this Rack Type.",
			Computed:            true,
		},
		"rack_type_data": schema.SingleNestedAttribute{
			MarkdownDescription: "Rack Type details.",
			Computed:            true,
			Attributes:          rackTypeData{}.dataSourceAttributes(),
		},
	}
}

func (o templateRackType) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"count":          types.Int64Type,
		"rack_type_data": types.ObjectType{AttrTypes: rackTypeData{}.attrTypes()},
	}
}

func (o *templateRackType) loadApiResponse(ctx context.Context, in *goapstra.TemplateRackBasedData, id goapstra.ObjectId, diags *diag.Diagnostics) {
	count, rt := in.GetRackTypeCount(id)
	if count == 0 {
		diags.AddError(errProviderBug, fmt.Sprintf("%d instances of Rack Type %q found in template.", count, id))
		return
	}
	if rt == nil {
		diags.AddError(errProviderBug, fmt.Sprintf("Rack Type %q in template is nil.", id))
		return
	}

	o.Count = types.Int64Value(int64(count))
	o.RackTypeData = newRackTypeDataObject(ctx, rt.Data, diags)
}

func newDesignTemplateRackTypeMap(ctx context.Context, in *goapstra.TemplateRackBasedData, diags *diag.Diagnostics) types.Map {
	rackTypes := make(map[string]templateRackType, len(in.RackTypeCounts))
	for _, rtc := range in.RackTypeCounts {
		var dtrt templateRackType
		dtrt.loadApiResponse(ctx, in, rtc.RackTypeId, diags)
		rackTypes[string(rtc.RackTypeId)] = dtrt
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: templateRackType{}.attrTypes()})
		}
	}

	result, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: templateRackType{}.attrTypes()}, rackTypes)
	diags.Append(d...)
	return result
}
