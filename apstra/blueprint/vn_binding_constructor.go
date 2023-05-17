package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/utils"
)

type VnBindingConstructor struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	VlanId      types.Int64  `tfsdk:"vlan_id"`
	SwitchIds   types.Set    `tfsdk:"switch_ids"`
	Bindings    types.Set    `tfsdk:"bindings"`
}

func (o VnBindingConstructor) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to determine " +
				"the redundancy group and access/leaf relationships of " +
				"each specified switch ID.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "VLAN ID will be populated directly into " +
				"the `bindings` output.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(design.VlanMin-1, design.VlanMax+1)},
		},
		"switch_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of graph db node IDs representing " +
				"access and/or leaf switches for which a binding should " +
				"be constructed.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
		"bindings": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "A set of bindings appropriate for use " +
				"in a `apstra_datacenter_virtual_network` resource",
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: VnBinding{}.DataSourceAttributesConstructorOutput(),
			},
		},
	}
}

func (o *VnBindingConstructor) Compute(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	bpClient, err := client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		diags.AddError("error creating blueprint client", err.Error())
		return
	}

	// only one VLAN per constructor; get it in the expected form
	var vlanId *apstra.Vlan
	if utils.Known(o.VlanId) {
		v := apstra.Vlan(o.VlanId.ValueInt64())
		vlanId = &v
	}

	// extract o.SwitchIds to []string
	var switchIds []string
	diags.Append(o.SwitchIds.ElementsAs(ctx, &switchIds, false)...)
	if diags.HasError() {
		return
	}

	// create two maps of redundancy group info: One keyed by redundancy group
	// id and one keyed by redundancy group member id.
	rgIdToRgInfo, err := redunancyGroupIdToRedundancyGroupInfo(ctx, bpClient)
	if err != nil {
		diags.AddError("error mapping redundancy group info", err.Error())
		return
	}
	memberIdToRgInfo := make(map[string]*redundancyGroup)
	for k, v := range rgIdToRgInfo {
		for i := range v.memberIds {
			memberIdToRgInfo[v.memberIds[i]] = rgIdToRgInfo[k]
		}
	}

	// filter `switchIds` into two slices: accessSwitchIds and leafSwitchIds.
	// This is where we'll produce an error if the ID is bogus or belongs to a
	// some other type of graph db node.
	sysIdToRole, err := getSystemRoles(ctx, switchIds, bpClient)
	if err != nil {
		diags.AddError("error determining system roles", err.Error())
		return
	}
	var leafSwitchIds, accessSwitchIds []string
	for k, v := range sysIdToRole {
		switch v {
		case apstra.SystemRoleAccess:
			accessSwitchIds = append(accessSwitchIds, k)
		case apstra.SystemRoleLeaf:
			leafSwitchIds = append(leafSwitchIds, k)
		default:
			diags.AddError("invalid system role",
				fmt.Sprintf("unhandled system role %q for node id %q, only 'access' and 'leaf' are expected",
					v, k),
			)
		}
	}
	if diags.HasError() {
		return
	}

	// Expand the list of access switches to include redundancy group peers in
	// case any are missing from the caller-supplied list.
	for i := len(accessSwitchIds) - 1; i >= 0; i-- { // loop backward through accessSwitchIds
		if rgInfo, ok := memberIdToRgInfo[accessSwitchIds[i]]; ok {
			// this ID is part of a redundancy group. Copy last slice element to
			// current index and replace last element with full member list.
			accessSwitchIds[i] = accessSwitchIds[len(accessSwitchIds)-1]
			accessSwitchIds = append(accessSwitchIds[:len(accessSwitchIds)-1], rgInfo.memberIds...)
		}
	}

	// Identify leaf switches upstream of access switches
	accessIdToParentLeafIdsMap, err := accessSwitchIdsToParentLeafIds(ctx, accessSwitchIds, bpClient)
	if err != nil {
		diags.AddError("error determining parent leaf of access switches", err.Error())
		return
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
				diags.AddError("wrong redundancy group type",
					fmt.Sprintf("access switch %q is a member of %q redundancy group %q",
						accessSwitchId, rg.role, rg.id),
				)
				return
			}
			redundantAccess = true
			redundantAccessID = rg.id
		}

		// find all parent switches of this access switch
		var parentLeafIDs []string
		if s, ok := accessIdToParentLeafIdsMap[accessSwitchId]; ok {
			parentLeafIDs = s
		}
		if len(parentLeafIDs) == 0 {
			diags.AddError("unable to find access switch parent",
				fmt.Sprintf("access switch %q doesn't have any parent leafs", accessSwitchId),
			)
			return
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
			diags.AddError("access switch has multiple parents not in a single redundancy group",
				fmt.Sprintf("failed to reduce access switch %q parents to a single ID, got '%v'",
					accessSwitchId, parentLeafIDs),
			)
			return
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
				VlanId:              vlanId,
			}
		}
	}

	// loop over leaf switches, create a leafToVnBinding entry as required
	for _, leafSwitchId := range leafSwitchIds {
		var leafBindingId apstra.ObjectId
		if rg, ok := memberIdToRgInfo[leafSwitchId]; ok {
			if rg.role != "leaf" {
				diags.AddError("redundancy group type mismatch",
					fmt.Sprintf("leaf switch %q is a member of %q redundancy group %q",
						leafSwitchId, rg.role, rg.id),
				)
				return
			}
			leafBindingId = apstra.ObjectId(rg.id)
		} else {
			leafBindingId = apstra.ObjectId(leafSwitchId)
		}

		// We may have already created a binding for this leafBindingId...
		if _, ok := leafToVnBinding[leafBindingId]; !ok {
			// binding not found. create one.
			leafToVnBinding[leafBindingId] = apstra.VnBinding{
				SystemId: leafBindingId,
				VlanId:   vlanId,
			}
		}
	}

	bindings := make([]attr.Value, len(leafToVnBinding))
	i := 0
	for _, v := range leafToVnBinding {
		var b VnBinding
		b.LoadApiData(ctx, v, diags)
		if diags.HasError() {
			return
		}

		var d diag.Diagnostics
		bindings[i], d = types.ObjectValueFrom(ctx, VnBinding{}.attrTypes(), b)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		i++
	}

	o.Bindings = types.SetValueMust(types.ObjectType{AttrTypes: VnBinding{}.attrTypes()}, bindings)
}

// accessSwitchIdsToParentLeafIds returns a map keyed by graph db 'system' node
// IDs representing access switches to slice of graph db 'system' node IDs
// representing leaf switches upstream of each access switch.
func accessSwitchIdsToParentLeafIds(ctx context.Context, ids []string, client *apstra.TwoStageL3ClosClient) (map[string][]string, error) {
	pathQuery := new(apstra.PathQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("system")},
			{Key: "role", Value: apstra.QEStringVal("access")},
			{Key: "name", Value: apstra.QEStringVal("n_access")},
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
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Do(ctx, queryResponse)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, item := range queryResponse.Items {
		if leafIds, ok := result[item.Access.Id]; ok {
			result[item.Access.Id] = append(leafIds, item.Leaf.Id)
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
			return nil, fmt.Errorf("system node with ID %q not found in blueprint %q", systemId, client.Id())
		}
	}
	return result, nil
}
