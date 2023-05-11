package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/utils"
)

type DatacenterVirtualNetwork struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	BlueprintId   types.String `tfsdk:"blueprint_id"`
	RoutingZoneId types.String `tfsdk:"routing_zone_id"`
	Type          types.String `tfsdk:"type"`
	LeafSwitchIds types.List   `tfsdk:"leaf_switch_ids"`
}

func (o DatacenterVirtualNetwork) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Name",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 30),
				stringvalidator.RegexMatches(regexp.MustCompile(design.AlphaNumericRegexp), "valid characters are: "+design.AlphaNumericChars),
			},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Type",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(apstra.VnTypeVxlan.String()),
			Validators:          []validator.String{apstravalidator.OneOfStringers(apstra.AllVirtualNetworkTypes())},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Routing Zone ID (required when `type == %s`", apstra.VnTypeVxlan),
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.StringRequiredWhenValueIs(path.MatchRelative().AtParent().AtName("type"), fmt.Sprintf("%q", apstra.VnTypeVxlan)),
			},
		},
		"leaf_switch_ids": resourceSchema.ListAttribute{
			MarkdownDescription: "Graph DB node IDs of Leaf Switches to which this Virtual Network should be bound",
			Required:            true, // todo: can become optional when access_switch_ids added
			ElementType:         types.StringType,
			Validators: []validator.List{
				//listvalidator.AtLeastOneOf(),
				listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *DatacenterVirtualNetwork) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.VirtualNetworkData {
	var err error

	var vnType apstra.VnType
	err = vnType.FromString(o.Type.ValueString())
	if err != nil {
		diags.Append(
			validatordiag.BugInProviderDiagnostic(
				fmt.Sprintf("error parsing virtual network type %q - %s", o.Type.String(), err.Error())))
		return nil
	}

	var leafSwitchNodeIds []apstra.ObjectId
	o.LeafSwitchIds.ElementsAs(ctx, &leafSwitchNodeIds, false)

	vnBindings := make([]apstra.VnBinding, len(leafSwitchNodeIds))
	for i := range leafSwitchNodeIds {
		vnBindings[i].SystemId = leafSwitchNodeIds[i]
	}

	return &apstra.VirtualNetworkData{
		DhcpService:               false,
		Ipv4Enabled:               false,
		Ipv4Subnet:                nil,
		Ipv6Enabled:               false,
		Ipv6Subnet:                nil,
		Label:                     o.Name.ValueString(),
		ReservedVlanId:            nil,
		RouteTarget:               "",
		RtPolicy:                  nil,
		SecurityZoneId:            apstra.ObjectId(o.RoutingZoneId.ValueString()),
		SviIps:                    nil,
		VirtualGatewayIpv4:        nil,
		VirtualGatewayIpv6:        nil,
		VirtualGatewayIpv4Enabled: false,
		VirtualGatewayIpv6Enabled: false,
		VnBindings:                vnBindings,
		VnId:                      nil,
		VnType:                    vnType,
		VirtualMac:                nil,
	}
}

func (o *DatacenterVirtualNetwork) LoadApiData(_ context.Context, in *apstra.VirtualNetworkData, _ *diag.Diagnostics) {
	leafSwitchIds := make([]attr.Value, len(in.VnBindings))
	for i, vnBinding := range in.VnBindings {
		leafSwitchIds[i] = types.StringValue(vnBinding.SystemId.String())
	}

	o.Name = types.StringValue(in.Label)
	o.Type = types.StringValue(in.VnType.String())
	o.RoutingZoneId = types.StringValue(in.SecurityZoneId.String())
	o.LeafSwitchIds = types.ListValueMust(types.StringType, leafSwitchIds)
}

//func (o DatacenterVirtualNetwork) DataSourceAttributes() map[string]datasourceSchema.Attribute {
//	return map[string]datasourceSchema.Attribute{
//		"id": datasourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra graph node ID.",
//			Computed:            true,
//		},
//		"name": datasourceSchema.StringAttribute{
//			MarkdownDescription: "Virtual Network Name",
//			Required:            true,
//			Validators:          []validator.String{stringvalidator.LengthBetween(1, 30)},
//		},
//		"type": datasourceSchema.StringAttribute{
//			MarkdownDescription: "Virtual Network Type",
//			Optional:            true,
//			Computed:            true,
//			Validators:          []validator.String{apstravalidator.OneOfStringers(apstra.AllNodeDeployModes())},
//		},
//		"blueprint_id": datasourceSchema.StringAttribute{},
//	}
//}

func includeRedundancyPeers(ctx context.Context, in []string, client *apstra.TwoStageL3ClosClient) ([]string, error) {
	// figure out which switches are part of redundancy groups and which are not
	groupIds, nonRedundantIds, err := identifyRedundantSystems(ctx, in, client)
	if err != nil {
		return nil, err
	}

	// Now use the group IDs to work backward to redundant system IDs.
	// Note that redundant IDs may include systems not cited by the caller
	// because expanding the group ID -> member IDs will catch *all* members,
	// even those omitted by the caller. This is deliberate.
	redundantIds, err := redundancyGroupToMembers(ctx, groupIds, client)
	if err != nil {
		return nil, err
	}

	return append(redundantIds, nonRedundantIds...), nil
}

func redundancyPeersFromIds(in []string, redundancyInfo map[string]*redundancyGroup) []string {
	peers := make(map[string]struct{})
	nonRedundant := make(map[string]struct{})

	for _, switchId := range in {
		if rgInfo, ok := redundancyInfo[switchId]; ok {
			for _, memberId := range rgInfo.memberIds {
				peers[memberId] = struct{}{}
			}
		} else {
			nonRedundant[switchId] = struct{}{}
		}
	}

	result := make([]string, len(peers)+len(nonRedundant))
	i := 0
	for peerId := range peers {
		result[i] = peerId
		i++
	}
	for nonRedundantId := range nonRedundant { // continue looping with accumulated 'i' value
		result[i] = nonRedundantId
		i++
	}

	return result
}

func accessSwitchIdsToParentLeafIds(ctx context.Context, ids []string, client *apstra.TwoStageL3ClosClient) (map[string][]string, error) {
	pathQuery := new(apstra.PathQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "role", Value: apstra.QEStringVal("access")},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
			{Key: "id", Value: apstra.QEStringValIsIn(ids)},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("hosted_interfaces")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("interface")},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("link")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("link")},
			{Key: "link_type", Value: apstra.QEStringVal("ethernet")},
		}).
		In([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("link")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("interface")},
		}).
		In([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("hosted_interfaces")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "role", Value: apstra.QEStringVal("leaf")},
			{Key: "name", Value: apstra.QEStringVal("n_leaf")},
		})

	queryResponse := &struct {
		Items []struct {
			Access struct {
				Id string `json:"id"`
			} `json:"n_access"`
			Leaf struct {
				Id string `json:"id"`
			} `json:"n_leaf"`
		} `json:"items"`
	}{}

	err := new(apstra.MatchQuery).
		Match(pathQuery).
		Distinct(apstra.MatchQueryDistinct{"n_leaf"}).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Do(ctx, queryResponse)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, item := range queryResponse.Items {
		if element, ok := result[item.Access.Id]; ok {
			element = append(element, item.Leaf.Id)
		} else {
			result[item.Access.Id] = []string{item.Leaf.Id}
		}
	}

	// filter duplicates from result
	for k, v := range result {
		result[k] = utils.Uniq(v)
	}

	return result, nil
}

