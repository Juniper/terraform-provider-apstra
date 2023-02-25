package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
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
	LogicalDeviceId    types.String `tfsdk:"logical_device_id"`
	LogicalDevice      types.Object `tfsdk:"logical_device"`
	MlagInfo           types.Object `tfsdk:"mlag_info"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	SpineLinkCount     types.Int64  `tfsdk:"spine_link_count"`
	SpineLinkSpeed     types.String `tfsdk:"spine_link_speed"`
	TagIds             types.Set    `tfsdk:"tag_ids"`
	Tags               types.Set    `tfsdk:"tags"`
}

func (o leafSwitch) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"logical_device_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in data source contexts.",
			Computed:            true,
		},
		"logical_device": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDevice{}.dataSourceAttributesNested(),
		},
		"mlag_info": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Details settings when the Leaf Switch is an MLAG-capable pair.",
			Computed:            true,
			Attributes:          mlagInfo{}.dataSourceAttributes(),
		},
		"redundancy_protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "When set, 'the switch' is actually a LAG-capable redundant pair of the given type.",
			Computed:            true,
		},
		"spine_link_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of links to each spine switch.",
			Computed:            true,
		},
		"spine_link_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Speed of links to spine switches.",
			Computed:            true,
		},
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in data source contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Leaf Switch.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
	}
}

func (o leafSwitch) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Leaf Switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes:          logicalDevice{}.resourceAttributesNested(),
		},
		"mlag_info": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Required when `redundancy_protocol` set to `%s`, "+
				"defines the connectivity between MLAG peers.", goapstra.LeafRedundancyProtocolMlag.String()),
			Optional:   true,
			Attributes: mlagInfo{}.resourceAttributes(),
			Validators: []validator.Object{validateSwitchLagInfo(goapstra.LeafRedundancyProtocolMlag.String())},
		},
		"redundancy_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Enabling a redundancy protocol converts a single "+
				"Leaf Switch into a LAG-capable switch pair. Must be one of '%s'.",
				strings.Join(leafRedundancyModes(), "', '")),
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(leafRedundancyModes()...),
				validateLeafSwitchRedundancyMode(),
				stringFabricConnectivityDesignMustBeWhenValue(goapstra.FabricConnectivityDesignL3Clos, "mlag"),
			},
		},
		"spine_link_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Links per spine.",
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
				int64FabricConnectivityDesignMustBe(goapstra.FabricConnectivityDesignL3Clos),
				int64FabricConnectivityDesignMustBeWhenNull(goapstra.FabricConnectivityDesignL3Collapsed),
			},
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"spine_link_speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of spine-facing links, something like '10G'",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringFabricConnectivityDesignMustBe(goapstra.FabricConnectivityDesignL3Clos),
				stringFabricConnectivityDesignMustBeWhenNull(goapstra.FabricConnectivityDesignL3Collapsed),
			},
		},
		"tag_ids": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Leaf Switch",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Leaf Switch",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: tag{}.resourceAttributesNested(),
			},
		},
	}
}

func (o leafSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id":   types.StringType,
		"logical_device":      types.ObjectType{AttrTypes: logicalDevice{}.attrTypes()},
		"mlag_info":           types.ObjectType{AttrTypes: mlagInfo{}.attrTypes()},
		"redundancy_protocol": types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tags":                types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *leafSwitch) loadApiData(ctx context.Context, in *goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	o.LogicalDeviceId = types.StringNull()
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

	o.TagIds = types.SetNull(types.StringType)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

func newLeafSwitchMap(ctx context.Context, in []goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) types.Map {
	leafSwitches := make(map[string]leafSwitch, len(in))
	for _, leafIn := range in {
		var ls leafSwitch
		ls.loadApiData(ctx, &leafIn, fcd, diags)
		leafSwitches[leafIn.Label] = ls
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: leafSwitch{}.attrTypes()})
		}
	}

	return mapValueOrNull(ctx, types.ObjectType{AttrTypes: leafSwitch{}.attrTypes()}, leafSwitches, diags)
}
