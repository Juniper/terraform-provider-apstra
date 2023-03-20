package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datacenterRoutingPolicyExport struct {
	superspine types.Bool `tfsdk:"spine_superspine_links"`
	spine      types.Bool `tfsdk:"spine_leaf_links"`
	l3Edge     types.Bool `tfsdk:"l3_edge_server_links"`
	l2Edge     types.Bool `tfsdk:"l2_edge_subnet_links"`
	loopback   types.Bool `tfsdk:"loopbacks"`
	static     types.Bool `tfsdk:"static_routes"`
}

func (o datacenterRoutingPolicyExport) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"spine_superspine_links": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all spine-leaf (fabric) links within a VRF. EVPN routing zones do not have " +
				"spine-leaf addressing, so this generated list may be empty. For routing zones of type Virtual L3 " +
				"Fabric, subinterfaces between spine-leaf will be included.",
			Optional: true,
			Computed: true,
		},
		"spine_leaf_links": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all spine-supersine (fabric) links within the default routing zone (VRF)",
			Optional:            true,
			Computed:            true,
		},
		"l3_edge_server_links": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all leaf to L3 server links within a routing zone (VRF). This will be an " +
				"empty list on a layer2 based blueprint",
			Optional: true,
			Computed: true,
		},
		"l2_edge_subnet_links": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all virtual networks (VLANs) that have L3 addresses within a routing zone (VRF).",
			Optional:            true,
			Computed:            true,
		},
		"loopbacks": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all loopbacks within a routing zone (VRF) across spine, leaf, and L3 servers.",
			Optional:            true,
			Computed:            true,
		},
		"static_routes": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all subnets in a VRF associated with static routes from all fabric systems " +
				"to external routers associated with this routing policy",
			Optional: true,
			Computed: true,
		},
	}
}

func (o datacenterRoutingPolicyExport) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"spine_superspine_links": types.BoolType,
		"spine_leaf_links":       types.BoolType,
		"l3_edge_server_links":   types.BoolType,
		"l2_edge_subnet_links":   types.BoolType,
		"loopbacks":              types.BoolType,
		"static_routes":          types.BoolType,
	}
}

func (o *datacenterRoutingPolicyExport) request() *goapstra.DcRoutingExportPolicy {
	// terrible implementation of default values. todo: replace when framework 1.2 rolls around
	if o.superspine.IsUnknown() || o.superspine.IsNull() {
		o.superspine = types.BoolValue(false)
	}
	if o.spine.IsUnknown() || o.spine.IsNull() {
		o.spine = types.BoolValue(false)
	}
	if o.l3Edge.IsUnknown() || o.l3Edge.IsNull() {
		o.l3Edge = types.BoolValue(false)
	}
	if o.l2Edge.IsUnknown() || o.l2Edge.IsNull() {
		o.l2Edge = types.BoolValue(false)
	}
	if o.loopback.IsUnknown() || o.loopback.IsNull() {
		o.loopback = types.BoolValue(false)
	}
	if o.static.IsUnknown() || o.static.IsNull() {
		o.static = types.BoolValue(false)
	}

	return &goapstra.DcRoutingExportPolicy{
		StaticRoutes:         o.superspine.ValueBool(),
		Loopbacks:            o.spine.ValueBool(),
		SpineSuperspineLinks: o.l3Edge.ValueBool(),
		L3EdgeServerLinks:    o.l2Edge.ValueBool(),
		SpineLeafLinks:       o.loopback.ValueBool(),
		L2EdgeSubnets:        o.static.ValueBool(),
	}
}

func (o *datacenterRoutingPolicyExport) loadApiData(_ context.Context, in *goapstra.DcRoutingExportPolicy, _ *diag.Diagnostics) {
	o.superspine = types.BoolValue(in.SpineSuperspineLinks)
	o.spine = types.BoolValue(in.SpineLeafLinks)
	o.l3Edge = types.BoolValue(in.L3EdgeServerLinks)
	o.l2Edge = types.BoolValue(in.L2EdgeSubnets)
	o.loopback = types.BoolValue(in.Loopbacks)
	o.static = types.BoolValue(in.StaticRoutes)
}
