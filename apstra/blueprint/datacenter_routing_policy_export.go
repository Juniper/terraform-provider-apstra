package blueprint

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
			MarkdownDescription: fmt.Sprintf("Exports all leaf to L3 server links within a routing zone (VRF). "+
				"This will be an empty list on a layer2 based blueprint. Valid only with Apstra %s and earlier.", apiversions.Apstra422),
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

func (o datacenterRoutingPolicyExport) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"export_loopbacks": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all loopbacks within a routing zone (VRF) across spine, leaf, and L3 servers.",
			Computed:            true,
		},
		"export_spine_superspine_links": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all spine-leaf (fabric) links within a VRF. EVPN routing zones do not have " +
				"spine-leaf addressing, so this generated list may be empty. For routing zones of type Virtual L3 " +
				"Fabric, subinterfaces between spine-leaf will be included.",
			Computed: true,
		},
		"export_spine_leaf_links": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all spine-supersine (fabric) links within the default routing zone (VRF)",
			Computed:            true,
		},
		"export_l3_edge_server_links": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all leaf to L3 server links within a routing zone (VRF). This will be an " +
				"empty list on a layer2 based blueprint",
			Computed: true,
		},
		"export_l2_edge_subnets": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all virtual networks (VLANs) that have L3 addresses within a routing zone (VRF).",
			Computed:            true,
		},
		"export_static_routes": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all subnets in a VRF associated with static routes from all fabric systems " +
				"to external routers associated with this routing policy",
			Computed: true,
		},
	}
}

func (o datacenterRoutingPolicyExport) dataSourceAttributesAsFilter() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"export_loopbacks": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all loopbacks within a routing zone (VRF) across spine, leaf, and L3 servers.",
			Optional:            true,
		},
		"export_spine_superspine_links": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all spine-leaf (fabric) links within a VRF. EVPN routing zones do not have " +
				"spine-leaf addressing, so this generated list may be empty. For routing zones of type Virtual L3 " +
				"Fabric, subinterfaces between spine-leaf will be included.",
			Optional: true,
		},
		"export_spine_leaf_links": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all spine-supersine (fabric) links within the default routing zone (VRF)",
			Optional:            true,
		},
		"export_l3_edge_server_links": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all leaf to L3 server links within a routing zone (VRF). This will be an " +
				"empty list on a layer2 based blueprint",
			Optional: true,
		},
		"export_l2_edge_subnets": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all virtual networks (VLANs) that have L3 addresses within a routing zone (VRF).",
			Optional:            true,
		},
		"export_static_routes": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Exports all subnets in a VRF associated with static routes from all fabric systems " +
				"to external routers associated with this routing policy",
			Optional: true,
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

func (o *datacenterRoutingPolicyExport) filterMatch(_ context.Context, in *datacenterRoutingPolicyExport, _ *diag.Diagnostics) bool {
	if !o.Loopback.IsNull() && !o.Loopback.Equal(in.Loopback) {
		return false
	}

	if !o.Superspine.IsNull() && !o.Superspine.Equal(in.Superspine) {
		return false
	}

	if !o.Spine.IsNull() && !o.Spine.Equal(in.Spine) {
		return false
	}

	if !o.L3Edge.IsNull() && !o.L3Edge.Equal(in.L3Edge) {
		return false
	}

	if !o.L2Edge.IsNull() && !o.L2Edge.Equal(in.L2Edge) {
		return false
	}

	if !o.Static.IsNull() && !o.Static.Equal(in.Static) {
		return false
	}

	return true
}
