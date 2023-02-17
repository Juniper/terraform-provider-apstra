package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"strings"
)

type rRackTypeLeafSwitch struct {
	LogicalDeviceData  types.Object `tfsdk:"logical_device"`
	LogicalDeviceId    types.String `tfsdk:"logical_device_id"`
	MlagInfo           types.Object `tfsdk:"mlag_info"`
	Name               types.String `tfsdk:"name"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	SpineLinkCount     types.Int64  `tfsdk:"spine_link_count"`
	SpineLinkSpeed     types.String `tfsdk:"spine_link_speed"`
	TagIds             types.Set    `tfsdk:"tag_ids"`
	TagData            types.Set    `tfsdk:"tag_data"`
}

func (o rRackTypeLeafSwitch) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Switch name, copied from map key, used when creating intra-rack links targeting this switch.",
			Computed:            true,
		},
		"logical_device_id": schema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Logical Device used to model this Leaf Switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"spine_link_count": schema.Int64Attribute{
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
		"spine_link_speed": schema.StringAttribute{
			MarkdownDescription: "Speed of spine-facing links, something like '10G'",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringFabricConnectivityDesignMustBe(goapstra.FabricConnectivityDesignL3Clos),
				stringFabricConnectivityDesignMustBeWhenNull(goapstra.FabricConnectivityDesignL3Collapsed),
			},
		},
		"redundancy_protocol": schema.StringAttribute{
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
		"logical_device": schema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes cloned from the Global Catalog at creation time.",
			Computed:            true,
			PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes:          logicalDeviceData{}.schemaAsResourceReadOnly(),
		},
		"mlag_info": schema.SingleNestedAttribute{
			MarkdownDescription: fmt.Sprintf("Required when `redundancy_protocol` set to `%s`, "+
				"defines the connectivity between MLAG peers.", goapstra.LeafRedundancyProtocolMlag.String()),
			Optional:   true,
			Attributes: mlagInfo{}.resourceAttributes(),
			Validators: []validator.Object{validateSwitchLagInfo(goapstra.LeafRedundancyProtocolMlag.String())},
		},
		"tag_ids": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Leaf Switch",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tag_data": schema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Leaf Switch",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tagData{}.resourceAttributes(),
			},
		},
	}
}

func (o rRackTypeLeafSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"logical_device_id":   types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"logical_device":      types.ObjectType{AttrTypes: logicalDeviceData{}.attrTypes()},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tag_data":            types.SetType{ElemType: types.ObjectType{AttrTypes: tagData{}.attrTypes()}},
		"mlag_info":           types.ObjectType{AttrTypes: mlagInfo{}.attrTypes()},
	}
}

func (o *rRackTypeLeafSwitch) copyWriteOnlyElements(ctx context.Context, src *rRackTypeLeafSwitch, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "rRackTypeLeafSwitch.copyWriteOnlyElements: attempt to copy from nil source")
		return
	}

	o.LogicalDeviceId = types.StringValue(src.LogicalDeviceId.ValueString())
	o.TagIds = setValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)
}

func (o *rRackTypeLeafSwitch) request(ctx context.Context, path path.Path, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) *goapstra.RackElementLeafSwitchRequest {
	var linkPerSpineCount int
	if o.SpineLinkCount.IsUnknown() && fcd == goapstra.FabricConnectivityDesignL3Clos {
		// config omits 'spine_link_count' set default value (1) for fabric designs which require it
		linkPerSpineCount = 1
	} else {
		// config includes 'spine_link_count' -- use the configured value
		linkPerSpineCount = int(o.SpineLinkCount.ValueInt64())
	}

	var linkPerSpineSpeed goapstra.LogicalDevicePortSpeed
	if !o.SpineLinkSpeed.IsNull() {
		linkPerSpineSpeed = goapstra.LogicalDevicePortSpeed(o.SpineLinkSpeed.ValueString())
	}

	redundancyProtocol := goapstra.LeafRedundancyProtocolNone
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

	var leafMlagInfo *goapstra.LeafMlagInfo
	if !o.MlagInfo.IsNull() {
		mi := mlagInfo{}
		d := o.MlagInfo.As(ctx, &mi, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}
		leafMlagInfo = mi.request(ctx, diags)
	}

	tagIds := make([]goapstra.ObjectId, len(o.TagIds.Elements()))
	o.TagIds.ElementsAs(ctx, &tagIds, false)

	return &goapstra.RackElementLeafSwitchRequest{
		Label:              o.Name.ValueString(),
		MlagInfo:           leafMlagInfo,
		LinkPerSpineCount:  linkPerSpineCount,
		LinkPerSpineSpeed:  linkPerSpineSpeed,
		RedundancyProtocol: redundancyProtocol,
		LogicalDeviceId:    goapstra.ObjectId(o.LogicalDeviceId.ValueString()),
		Tags:               tagIds,
	}
}

func (o *rRackTypeLeafSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	if fcd != goapstra.FabricConnectivityDesignL3Collapsed {
		o.SpineLinkCount = types.Int64Value(int64(in.LinkPerSpineCount))
		o.SpineLinkSpeed = types.StringValue(string(in.LinkPerSpineSpeed))
	}

	switch in.RedundancyProtocol {
	case goapstra.LeafRedundancyProtocolNone:
		o.RedundancyProtocol = types.StringNull()
		o.MlagInfo = types.ObjectNull(mlagInfo{}.attrTypes())
	case goapstra.LeafRedundancyProtocolEsi:
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
		o.MlagInfo = types.ObjectNull(mlagInfo{}.attrTypes())
	case goapstra.LeafRedundancyProtocolMlag:
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
		o.MlagInfo = newMlagInfoObject(ctx, in.MlagInfo, diags)
	}
	if diags.HasError() {
		return
	}

	// null set for now to avoid nil pointer dereference error because the API
	// response doesn't contain the tag IDs. See copyWriteOnlyElements() method.
	o.TagIds = types.SetNull(types.StringType)

	o.TagData = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}

	o.LogicalDeviceData = newLogicalDeviceDataObject(ctx, in.LogicalDevice, diags)
	if diags.HasError() {
		return
	}
}

// leafRedundancyModes returns permitted fabric_connectivity_design mode strings
func leafRedundancyModes() []string {
	return []string{
		goapstra.LeafRedundancyProtocolEsi.String(),
		goapstra.LeafRedundancyProtocolMlag.String()}
}
