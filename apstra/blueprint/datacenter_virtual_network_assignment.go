package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkAssignment struct {
	BlueprintID types.String       `tfsdk:"blueprint_id"`
	VNID        types.String       `tfsdk:"virtual_network_id"`
	LeafID      types.String       `tfsdk:"leaf_switch_id"`
	LeafRGID    types.String       `tfsdk:"leaf_redundancy_group_id"`
	VLAN        types.Int64        `tfsdk:"vlan"`
	IPv4Mode    types.String       `tfsdk:"ipv4_mode"`
	IPv4Address cidrtypes.IPPrefix `tfsdk:"ipv4_address"`
	IPv6Mode    types.String       `tfsdk:"ipv6_mode"`
	IPv6Address cidrtypes.IPPrefix `tfsdk:"ipv6_address"`
	AccessIDs   types.Set          `tfsdk:"access_ids"`
}

func (vna *VirtualNetworkAssignment) DatasourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"blueprint_id":             datasourceSchema.StringAttribute{},
		"virtual_network_id":       datasourceSchema.StringAttribute{},
		"leaf_switch_id":           datasourceSchema.StringAttribute{},
		"leaf_redundancy_group_id": datasourceSchema.StringAttribute{},
		"vlan":                     datasourceSchema.Int64Attribute{},
		"ipv4_mode":                datasourceSchema.StringAttribute{},
		"ipv4_address":             datasourceSchema.StringAttribute{},
		"ipv6_mode":                datasourceSchema.StringAttribute{},
		"ipv6_address":             datasourceSchema.StringAttribute{},
		"access_ids":               datasourceSchema.SetAttribute{},
	}
}

func (vna *VirtualNetworkAssignment) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id":             resourceSchema.StringAttribute{},
		"virtual_network_id":       resourceSchema.StringAttribute{},
		"leaf_switch_id":           resourceSchema.StringAttribute{},
		"leaf_redundancy_group_id": resourceSchema.StringAttribute{},
		"vlan":                     resourceSchema.Int64Attribute{},
		"ipv4_mode":                resourceSchema.StringAttribute{},
		"ipv4_address":             resourceSchema.StringAttribute{},
		"ipv6_mode":                resourceSchema.StringAttribute{},
		"ipv6_address":             resourceSchema.StringAttribute{},
		"access_ids":               resourceSchema.SetAttribute{},
	}
}

// FetchLeafRedundancyGroupID queries the Blueprint for the Leaf Switch and its Redundancy Group (if any)
// and sets the LeafRGID field accordingly. If the Leaf Switch is not part of a Redundancy Group, LeafRGID
// will be set to null. This function should be run only once: Early during Create()
func (vna *VirtualNetworkAssignment) FetchLeafRedundancyGroupID(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	query := new(apstra.MatchQuery).
		SetClient(bp.Client()).
		SetBlueprintId(bp.Id()).
		Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "id", Value: apstra.QEStringVal(vna.LeafID.ValueString())},
				{Key: "name", Value: apstra.QEStringVal("leaf")},
			}),
		).Optional(
		new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("leaf")}}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRedundancyGroup.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeRedundancyGroup.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("redundancy_group")},
			}),
	)

	var target struct {
		Items []struct {
			Leaf struct {
				Type       string `tfsdk:"type"`
				SystemType string `tfsdk:"system_type"`
				Role       string `tfsdk:"role"`
				ID         string `json:"id"`
			} `json:"leaf"`
			RedundancyGroup *struct {
				ID string `json:"id"`
			} `json:"redundancy_group"`
		} `json:"items"`
	}

	err := query.Do(ctx, &target)
	if err != nil {
		diags.AddError("failed while checking for leaf redundancy group", err.Error())
		return
	}

	switch len(target.Items) {
	case 0:
		diags.AddError("failed while checking for leaf redundancy group", "leaf switch not found")
	case 1: // Expected case falls through.
	default:
		diags.AddError("failed while checking for leaf redundancy group", "found multiple matches")
	}
	if diags.HasError() {
		return
	}

	leaf := &target.Items[0].Leaf
	group := target.Items[0].RedundancyGroup

	switch {
	case leaf.Type != apstra.NodeTypeSystem.String():
		diags.AddError("failed while checking for leaf redundancy group", "leaf switch node has type "+leaf.Type+", expected "+apstra.NodeTypeSystem.String())
	case leaf.SystemType != enum.SystemTypeSwitch.String():
		diags.AddError("failed while checking for leaf redundancy group", "leaf switch node has system_type "+leaf.SystemType+", expected "+enum.SystemTypeSwitch.String())
	case leaf.Role != enum.SystemNodeRoleLeaf.String():
		diags.AddError("failed while checking for leaf redundancy group", "leaf switch node has role "+leaf.Role+", expected "+enum.SystemNodeRoleLeaf.String())
	case leaf.ID == "":
		diags.AddError("failed while checking for leaf redundancy group", "leaf switch node has empty ID")
	case group != nil && group.ID == "":
		diags.AddError("failed while checking for leaf redundancy group", "leaf switch redundancy group node has empty ID")
	}
	if diags.HasError() {
		return
	}

	if group == nil {
		vna.LeafRGID = types.StringNull()
	} else {
		vna.LeafRGID = types.StringValue(group.ID)
	}
}

func (vna VirtualNetworkAssignment) Request(ctx context.Context, diags *diag.Diagnostics) apstra.VirtualNetworkBindingsRequest {

}

func (vna *VirtualNetworkAssignment) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {

}
