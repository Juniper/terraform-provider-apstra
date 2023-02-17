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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rRackLink struct {
	Name             types.String `tfsdk:"name"`
	TargetSwitchName types.String `tfsdk:"target_switch_name"`
	LagMode          types.String `tfsdk:"lag_mode"`
	LinksPerSwitch   types.Int64  `tfsdk:"links_per_switch"`
	Speed            types.String `tfsdk:"speed"`
	SwitchPeer       types.String `tfsdk:"switch_peer"`
	TagData          types.Set    `tfsdk:"tag_data"`
	TagIds           types.Set    `tfsdk:"tag_ids"`
}

func (o rRackLink) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of this link, copied from map key.",
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"target_switch_name": schema.StringAttribute{
			MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"lag_mode": schema.StringAttribute{
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
		"links_per_switch": schema.Int64Attribute{
			MarkdownDescription: "Number of Links to each switch.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"speed": schema.StringAttribute{
			MarkdownDescription: "Speed of this Link.",
			Required:            true,
		},
		"switch_peer": schema.StringAttribute{
			MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				goapstra.RackLinkSwitchPeerFirst.String(),
				goapstra.RackLinkSwitchPeerSecond.String(),
			)},
		},
		"tag_ids": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Set of Tag IDs to be applied to this Link",
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"tag_data": schema.SetNestedAttribute{
			MarkdownDescription: "Set of Tags (Name + Description) applied to this Link",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tagData{}.resourceAttributes(),
			},
		},
	}
}

func (o rRackLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"target_switch_name": types.StringType,
		"lag_mode":           types.StringType,
		"links_per_switch":   types.Int64Type,
		"speed":              types.StringType,
		"switch_peer":        types.StringType,
		"tag_ids":            types.SetType{ElemType: types.StringType},
		"tag_data":           types.SetType{ElemType: types.ObjectType{AttrTypes: tagData{}.attrTypes()}},
	}
}

func (o *rRackLink) loadApiResponse(ctx context.Context, in *goapstra.RackLink, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.TargetSwitchName = types.StringValue(in.TargetSwitchLabel)
	o.LinksPerSwitch = types.Int64Value(int64(in.LinkPerSwitchCount))
	o.Speed = types.StringValue(string(in.LinkSpeed))

	if in.LagMode == goapstra.RackLinkLagModeNone {
		o.LagMode = types.StringNull()
	} else {
		o.LagMode = types.StringValue(in.LagMode.String())
	}

	if in.SwitchPeer == goapstra.RackLinkSwitchPeerNone {
		o.SwitchPeer = types.StringNull()
	} else {
		o.SwitchPeer = types.StringValue(in.SwitchPeer.String())
	}

	// null set for now to avoid nil pointer dereference error because the API
	// response doesn't contain the tag IDs. See copyWriteOnlyElements() method.
	o.TagIds = types.SetNull(types.StringType)

	o.TagData = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}
}

func (o *rRackLink) copyWriteOnlyElements(ctx context.Context, src *rRackLink, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddError(errProviderBug, "rRackLink.copyWriteOnlyElements: attempt to copy from nil source")
		return
	}
	o.TagIds = setValueOrNull(ctx, types.StringType, src.TagIds.Elements(), diags)

}

func (o *rRackLink) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackLinkRequest {
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
			fmt.Sprintf("target switch %q not found in rack type %q", o.TargetSwitchName, rack.Id))
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
		Label:              o.Name.ValueString(),
		Tags:               tagIds,
		LinkPerSwitchCount: linksPerSwitch,
		LinkSpeed:          goapstra.LogicalDevicePortSpeed(o.Speed.ValueString()),
		TargetSwitchLabel:  o.TargetSwitchName.ValueString(),
		AttachmentType:     o.linkAttachmentType(upstreamRedundancyProtocol, diags),
		LagMode:            lagMode,
		SwitchPeer:         switchPeer,
	}
}

func newResourceLinkMap(ctx context.Context, in []goapstra.RackLink, diags *diag.Diagnostics) types.Map {
	if len(in) == 0 {
		return types.MapNull(types.ObjectType{AttrTypes: rRackLink{}.attrTypes()})
	}

	links := make(map[string]rRackLink, len(in))
	for i := range in {
		var link rRackLink
		link.loadApiResponse(ctx, &in[i], diags)
		links[in[i].Label] = link
	}

	result, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: rRackLink{}.attrTypes()}, &links)
	diags.Append(d...)

	return result
}

func (o *rRackLink) linkAttachmentType(upstreamRedundancyMode fmt.Stringer, _ *diag.Diagnostics) goapstra.RackLinkAttachmentType {
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
