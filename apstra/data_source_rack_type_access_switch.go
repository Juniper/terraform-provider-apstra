package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateAccessSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	as := rt.Data.AccessSwitches[i]
	if as.RedundancyProtocol == goapstra.AccessRedundancyProtocolEsi && as.EsiLagInfo == nil {
		diags.AddError("access switch ESI LAG Info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, as.Label, as.RedundancyProtocol.String()))
	}
	if as.LogicalDevice == nil {
		diags.AddError("access switch logical device info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' logical device is nil",
				rt.Id, as.Label))
	}
}

type dRackTypeAccessSwitch struct {
	Count              types.Int64  `tfsdk:"count"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	EsiLagInfo         types.Object `tfsdk:"esi_lag_info"`
	LogicalDeviceData  types.Object `tfsdk:"logical_device"`
	TagData            types.Set    `tfsdk:"tag_data"`
	Links              types.Set    `tfsdk:"links"`
}

func (o dRackTypeAccessSwitch) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"count": schema.Int64Attribute{
			MarkdownDescription: "Count of Access Switches of this type.",
			Computed:            true,
		},
		"redundancy_protocol": schema.StringAttribute{
			MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
			Computed:            true,
		},
		"esi_lag_info": esiLagInfo{}.schemaAsDataSource(),
		"logical_device": schema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDeviceData{}.dataSourceAttributes(),
		},
		"tag_data": schema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Access Switch.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tagData{}.dataSourceAttributes(),
			},
		},
		"links": schema.SetNestedAttribute{
			MarkdownDescription: "Details links from this Access Switch to upstream switches within this Rack Type.",
			Computed:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: schema.NestedAttributeObject{
				Attributes: dRackLink{}.attributes(),
			},
		},
	}
}

func (o dRackTypeAccessSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"count":               types.Int64Type,
		"redundancy_protocol": types.StringType,
		"esi_lag_info":        esiLagInfo{}.attrType(),
		"logical_device":      logicalDeviceData{}.attrType(),
		"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
		"links":               types.SetType{ElemType: dRackLink{}.attrType()},
	}
}

func (o dRackTypeAccessSwitch) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *dRackTypeAccessSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
	o.Count = types.Int64Value(int64(in.InstanceCount))

	if in.RedundancyProtocol == goapstra.AccessRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringNull()
	} else {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	o.EsiLagInfo = newEsiLagInfo(ctx, in.EsiLagInfo, diags)
	if diags.HasError() {
		return
	}

	o.TagData = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}

	o.LogicalDeviceData = newLogicalDeviceDataObject(ctx, in.LogicalDevice, diags)
	if diags.HasError() {
		return
	}

	o.Links = newDataSourceLinkSet(ctx, in.Links, diags)
	if diags.HasError() {
		return
	}
}