func switchIdsToBindings(ctx context.Context, in []string, vlan *apstra.Vlan, client *apstra.TwoStageL3ClosClient) ([]apstra.VnBinding, error) {
	// create two maps of redundancy group info: One keyed by redundancy group
	// id and one keyed by redundancy group member id.
	rgIdToRgInfo, err := redunancyGroupIdToRedundancyGroupInfo(ctx, client)
	if err != nil {
		return nil, err
	}
	memberIdToRgInfo := make(map[string]*redundancyGroup)
	for k, v := range rgIdToRgInfo {
		for i := range v.memberIds {
			memberIdToRgInfo[v.memberIds[i]] = rgIdToRgInfo[k]
		}
	}

	// filter `in` into two slices: accessSwitchIds and leafSwitchIds. This is
	// where we'll produce an error if the ID is bogus or belongs to a different
	// type of graph db node.
	sysIdToRole, err := getSystemRoles(ctx, in, client)
	if err != nil {
		return nil, err
	}
	var leafSwitchIds, accessSwitchIds []string
	for k, v := range sysIdToRole {
		switch v {
		case apstra.SystemRoleAccess:
			accessSwitchIds = append(accessSwitchIds, k)
		case apstra.SystemRoleLeaf:
			leafSwitchIds = append(leafSwitchIds, k)
		default:
			return nil, fmt.Errorf("unhandled system role %q for node id %q, only 'access' and 'leaf' are expected", v, k)
		}
	}

	// Expand the list of access switches to include redundancy group peers in
	// case any are missing from the caller-supplied list.
	accessSwitchIds = redundancyPeersFromIds(accessSwitchIds, memberIdToRgInfo)

	// Identify leaf switches upstream of access switches
	parentLeafMap, err := accessSwitchIdsToParentLeafIds(ctx, accessSwitchIds, client)
	if err != nil {
		return nil, err
	}

	// leafToVnBinding is a map keyed by either leaf node ID (non-redundant
	// leafs) or leaf redundancy group ID
	leafToVnBinding := make(map[apstra.ObjectId]apstra.VnBinding)

	// loop over access switches, create or update leafToVnBinding entry for each
	for _, accessSwitchId := range accessSwitchIds {
		var redundantAccess bool
		var redundantAccessID string
		if rg, ok := memberIdToRgInfo[accessSwitchId]; ok {
			if rg.role != "access" {
				return nil, fmt.Errorf("access switch %q is a member of %q redundancy group %q", accessSwitchId, rg.role, rg.id)
			}
			redundantAccess = true
			redundantAccessID = rg.id
		}

		// find all parent switches of this access switch
		var parentLeafIDs []string
		if s, ok := parentLeafMap[accessSwitchId]; ok {
			parentLeafIDs = s
		}
		if len(parentLeafIDs) == 0 {
			return nil, fmt.Errorf("access switch %q doesn't have any parent leafs", accessSwitchId)
		}

		// swap each parent leaf ID for its redundancy group ID (if any), and
		// then reduce the list to a single ID representing all parents. That
		// single ID will be either a leaf ID (standalone leaf) or a redundancy
		// group ID (ESI leaf)
		for i, plID := range parentLeafIDs {
			if rg, ok := memberIdToRgInfo[plID]; ok {
				parentLeafIDs[i] = rg.id
			}
		}
		parentLeafIDs = utils.Uniq(parentLeafIDs)
		if len(parentLeafIDs) != 1 {
			return nil, fmt.Errorf("failed to reduce access switch %q parents to a single ID, got '%v'", accessSwitchId, parentLeafIDs)
		}

		// our desired IDs for the bindings are now scattered around. Pick 'em
		// out and turn them into apstra.ObjectId for use in the VnBinding{}
		var leafBindingId, accessBindingId apstra.ObjectId
		leafBindingId = apstra.ObjectId(parentLeafIDs[0])
		if redundantAccess {
			accessBindingId = apstra.ObjectId(redundantAccessID)
		} else {
			accessBindingId = apstra.ObjectId(accessSwitchId)
		}

		// We may have already created a binding for this leaf...
		if vnb, ok := leafToVnBinding[leafBindingId]; ok {
			// binding exists, add our access ID to the list
			vnb.AccessSwitchNodeIds = utils.Uniq(append(vnb.AccessSwitchNodeIds, accessBindingId))
			leafToVnBinding[leafBindingId] = vnb
		} else {
			// binding not found, create a new one
			leafToVnBinding[leafBindingId] = apstra.VnBinding{
				AccessSwitchNodeIds: []apstra.ObjectId{accessBindingId},
				SystemId:            leafBindingId,
				VlanId:              vlan,
			}
		}
	}

	// loop over leaf switches, create a leafToVnBinding entry as required
	for _, leafSwitchId := range leafSwitchIds {
		var leafBindingId apstra.ObjectId
		if rg, ok := memberIdToRgInfo[leafSwitchId]; ok {
			if rg.role != "leaf" {
				return nil, fmt.Errorf("leaf switch %q is a member of %q redundancy group %q", leafSwitchId, rg.role, rg.id)
			}
			leafBindingId = apstra.ObjectId(rg.id)
		} else {
			leafBindingId = apstra.ObjectId(leafSwitchId)
		}

		// We may have already created a binding for this leaf...
		if _, ok := leafToVnBinding[leafBindingId]; !ok {
			// binding not found. create one.
			leafToVnBinding[leafBindingId] = apstra.VnBinding{
				SystemId: leafBindingId,
				VlanId:   vlan,
			}
		}
	}

	result := make([]apstra.VnBinding, len(leafToVnBinding))
	i := 0
	for _, v := range leafToVnBinding {
		result[i] = v
		i++
	}

	return result, nil
}

