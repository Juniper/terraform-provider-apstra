package design

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func ValidateLeafSwitch(rt *apstra.RackType, i int, diags *diag.Diagnostics) {
	ls := rt.Data.LeafSwitches[i]
	if ls.RedundancyProtocol == apstra.LeafRedundancyProtocolMlag && ls.MlagInfo == nil {
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

type LeafSwitch struct {
	LogicalDeviceId    types.String `tfsdk:"logical_device_id"`
	LogicalDevice      types.Object `tfsdk:"logical_device"`
	MlagInfo           types.Object `tfsdk:"mlag_info"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	SpineLinkCount     types.Int64  `tfsdk:"spine_link_count"`
	SpineLinkSpeed     types.String `tfsdk:"spine_link_speed"`
	TagIds             types.Set    `tfsdk:"tag_ids"`
	Tags               types.Set    `tfsdk:"tags"`
}

func (o LeafSwitch) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"logical_device_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in data source contexts.",
			Computed:            true,
		},
		"logical_device": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          LogicalDevice{}.DataSourceAttributesNested(),
		},
		"mlag_info": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Details settings when the Leaf Switch is an MLAG-capable pair.",
			Computed:            true,
			Attributes:          MlagInfo{}.DataSourceAttributes(),
		},
		"redundancy_protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "When set, 'the switch' is actually a LAG-capable redundant pair of the given type.",
			Computed:            true,
		},
		"spine_link_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of links to each Spine switch.",
			Computed:            true,
		},
		"spine_link_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Speed of links to Spine switches.",
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
				Attributes: Tag{}.DataSourceAttributesNested(),
			},
		},
	}
}

func (o LeafSwitch) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Leaf Switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          LogicalDevice{}.ResourceAttributesNested(),
		},
		"mlag_info": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Required when `redundancy_protocol` set to `%s`, "+
				"defines the connectivity between MLAG peers.", apstra.LeafRedundancyProtocolMlag.String()),
			Optional:   true,
			Attributes: MlagInfo{}.ResourceAttributes(),
			Validators: []validator.Object{
				apstravalidator.RequiredWhenValueIs(path.MatchRelative().AtParent().AtName("redundancy_protocol"), types.StringValue(apstra.LeafRedundancyProtocolMlag.String())),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("redundancy_protocol"), types.StringValue(apstra.LeafRedundancyProtocolEsi.String())),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("redundancy_protocol"), types.StringNull()),
			},
		},
		"redundancy_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Enabling a redundancy protocol converts a single "+
				"Leaf Switch into a LAG-capable switch pair. Must be one of '%s'.",
				strings.Join(LeafRedundancyModes(), "', '")),
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(LeafRedundancyModes()...),
				apstravalidator.StringFabricConnectivityDesignMustBeWhenValue(apstra.FabricConnectivityDesignL3Clos, "mlag"),
			},
		},
		"spine_link_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Links per Spine.",
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
				apstravalidator.Int64FabricConnectivityDesignMustBe(apstra.FabricConnectivityDesignL3Clos),
				apstravalidator.Int64FabricConnectivityDesignMustBeWhenNull(apstra.FabricConnectivityDesignL3Collapsed),
			},
			Optional: true,
			Computed: true,
			// Default: ** do not default this attribute because L3Collapsed designs can't use it **
		},
		"spine_link_speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of Spine-facing links, something like '10G'",
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ParseSpeed(),
				apstravalidator.StringFabricConnectivityDesignMustBe(apstra.FabricConnectivityDesignL3Clos),
				apstravalidator.StringFabricConnectivityDesignMustBeWhenNull(apstra.FabricConnectivityDesignL3Collapsed),
			},
		},
		"tag_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag IDs to be applied to this Leaf Switch",
			Optional:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			ElementType:         types.StringType,
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Leaf Switch",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: Tag{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o LeafSwitch) ResourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in nested contexts.",
			Computed:            true,
		},
		"logical_device": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			Attributes:          LogicalDevice{}.ResourceAttributesNested(),
		},
		"mlag_info": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Defines connectivity between MLAG peers when "+
				"`redundancy_protocol` is set to `%s`.", apstra.LeafRedundancyProtocolMlag.String()),
			Computed:   true,
			Attributes: MlagInfo{}.ResourceAttributes(),
		},
		"redundancy_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Enabling a redundancy protocol converts a single "+
				"Leaf Switch into a LAG-capable switch pair. Must be one of '%s'.",
				strings.Join(LeafRedundancyModes(), "', '")),
			Computed: true,
		},
		"spine_link_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Links per Spine.",
			Computed:            true,
		},
		"spine_link_speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of Spine-facing links, something like '10G'",
			Computed:            true,
		},
		"tag_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in nested contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Leaf Switch",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: Tag{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o LeafSwitch) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id":   types.StringType,
		"logical_device":      types.ObjectType{AttrTypes: LogicalDevice{}.AttrTypes()},
		"mlag_info":           types.ObjectType{AttrTypes: MlagInfo{}.AttrTypes()},
		"redundancy_protocol": types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tags":                types.SetType{ElemType: types.ObjectType{AttrTypes: Tag{}.AttrTypes()}},
	}
}

func (o *LeafSwitch) Request(ctx context.Context, path path.Path, fcd apstra.FabricConnectivityDesign, diags *diag.Diagnostics) *apstra.RackElementLeafSwitchRequest {
	var linkPerSpineCount int
	if o.SpineLinkCount.IsUnknown() && fcd == apstra.FabricConnectivityDesignL3Clos {
		// config omits 'spine_link_count' set default value (1) for fabric designs which require it
		linkPerSpineCount = 1
	} else {
		// config includes 'spine_link_count' -- use the configured value
		linkPerSpineCount = int(o.SpineLinkCount.ValueInt64())
	}

	var linkPerSpineSpeed apstra.LogicalDevicePortSpeed
	if !o.SpineLinkSpeed.IsNull() {
		linkPerSpineSpeed = apstra.LogicalDevicePortSpeed(o.SpineLinkSpeed.ValueString())
	}

	redundancyProtocol := apstra.LeafRedundancyProtocolNone
	if !o.RedundancyProtocol.IsNull() {
		err := redundancyProtocol.FromString(o.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path.AtMapKey("redundancy_protocol"),
				fmt.Sprintf("error parsing redundancy_protocol '%s'",
					o.RedundancyProtocol.ValueString()),
				err.Error())
			return nil
		}
	}

	var leafMlagInfo *apstra.LeafMlagInfo
	if !o.MlagInfo.IsNull() {
		mi := MlagInfo{}
		d := o.MlagInfo.As(ctx, &mi, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}
		leafMlagInfo = mi.Request(ctx, diags)
	}

	tagIds := make([]apstra.ObjectId, len(o.TagIds.Elements()))
	o.TagIds.ElementsAs(ctx, &tagIds, false)

	return &apstra.RackElementLeafSwitchRequest{
		MlagInfo:           leafMlagInfo,
		LinkPerSpineCount:  linkPerSpineCount,
		LinkPerSpineSpeed:  linkPerSpineSpeed,
		RedundancyProtocol: redundancyProtocol,
		LogicalDeviceId:    apstra.ObjectId(o.LogicalDeviceId.ValueString()),
		Tags:               tagIds,
	}
}

func (o *LeafSwitch) LoadApiData(ctx context.Context, in *apstra.RackElementLeafSwitch, fcd apstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	o.LogicalDeviceId = types.StringNull()
	o.LogicalDevice = NewLogicalDeviceObject(ctx, in.LogicalDevice, diags)

	switch in.RedundancyProtocol {
	case apstra.LeafRedundancyProtocolMlag:
		o.MlagInfo = NewMlagInfoObject(ctx, in.MlagInfo, diags)
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	case apstra.LeafRedundancyProtocolEsi:
		o.MlagInfo = types.ObjectNull(MlagInfo{}.AttrTypes())
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	default:
		o.MlagInfo = types.ObjectNull(MlagInfo{}.AttrTypes())
		o.RedundancyProtocol = types.StringNull()
	}

	if fcd == apstra.FabricConnectivityDesignL3Collapsed {
		o.SpineLinkCount = types.Int64Null()
		o.SpineLinkSpeed = types.StringNull()
	} else {
		o.SpineLinkCount = types.Int64Value(int64(in.LinkPerSpineCount))
		o.SpineLinkSpeed = types.StringValue(string(in.LinkPerSpineSpeed))
	}

	o.TagIds = types.SetNull(types.StringType)
	o.Tags = NewTagSet(ctx, in.Tags, diags)
}

func (o *LeafSwitch) CopyWriteOnlyElements(ctx context.Context, src *LeafSwitch, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "LeafSwitch.CopyWriteOnlyElements: attempt to copy from nil source")
		return
	}

	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	o.TagIds = utils.SetValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)
}

func NewLeafSwitchMap(ctx context.Context, in []apstra.RackElementLeafSwitch, fcd apstra.FabricConnectivityDesign, diags *diag.Diagnostics) types.Map {
	leafSwitches := make(map[string]LeafSwitch, len(in))
	for _, leafIn := range in {
		var ls LeafSwitch
		ls.LoadApiData(ctx, &leafIn, fcd, diags)
		leafSwitches[leafIn.Label] = ls
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: LeafSwitch{}.AttrTypes()})
		}
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: LeafSwitch{}.AttrTypes()}, leafSwitches, diags)
}

// LeafRedundancyModes returns permitted fabric_connectivity_design mode strings
func LeafRedundancyModes() []string {
	return []string{
		apstra.LeafRedundancyProtocolEsi.String(),
		apstra.LeafRedundancyProtocolMlag.String(),
	}
}
