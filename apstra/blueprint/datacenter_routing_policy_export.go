package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datacenterRoutingPolicyExport struct {
	Loopback   types.Bool `tfsdk:"export_loopbacks"`
	Superspine types.Bool `tfsdk:"export_spine_superspine_links"`
	Spine      types.Bool `tfsdk:"export_spine_leaf_links"`
	L3Edge     types.Bool `tfsdk:"export_l3_edge_server_links"`
	L2Edge     types.Bool `tfsdk:"export_l2_edge_subnets"`
	Static     types.Bool `tfsdk:"export_static_routes"`
}

func (o datacenterRoutingPolicyExport) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"export_loopbacks": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all loopbacks within a routing zone (VRF) across spine, leaf, and L3 servers.",
			Computed:            true,
			Optional:            true,
			Default:             booldefault.StaticBool(false),
		},
		"export_spine_superspine_links": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all spine-leaf (fabric) links within a VRF. EVPN routing zones do not have " +
				"spine-leaf addressing, so this generated list may be empty. For routing zones of type Virtual L3 " +
				"Fabric, subinterfaces between spine-leaf will be included.",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(false),
		},
		"export_spine_leaf_links": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all spine-supersine (fabric) links within the default routing zone (VRF)",
			Computed:            true,
			Optional:            true,
			Default:             booldefault.StaticBool(false),
		},
		"export_l3_edge_server_links": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all leaf to L3 server links within a routing zone (VRF). This will be an " +
				"empty list on a layer2 based blueprint",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(false),
		},
		"export_l2_edge_subnets": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all virtual networks (VLANs) that have L3 addresses within a routing zone (VRF).",
			Computed:            true,
			Optional:            true,
			Default:             booldefault.StaticBool(false),
		},
		"export_static_routes": resourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all subnets in a VRF associated with static routes from all fabric systems " +
				"to external routers associated with this routing policy",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(false),
		},
	}
}

func (o datacenterRoutingPolicyExport) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"export_loopbacks":              types.BoolType,
		"export_spine_superspine_links": types.BoolType,
		"export_spine_leaf_links":       types.BoolType,
		"export_l3_edge_server_links":   types.BoolType,
		"export_l2_edge_subnets":        types.BoolType,
		"export_static_routes":          types.BoolType,
	}
}

func (o datacenterRoutingPolicyExport) defaultObject() map[string]attr.Value {
	return map[string]attr.Value{
		"export_loopbacks":              types.BoolValue(false),
		"export_spine_superspine_links": types.BoolValue(false),
		"export_spine_leaf_links":       types.BoolValue(false),
		"export_l3_edge_server_links":   types.BoolValue(false),
		"export_l2_edge_subnets":        types.BoolValue(false),
		"export_static_routes":          types.BoolValue(false),
	}
}

func (o *datacenterRoutingPolicyExport) request() *apstra.DcRoutingExportPolicy {
	return &apstra.DcRoutingExportPolicy{
		Loopbacks:            o.Loopback.ValueBool(),
		SpineSuperspineLinks: o.Superspine.ValueBool(),
		SpineLeafLinks:       o.Spine.ValueBool(),
		L3EdgeServerLinks:    o.L3Edge.ValueBool(),
		L2EdgeSubnets:        o.L2Edge.ValueBool(),
		StaticRoutes:         o.Static.ValueBool(),
	}
}

func (o *datacenterRoutingPolicyExport) loadApiData(_ context.Context, in *apstra.DcRoutingExportPolicy, _ *diag.Diagnostics) {
	o.Loopback = types.BoolValue(in.Loopbacks)
	o.Superspine = types.BoolValue(in.SpineSuperspineLinks)
	o.Spine = types.BoolValue(in.SpineLeafLinks)
	o.L3Edge = types.BoolValue(in.L3EdgeServerLinks)
	o.L2Edge = types.BoolValue(in.L2EdgeSubnets)
	o.Static = types.BoolValue(in.StaticRoutes)
}
