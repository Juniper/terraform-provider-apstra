package private

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type accessRedundancyGroupMembers map[string][2]string

func (gm accessRedundancyGroupMembers) AsMapKeyedBySwitch() map[string]string {
	result := make(map[string]string, 2*len(gm))
	for k, v := range gm {
		result[v[0]] = k
		result[v[1]] = k
	}
	return result
}

type ResourceDatacenterVirtualNetworkAssignment struct {
	ConfiguredLeafID             string                       `json:"configured_leaf_id"` // may be a leaf redundancy group ID or a leaf switch ID
	ConfiguredLeafNodeType       apstra.NodeType              `json:"configured_leaf_node_type"`
	ConfiguredAccessIDs          []string                     `json:"configured_access_ids"`
	AccessRedundancyGroupMembers accessRedundancyGroupMembers `json:"access_redundancy_group_members"`
	PriorVLAN                    *int64                       `json:"prior_vlan"`
}

func (vna *ResourceDatacenterVirtualNetworkAssignment) FetchRedundancyGroups(ctx context.Context, vnID string, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if vna.ConfiguredLeafID == "" {
		diags.AddError(constants.ErrProviderBug, "Leaf ID in private state is empty")
	}
	if diags.HasError() {
		return
	}

	bp.UpdateVirtualNetworkLeafBindings(ctx, apstra.VirtualNetworkBindingsRequest{
		VnId:               "",
		VnBindings:         nil,
		SviIps:             nil,
		DhcpServiceEnabled: nil,
	})

	query := new(apstra.MatchQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client())

	// Begin by matching the Leaf Switch, Leaf Redundancy Group and Virtual Network
	switch {
	case vna.ConfiguredLeafNodeType == apstra.NodeTypeSystem:
		// Mandatory: Leaf -> ... -> Virtual Network
		query.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "system_type", Value: apstra.QEStringVal(enum.SystemTypeSwitch.String())},
				{Key: "role", Value: apstra.QEStringVal(enum.SystemNodeRoleLeaf.String())},
				{Key: "id", Value: apstra.QEStringVal(vna.ConfiguredLeafID)},
				{Key: "name", Value: apstra.QEStringVal("leaf_switch")},
			}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedVnInstances.QEEAttribute()}).
			Node([]apstra.QEEAttribute{apstra.NodeTypeVirtualNetworkInstance.QEEAttribute()}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeInstantiatedBy.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetwork.QEEAttribute(),
				{Key: "id", Value: apstra.QEStringVal(vnID)},
			}),
		)
		// Optional: Leaf -> ... -> Redundancy Group
		query.Optional(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("leaf_switch")}}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRedundancyGroup.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeRedundancyGroup.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("leaf_group")},
			}),
		)
	case vna.ConfiguredLeafNodeType == apstra.NodeTypeRedundancyGroup:
		// Redundancy Group -> ... -> Leaf -> ... -> Virtual Network
		query.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeRedundancyGroup.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("leaf_group")},
				{Key: "id", Value: apstra.QEStringVal(vna.ConfiguredLeafID)},
			}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOfSystems.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("leaf_switch")},
			}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedVnInstances.QEEAttribute()}).
			Node([]apstra.QEEAttribute{apstra.NodeTypeVirtualNetworkInstance.QEEAttribute()}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeInstantiatedBy.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeVirtualNetwork.QEEAttribute(),
				{Key: "id", Value: apstra.QEStringVal(vnID)},
			}),
		)
	default:
		diags.AddError(constants.ErrProviderBug, fmt.Sprintf("Invalid leaf node type in private state: %q", vna.ConfiguredLeafNodeType))
		return
	}

	// Continue matching any Access Switches and any Access Switch Redudancy Groups.
	query.Optional( // All of this is optional.
		new(apstra.MatchQuery).
			Match(new(apstra.PathQuery). // Mandatory component: Leaf -> ... -> Access
							Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("leaf_switch")}}).
							Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
							Node([]apstra.QEEAttribute{
					apstra.NodeTypeInterface.QEEAttribute(),
					{Key: "if_type", Value: apstra.QEStringVal(enum.InterfaceTypeEthernet.String())},
				}).
				Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
				Node([]apstra.QEEAttribute{apstra.NodeTypeLink.QEEAttribute()}).
				In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
				Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
				In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
				Node([]apstra.QEEAttribute{
					apstra.NodeTypeSystem.QEEAttribute(),
					{Key: "name", Value: apstra.QEStringVal("access_switch")},
					{Key: "system_type", Value: apstra.QEStringVal(enum.SystemTypeSwitch.String())},
					{Key: "role", Value: apstra.QEStringVal(enum.SystemNodeRoleAccess.String())},
				}),
			).
			Optional(
				new(apstra.PathQuery). // Optional component: Access -> ... -> Redundancy Group
							Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("access_switch")}}).
							Out([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRedundancyGroup.QEEAttribute()}).
							Node([]apstra.QEEAttribute{
						apstra.NodeTypeRedundancyGroup.QEEAttribute(),
						{Key: "name", Value: apstra.QEStringVal("access_group")},
					}).
					Out([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOfSystems.QEEAttribute()}).
					Node([]apstra.QEEAttribute{
						apstra.NodeTypeSystem.QEEAttribute(),
						{Key: "name", Value: apstra.QEStringVal("access_switch_peer")},
					}).
					EnsureDifferent("access_switch", "access_switch_peer"),
			),
	)

	// Run the query and store the results in target.
	var target struct {
		Items []struct {
			AccessGroup *struct {
				ID string `json:"id"`
			} `json:"access_group"`
			AccessSwitch *struct {
				ID string `json:"ID"`
			} `json:"access_switch"`
			AccessPeer *struct {
				ID string `json:"ID"`
			} `json:"access_switch_peer"`
			LeafGroup *struct {
				ID string `json:"id"`
			} `json:"leaf_group"`
			LeafSwitch *struct {
				ID string `json:"ID"`
			} `json:"leaf_switch"`
		} `json:"items"`
	}
	err := query.Do(ctx, &target)
	if err != nil {
		diags.AddError("failed to quering for leaf relationships", err.Error())
		return
	}
	log.Println(query.String())

	vna.AccessRedundancyGroupMembers = make(accessRedundancyGroupMembers)
	for _, item := range target.Items {
		if item.AccessSwitch != nil && item.AccessGroup != nil {
			vna.AccessRedundancyGroupMembers[item.AccessGroup.ID] = [2]string{
				item.AccessSwitch.ID,
				item.AccessPeer.ID,
			}
		}
	}

}

//func (o *ResourceDatacenterVirtualNetworkAssignment) LoadRedundancyGroupIdToSystemIDsApiData(_ context.Context, rgiMap map[string]*apstra.RedundancyGroupInfo, ids []string, _ *diag.Diagnostics) {
//	o.RedundancyGroupIdToSystemIDs = make(map[string][]string)
//	for _, id := range ids {
//		if entry, ok := rgiMap[id]; ok {
//			o.RedundancyGroupIdToSystemIDs[entry.Id.String()] = []string{entry.SystemIds[0].String(), entry.SystemIds[1].String()}
//		}
//	}
//}

func (vna *ResourceDatacenterVirtualNetworkAssignment) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, d := ps.GetKey(ctx, fmt.Sprintf("%T", *vna))
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	if len(b) == 0 {
		return
	}

	err := json.Unmarshal(b, &vna)
	if err != nil {
		diags.AddError("failed to unmarshal private state", err.Error())
		return
	}
}

func (vna *ResourceDatacenterVirtualNetworkAssignment) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	b, err := json.Marshal(vna)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	diags.Append(ps.SetKey(ctx, fmt.Sprintf("%T", *vna), b)...)
}
