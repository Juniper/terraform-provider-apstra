package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateLeafSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	ls := rt.Data.LeafSwitches[i]
	if ls.RedundancyProtocol == goapstra.LeafRedundancyProtocolMlag && ls.MlagInfo == nil {
		diags.AddError("leaf switch MLAG Info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, ls.Label, ls.RedundancyProtocol.String()))
	}
	if ls.LogicalDevice == nil {
		diags.AddError("leaf switch logical device info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' logical device is nil",
				rt.Id, ls.Label))
	}
}

type dRackTypeLeafSwitch struct {
	Name               types.String `tfsdk:"name"`
	SpineLinkCount     types.Int64  `tfsdk:"spine_link_count"`
	SpineLinkSpeed     types.String `tfsdk:"spine_link_speed"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	MlagInfo           types.Object `tfsdk:"mlag_info"`
	LogicalDevice      types.Object `tfsdk:"logical_device"`
	TagData            types.Set    `tfsdk:"tag_data"`
}

func (o dRackTypeLeafSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"mlag_info":           mlagInfo{}.attrType(),
		"logical_device":      logicalDeviceData{}.attrType(),
		"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
	}
}

func (o dRackTypeLeafSwitch) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *dRackTypeLeafSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	var d diag.Diagnostics

	o.Name = types.StringValue(in.Label)
	switch fcd {
	case goapstra.FabricConnectivityDesignL3Collapsed:
		o.SpineLinkCount = types.Int64Null()
		o.SpineLinkSpeed = types.StringNull()
	case goapstra.FabricConnectivityDesignL3Clos:
		o.SpineLinkCount = types.Int64Value(int64(in.LinkPerSpineCount))
		o.SpineLinkSpeed = types.StringValue(string(in.LinkPerSpineSpeed))
	}

	if in.RedundancyProtocol == goapstra.LeafRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringNull()
	} else {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	if in.MlagInfo != nil && in.MlagInfo.LeafLeafLinkCount > 0 {
		var mlagInfo mlagInfo
		mlagInfo.loadApiResponse(ctx, in.MlagInfo, diags)
		if diags.HasError() {
			return
		}
		o.MlagInfo, d = types.ObjectValueFrom(ctx, mlagInfo.attrTypes(), &mlagInfo)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		o.MlagInfo = types.ObjectNull(mlagInfo{}.attrTypes())
	}

	o.TagData = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}

	var logicalDevice logicalDeviceData
	logicalDevice.loadApiResponse(ctx, in.LogicalDevice, diags)
	if diags.HasError() {
		return
	}
	o.LogicalDevice, d = types.ObjectValueFrom(ctx, logicalDevice.attrTypes(), &logicalDevice)
}