func redundancyGroupToMembers(ctx context.Context, in []string, client *apstra.TwoStageL3ClosClient) ([]string, error) {
	query := new(apstra.PathQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("redundancy_group")},
			{Key: "id", Value: apstra.QEStringValIsIn(in)},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("composed_of_systems")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})

	queryResult := &struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}{}

	err := query.Do(ctx, queryResult)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(queryResult.Items))
	for i := range queryResult.Items {
		result[i] = queryResult.Items[i].System.Id
	}

	return result, nil
}

// identifyRedundantSystems accepts a []string representing graph DB nodes of
// type 'system' and returns a []string representing redundancy_group IDs
// including those IDs and another []string including IDs which were not found
// to be part of a redundancy_group.
// func identifyRedundantSystems(ctx context.Context, req identifyRedundantSystemsRequest) (*identifyRedundantSystemsResponse, error) {
func identifyRedundantSystems(ctx context.Context, in []string, client *apstra.TwoStageL3ClosClient) ([]string, []string, error) {
	query := new(apstra.PathQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "id", Value: apstra.QEStringValIsIn(in)},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("part_of_redundancy_group")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("redundancy_group")},
			{Key: "name", Value: apstra.QEStringVal("n_redundancy_group")},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("composed_of_systems")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})

	queryResult := &struct {
		Items []struct {
			RedundancyGroup struct {
				Id string `json:"id"`
			} `json:"n_redundancy_group"`
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}{}

	err := query.Do(ctx, queryResult)
	if err != nil {
		return nil, nil, err
	}

	// store the discovered redundancy group and group member IDs as a map for easy lookup
	redundancyGroupIds := make(map[string]struct{})
	redundancyGroupMemberIds := make(map[string]struct{})
	for i := range queryResult.Items {
		redundancyGroupIds[queryResult.Items[i].RedundancyGroup.Id] = struct{}{}
		redundancyGroupMemberIds[queryResult.Items[i].System.Id] = struct{}{}
	}

	// loop through the group ID map, populate the returned group ID slice
	groupIdResult := make([]string, len(redundancyGroupIds))
	i := 0
	for id := range redundancyGroupIds {
		groupIdResult[i] = id
		i++
	}

	// loop through the supplied switch IDs, add IDs which are not group
	// members to the returned non-redundant slice
	var nonRedundantIdResult []string
	for i := range in {
		if _, ok := redundancyGroupMemberIds[in[i]]; !ok {
			nonRedundantIdResult = append(nonRedundantIdResult, in[i])
		}
	}

	return groupIdResult, nonRedundantIdResult, nil
}

