package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	//TagData          types.Set    `tfsdk:"tag_data"`
	//TagIds           types.Set    `tfsdk:"tag_ids"`
}

func (o rRackLink) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of this link.",
			Required:            true,
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
				goapstra.RackLinkLagModeStatic.String())},
		},
		"links_per_switch": schema.Int64Attribute{
			MarkdownDescription: "Number of Links to each switch.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			Validators:          []validator.Int64{int64validator.AtLeast(2)},
		},
		"speed": schema.StringAttribute{
			MarkdownDescription: "Speed of this Link.",
			Required:            true,
		},
		"switch_peer": schema.StringAttribute{
			MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
			Optional:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				goapstra.RackLinkSwitchPeerFirst.String(),
				goapstra.RackLinkSwitchPeerSecond.String(),
			)},
		},
		//			"tag_ids":  tagIdsAttributeSchema(),
		//			"tag_data": tagsDataAttributeSchema(),
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
		//"tag_ids":            types.SetType{ElemType: types.StringType},
		//"tag_data":           types.SetType{ElemType: tagData{}.attrType()},
	}
}

func (o rRackLink) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
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

	// o.Tags = types.SetNull() // fill in later with copyWriteOnlyAttributes
	//o.TagData = newTagSet(ctx, in.Tags, diags)
	//if diags.HasError() {
	//	return
	//}
}

func (o *rRackLink) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackLinkRequest {
	var err error

	//tags := make([]goapstra.ObjectId, len(o.TagIds))
	//for i, tag := range o.TagIds {
	//	tags[i] = goapstra.ObjectId(tag)
	//}

	lagMode := goapstra.RackLinkLagModeNone
	if !o.LagMode.IsNull() {
		err = lagMode.FromString(o.LagMode.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing lag_mode", err.Error())
			return nil
		}
	}

	switchPeer := goapstra.RackLinkSwitchPeerNone
	if !o.SwitchPeer.IsNull() {
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
		Label: o.Name.ValueString(),
		//Tags:               tags,
		LinkPerSwitchCount: linksPerSwitch,
		LinkSpeed:          goapstra.LogicalDevicePortSpeed(o.Speed.ValueString()),
		TargetSwitchLabel:  o.TargetSwitchName.ValueString(),
		AttachmentType:     o.linkAttachmentType(upstreamRedundancyProtocol),
		LagMode:            lagMode,
		SwitchPeer:         switchPeer,
	}
}

func newResourceLinkMap(ctx context.Context, in []goapstra.RackLink, diags *diag.Diagnostics) types.Map {
	if len(in) == 0 {
		return types.MapNull(rRackLink{}.attrType())
	}

	links := make(map[string]rRackLink, len(in))
	for i := range in {
		var link rRackLink
		link.loadApiResponse(ctx, &in[i], diags)
		links[in[i].Label] = link
	}

	result, d := types.MapValueFrom(ctx, rRackLink{}.attrType(), &links)
	diags.Append(d...)

	return result
}
