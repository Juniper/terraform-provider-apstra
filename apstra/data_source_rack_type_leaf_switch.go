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
	}
	if ls.LogicalDevice == nil {
		diags.AddError("leaf switch logical device info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' logical device is nil",
				rt.Id, ls.Label))
	}
}

type dRackTypeLeafSwitch struct {
	LogicalDeviceData  types.Object `tfsdk:"logical_device"`
	MlagInfo           types.Object `tfsdk:"mlag_info"`
	Name               types.String `tfsdk:"name"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	SpineLinkCount     types.Int64  `tfsdk:"spine_link_count"`
	SpineLinkSpeed     types.String `tfsdk:"spine_link_speed"`
	TagData            types.Set    `tfsdk:"tag_data"`
}

func (o dRackTypeLeafSwitch) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
			Computed:            true,
		},
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
			Attributes:          logicalDeviceData{}.dataSourceAttributes(),
		},
		"tag_data": schema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Leaf Switch.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tagData{}.attributes(),
			},
		},
	}
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
	o.Name = types.StringValue(in.Label)

	if fcd == goapstra.FabricConnectivityDesignL3Collapsed {
		o.SpineLinkCount = types.Int64Null()
		o.SpineLinkSpeed = types.StringNull()
	} else {
		o.SpineLinkCount = types.Int64Value(int64(in.LinkPerSpineCount))
		o.SpineLinkSpeed = types.StringValue(string(in.LinkPerSpineSpeed))
	}

	if in.RedundancyProtocol == goapstra.LeafRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringNull()
	} else {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	o.MlagInfo = newMlagInfoObject(ctx, in.MlagInfo, diags)
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
}