// redundancyGroupAndMembersFromSwitchIds accepts a slice of switch IDs and
// organizes them into a map keyed by redundancy group ID, with values being
// all switch IDs participating in the redundancy group (even those not part
// of the supplied slice).
//
// The returned []string includes IDs which are not found to represent switches
// participating in a redundancy group.
func redundancyGroupAndMembersFromSwitchIds(ctx context.Context, in []string, client *apstra.TwoStageL3ClosClient) (map[string][]string, []string, error) {
	result := make(map[string][]string)
	var nonRedundantIds []string
	for i := range in {
		query, queryResponse := redundancyGroupQueryAndResponse(in[i])
		err := query.SetClient(client.Client()).
			SetBlueprintType(apstra.BlueprintTypeStaging).
			SetBlueprintId(client.Id()).
			Do(ctx, queryResponse)
		if err != nil {
			return nil, nil, err
		}

		var groupMembers []string
		for _, item := range queryResponse.Items {
			groupMembers = append(groupMembers, item.System.Id)
		}

		if len(queryResponse.Items) == 0 {
			nonRedundantIds = append(nonRedundantIds, in[i])
		} else {
			result[queryResponse.Items[0].RedundancyGroup.Id] = groupMembers
		}
	}
	return result, nonRedundantIds, nil
}

func getSystemRoles(ctx context.Context, systemIds []string, client *apstra.TwoStageL3ClosClient) (map[string]apstra.SystemRole, error) {
	nodesResponse := new(struct {
		Nodes map[string]struct {
			Role string `json:"role"`
		} `json:"nodes"`
	})
	err := client.GetNodes(ctx, apstra.NodeTypeSystem, nodesResponse)
	if err != nil {
		return nil, fmt.Errorf("error querying for system nodes - %w", err)
	}

	result := make(map[string]apstra.SystemRole, len(systemIds))
	for i, systemId := range systemIds {
		if node, ok := nodesResponse.Nodes[systemId]; ok {
			role := new(apstra.SystemRole)
			err = role.FromString(node.Role)
			if err != nil {
				return nil, fmt.Errorf("error parsing %q as system role - %w", systemId, err)
			}
			result[systemIds[i]] = *role
		} else {
			return nil, fmt.Errorf("system node with ID %q not found in blueprint %q", systemIds, client.Id())
		}
	}
	return result, nil
}

