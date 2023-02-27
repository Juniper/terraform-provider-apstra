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
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rackLink struct {
	TargetSwitchName types.String `tfsdk:"target_switch_name"`
	LagMode          types.String `tfsdk:"lag_mode"`
	LinksPerSwitch   types.Int64  `tfsdk:"links_per_switch"`
	Speed            types.String `tfsdk:"speed"`
	SwitchPeer       types.String `tfsdk:"switch_peer"`
	TagIds           types.Set    `tfsdk:"tag_ids"`
	Tags             types.Set    `tfsdk:"tags"`
}

func (o rackLink) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"target_switch_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
			Computed:            true,
		},
		"lag_mode": dataSourceSchema.StringAttribute{
			MarkdownDescription: "LAG negotiation mode of the Link.",
			Computed:            true,
		},
		"links_per_switch": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Links to each switch.",
			Computed:            true,
		},
		"speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Speed of this Link.",
			Computed:            true,
		},
		"switch_peer": dataSourceSchema.StringAttribute{
			MarkdownDescription: "For non-LAG connections to redundant switch pairs, this field selects the target switch.",
			Computed:            true,
		},
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in data source contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Link.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
	}
}

func (o rackLink) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"target_switch_name": resourceSchema.StringAttribute{
			MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"lag_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "LAG negotiation mode of the Link.",
			Computed:            true,
			Optional:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Validators: []validator.String{stringvalidator.OneOf(
				goapstra.RackLinkLagModeActive.String(),
				goapstra.RackLinkLagModePassive.String(),
				goapstra.RackLinkLagModeStatic.String(),
			)},
		},
		"links_per_switch": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of Links to each switch.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of this Link.",
			Required:            true,
		},
		"switch_peer": resourceSchema.StringAttribute{
			MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				goapstra.RackLinkSwitchPeerFirst.String(),
				goapstra.RackLinkSwitchPeerSecond.String(),
			)},
		},
		"tag_ids": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Link",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tags": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Link",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: tag{}.resourceAttributes(),
			},
		},
	}
}

func (o rackLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"target_switch_name": types.StringType,
		"lag_mode":           types.StringType,
		"links_per_switch":   types.Int64Type,
		"speed":              types.StringType,
		"switch_peer":        types.StringType,
		"tag_ids":            types.SetType{ElemType: types.StringType},
		"tags":               types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *rackLink) request(ctx context.Context, path path.Path, rack *rackType, diags *diag.Diagnostics) *goapstra.RackLinkRequest {
	var err error

	tagIds := make([]goapstra.ObjectId, len(o.TagIds.Elements()))
	o.TagIds.ElementsAs(ctx, &tagIds, false)

	lagMode := goapstra.RackLinkLagModeNone
	if !o.LagMode.IsNull() {
		err = lagMode.FromString(o.LagMode.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing lag_mode", err.Error())
			return nil
		}
	}

	switchPeer := goapstra.RackLinkSwitchPeerNone
	if !o.SwitchPeer.IsNull() && !o.SwitchPeer.IsUnknown() {
		err = switchPeer.FromString(o.SwitchPeer.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing switch_peer", err.Error())
			return nil
		}
	}

	leaf := rack.leafSwitchByName(ctx, o.TargetSwitchName.ValueString(), diags)
	access := rack.accessSwitchByName(ctx, o.TargetSwitchName.ValueString(), diags)
	if leaf == nil && access == nil {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("target switch %q not found in rack type %q", o.TargetSwitchName.ValueString(), rack.Id))
		return nil
	}
	if leaf != nil && access != nil {
		diags.AddError(errProviderBug, "link seems to be attached to both leaf and access switches")
		return nil
	}

	upstreamRedundancyProtocol := rack.getSwitchRedundancyProtocolByName(ctx, o.TargetSwitchName.ValueString(), path, diags)
	if diags.HasError() {
		return nil
	}

	linksPerSwitch := 1
	if !o.LinksPerSwitch.IsNull() {
		linksPerSwitch = int(o.LinksPerSwitch.ValueInt64())
	}

	return &goapstra.RackLinkRequest{
		Tags:               tagIds,
		LinkPerSwitchCount: linksPerSwitch,
		LinkSpeed:          goapstra.LogicalDevicePortSpeed(o.Speed.ValueString()),
		TargetSwitchLabel:  o.TargetSwitchName.ValueString(),
		AttachmentType:     o.linkAttachmentType(upstreamRedundancyProtocol, diags),
		LagMode:            lagMode,
		SwitchPeer:         switchPeer,
	}
}

func (o *rackLink) loadApiData(ctx context.Context, in *goapstra.RackLink, diags *diag.Diagnostics) {
	o.TargetSwitchName = types.StringValue(in.TargetSwitchLabel)
	o.LinksPerSwitch = types.Int64Value(int64(in.LinkPerSwitchCount))
	o.Speed = types.StringValue(string(in.LinkSpeed))
	o.LagMode = stringValueWithNull(ctx, in.LagMode.String(), goapstra.RackLinkLagModeNone.String(), diags)
	o.SwitchPeer = stringValueWithNull(ctx, in.SwitchPeer.String(), goapstra.RackLinkSwitchPeerNone.String(), diags)
	o.TagIds = types.SetNull(types.StringType)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

func (o *rackLink) copyWriteOnlyElements(ctx context.Context, src *rackLink, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "rackLink.copyWriteOnlyElements: attempt to copy from nil source")
		return
	}
	o.TagIds = setValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)
}

func (o *rackLink) linkAttachmentType(upstreamRedundancyMode fmt.Stringer, _ *diag.Diagnostics) goapstra.RackLinkAttachmentType {
	switch upstreamRedundancyMode.String() {
	case goapstra.LeafRedundancyProtocolNone.String():
		return goapstra.RackLinkAttachmentTypeSingle
	case goapstra.AccessRedundancyProtocolNone.String():
		return goapstra.RackLinkAttachmentTypeSingle
	}

	if o.LagMode.IsNull() {
		return goapstra.RackLinkAttachmentTypeSingle
	}

	if !o.SwitchPeer.IsNull() && !o.SwitchPeer.IsUnknown() {
		return goapstra.RackLinkAttachmentTypeSingle
	}

	switch o.LagMode.ValueString() {
	case goapstra.RackLinkLagModeActive.String():
		return goapstra.RackLinkAttachmentTypeDual
	case goapstra.RackLinkLagModePassive.String():
		return goapstra.RackLinkAttachmentTypeDual
	case goapstra.RackLinkLagModeStatic.String():
		return goapstra.RackLinkAttachmentTypeDual
	}
	return goapstra.RackLinkAttachmentTypeSingle
}

func newLinkMap(ctx context.Context, in []goapstra.RackLink, diags *diag.Diagnostics) types.Map {
	links := make(map[string]rackLink, len(in))
	for _, link := range in {
		var l rackLink
		l.loadApiData(ctx, &link, diags)
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: rackLink{}.attrTypes()})
		}
		links[link.Label] = l
	}

	return mapValueOrNull(ctx, types.ObjectType{AttrTypes: rackLink{}.attrTypes()}, links, diags)
}
