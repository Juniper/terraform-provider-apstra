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

func validateLeafSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	ls := rt.Data.LeafSwitches[i]
	if ls.RedundancyProtocol == goapstra.LeafRedundancyProtocolMlag && ls.MlagInfo == nil {
		diags.AddError("leaf switch MLAG Info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, ls.Label, ls.RedundancyProtocol.String()))
		return
	}
	if ls.LogicalDevice == nil {
		diags.AddError("leaf switch logical device info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' logical device is nil",
				rt.Id, ls.Label))
		return
	}
}

type leafSwitch struct {
	LogicalDevice      types.Object `tfsdk:"logical_device"`
	MlagInfo           types.Object `tfsdk:"mlag_info"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	SpineLinkCount     types.Int64  `tfsdk:"spine_link_count"`
	SpineLinkSpeed     types.String `tfsdk:"spine_link_speed"`
	TagData            types.Set    `tfsdk:"tag_data"`
}

func (o leafSwitch) dataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"spine_link_count": schema.Int64Attribute{
			MarkdownDescription: "Number of links to each spine switch.",
			Computed:            true,
		},
		"spine_link_speed": schema.StringAttribute{
			MarkdownDescription: "Speed of links to spine switches.",
			Computed:            true,
		},
		"redundancy_protocol": schema.StringAttribute{
			MarkdownDescription: "When set, 'the switch' is actually a LAG-capable redundant pair of the given type.",
			Computed:            true,
		},
		"mlag_info": schema.SingleNestedAttribute{
			MarkdownDescription: "Details settings when the Leaf Switch is an MLAG-capable pair.",
			Computed:            true,
			Attributes:          mlagInfo{}.dataSourceAttributes(),
		},
		"logical_device": schema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDevice{}.dataSourceAttributesNested(),
		},
		"tag_data": schema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Leaf Switch.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tagData{}.dataSourceAttributes(),
			},
		},
	}
}

func (o leafSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device":      types.ObjectType{AttrTypes: logicalDevice{}.attrTypes()},
		"mlag_info":           types.ObjectType{AttrTypes: mlagInfo{}.attrTypes()},
		"redundancy_protocol": types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"tag_data":            types.SetType{ElemType: types.ObjectType{AttrTypes: tagData{}.attrTypes()}},
	}
}

func (o *leafSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	o.LogicalDevice = newLogicalDeviceObject(ctx, in.LogicalDevice, diags)

	switch in.RedundancyProtocol {
	case goapstra.LeafRedundancyProtocolMlag:
		o.MlagInfo = newMlagInfoObject(ctx, in.MlagInfo, diags)
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	case goapstra.LeafRedundancyProtocolEsi:
		o.MlagInfo = types.ObjectNull(mlagInfo{}.attrTypes())
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	default:
		o.MlagInfo = types.ObjectNull(mlagInfo{}.attrTypes())
		o.RedundancyProtocol = types.StringNull()
	}

	if fcd == goapstra.FabricConnectivityDesignL3Collapsed {
		o.SpineLinkCount = types.Int64Null()
		o.SpineLinkSpeed = types.StringNull()
	} else {
		o.SpineLinkCount = types.Int64Value(int64(in.LinkPerSpineCount))
		o.SpineLinkSpeed = types.StringValue(string(in.LinkPerSpineSpeed))
	}

	o.TagData = newTagSet(ctx, in.Tags, diags)
}