func redundancyGroupQueryAndResponse(switchId string) (*apstra.PathQuery, *struct {
	Items []struct {
		RedundancyGroup struct {
			Id string `json:"id"`
		} `json:"n_redundancy_group"`
		System struct {
			Id string `json:"id"`
		} `json:"n_system"`
	} `json:"items"`
}) {
	query := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "id", Value: apstra.QEStringVal(switchId)},
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
			{Key: "role", Value: apstra.QEStringValIsIn{"access", "leaf"}},
		}).
		Out([]apstra.QEEAttribute{{Key: "type", Value: apstra.QEStringVal("part_of_redundancy_group")}}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("redundancy_group")},
			{Key: "name", Value: apstra.QEStringVal("n_redundancy_group")},
		}).
		Out([]apstra.QEEAttribute{{Key: "type", Value: apstra.QEStringVal("composed_of_systems")}}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
			{Key: "role", Value: apstra.QEStringValIsIn{"access", "leaf"}},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})

	queryResponse := &struct {
		Items []struct {
			RedundancyGroup struct {
				Id string `json:"id"`
			} `json:"n_redundancy_group"`
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}{}

	return query, queryResponse
}

//func accessSwitchesUpstreamLeafs(ctx context.Context, ids []string, client *apstra.TwoStageL3ClosClient) ([]string, error) {
//	pathQuery := new(apstra.PathQuery).
//		SetClient(client.Client()).
//		SetBlueprintId(client.Id()).
//		SetBlueprintType(apstra.BlueprintTypeStaging).
//		Node([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("system")},
//			{Key: "role", Value: apstra.QEStringVal("access")},
//			{Key: "id", Value: apstra.QEStringValIsIn(ids)},
//		}).
//		Out([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("hosted_interfaces")},
//		}).
//		Node([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("interface")},
//		}).
//		Out([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("link")},
//		}).
//		Node([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("link")},
//			{Key: "link_type", Value: apstra.QEStringVal("ethernet")},
//		}).
//		In([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("hosted_interfaces")},
//		}).
//		Node([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("interface")},
//		}).
//		In([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("hosted_interfaces")},
//		}).
//		Node([]apstra.QEEAttribute{
//			{Key: "type", Value: apstra.QEStringVal("system")},
//			{Key: "role", Value: apstra.QEStringVal("leaf")},
//			{Key: "name", Value: apstra.QEStringVal("n_leaf")},
//		})
//
//	queryResponse := &struct {
//		Items []struct {
//			Leaf struct {
//				Id string `json:"id"`
//			} `json:"n_leaf"`
//		} `json:"items"`
//	}{}
//
//	err := new(apstra.MatchQuery).
//		Match(pathQuery).
//		Distinct(apstra.MatchQueryDistinct{"n_leaf"}).
//		SetClient(client.Client()).
//		SetBlueprintId(client.Id()).
//		SetBlueprintType(apstra.BlueprintTypeStaging).
//		Do(ctx, queryResponse)
//	if err != nil {
//		return nil, err
//	}
//
//	result := make([]string, len(queryResponse.Items))
//	for i := range queryResponse.Items {
//		result[i] = queryResponse.Items[i].Leaf.Id
//	}
//
//	return result, nil
//}

type redundancyGroup struct {
	role      string   // access / generic
	id        string   // redundancy_group graph db node id
	memberIds []string // id of leaf/access nodes in the group
}

func redunancyGroupIdToRedundancyGroupInfo(ctx context.Context, client *apstra.TwoStageL3ClosClient) (map[string]*redundancyGroup, error) {
	pathQuery := new(apstra.PathQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("redundancy_group")},
			{Key: "name", Value: apstra.QEStringVal("n_redundancy_group")},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("composed_of_systems")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
			{Key: "role", Value: apstra.QEStringValIsIn{"access", "leaf"}},
		})

	queryResponse := &struct {
		Items []struct {
			RedundancyGroup struct {
				Id string `json:"id"`
			} `json:"n_redundancy_group"`
			System struct {
				Id   string `json:"id"`
				Role string `json:"role"`
			} `json:"n_system"`
		} `json:"items"`
	}{}

	err := new(apstra.MatchQuery).
		Match(pathQuery).
		Distinct(apstra.MatchQueryDistinct{"n_leaf"}).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Do(ctx, queryResponse)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*redundancyGroup)
	for _, item := range queryResponse.Items {
		id := item.RedundancyGroup.Id
		if rg, ok := result[id]; ok {
			rg.memberIds = append(rg.memberIds, item.System.Id)
			result[id] = rg
		} else {
			result[id] = &redundancyGroup{
				role:      item.System.Role,
				id:        item.RedundancyGroup.Id,
				memberIds: []string{item.System.Id},
			}
		}
	}

	return result, nil
}
