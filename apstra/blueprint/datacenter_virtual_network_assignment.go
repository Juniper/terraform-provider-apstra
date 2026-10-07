package blueprint

import (
	"context"
	"fmt"
	"maps"
	"net"
	"slices"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	cache "github.com/Juniper/terraform-provider-apstra/internal/system_redundancy_cache"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type VirtualNetworkAssignment struct {
	BlueprintID types.String         `tfsdk:"blueprint_id"`
	VNID        types.String         `tfsdk:"virtual_network_id"`
	LeafID      types.String         `tfsdk:"leaf_switch_id"`
	VLAN        types.Int64          `tfsdk:"vlan"`
	IPv4Mode    types.String         `tfsdk:"ipv4_mode"`
	IPv4Address cidrtypes.IPv4Prefix `tfsdk:"ipv4_address"`
	IPv6Mode    types.String         `tfsdk:"ipv6_mode"`
	IPv6Address cidrtypes.IPv6Prefix `tfsdk:"ipv6_address"`
	AccessIDs   types.Set            `tfsdk:"access_switch_ids"`
}

func (vna VirtualNetworkAssignment) DatasourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"blueprint_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"virtual_network_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"leaf_switch_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"vlan": datasourceSchema.Int64Attribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"ipv4_mode": datasourceSchema.StringAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"ipv4_address": datasourceSchema.StringAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"ipv6_mode": datasourceSchema.StringAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"ipv6_address": datasourceSchema.StringAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"access_switch_ids": datasourceSchema.SetAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
	}
}

func (vna VirtualNetworkAssignment) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Blueprint in which this Virtual Network Assignment is being configured.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Virtual Network to be assigned to a Leaf Switch.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"leaf_switch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Leaf Switch to which the Virtual Network is being assigned.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"vlan": resourceSchema.Int64Attribute{
			MarkdownDescription: "VLAN to use when associating the Virtual Network with this Leaf Switch.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(design.VlanMin, design.VlanMax)},
		},
		"ipv4_mode": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("When and how to render a unique IPv4 address for this Leaf Switch on this Virtual Network. "+
				"One of: `%s`. Note that this attribute is being written into the *config* graph. Data in the *staging* graph may not agree depending "+
				"on what other features have been configured in the Blueprint. For example, attaching a Connectivity Template which requires the Leaf "+
				"Switch to peer via BGP with a Generic System may cause the *staging* graph to reflect `%s` when `%s` was configured via this resource. "+
				"Furthermore, features configured in the Blueprint (like attaching a Connecitivty Template) may constrain the values which may be "+
				"configured via this attribute. Requires IPv4 connectivity to be enabled in the Virtual Network.",
				strings.Join(enum.IPv4SVIModes.Values(), "`, `"), enum.IPv4SVIModeForced, enum.IPv4SVIModeEnabled),
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.OneOf(enum.IPv4SVIModes.Values()...)},
		},
		"ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("An IPv4 address to configure for this Leaf Switch on this Virtual Network. Not compatible with "+
				"`ipv4_mode = %q`. Requires IPv4 to be enabled on the Virtual Network.", enum.IPv4SVIModeDisabled),
			CustomType: cidrtypes.IPv4PrefixType{},
			Optional:   true,
			Validators: []validator.String{
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("ipv4_mode"), types.StringValue(enum.IPv4SVIModeDisabled.String())),
			},
		},
		"ipv6_mode": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("When and how to render a unique IPv6 address for this Leaf Switch on this Virtual Network. "+
				"One of: `%s`. Note that this attribute is being written into the *config* graph. Data in the *staging* graph may not agree depending "+
				"on what other features have been configured in the Blueprint. For example, attaching a Connectivity Template which requires the Leaf "+
				"Switch to peer via BGP with a Generic System may cause the *staging* graph to reflect `%s` when `%s` was configured via this resource. "+
				"Furthermore, features configured in the Blueprint (like attaching a Connecitivty Template) may constrain the values which may be "+
				"configured via this attribute. Requires IPv6 connectivity to be enabled in the Virtual Network.",
				strings.Join(enum.IPv6SVIModes.Values(), "`, `"), enum.IPv6SVIModeForced, enum.IPv6SVIModeEnabled),
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.OneOf(enum.IPv6SVIModes.Values()...)},
		},
		"ipv6_address": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("An IPv6 address to configure for this Leaf Switch on this Virtual Network. Not compatible with "+
				"`ipv6_mode = %q` or `ipv6_mode = %q`. Requires IPv6 to be enabled on the Virtual Network.", enum.IPv6SVIModeDisabled, enum.IPv6SVIModeLinkLocal),
			CustomType: cidrtypes.IPv6PrefixType{},
			Optional:   true,
			Validators: []validator.String{
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("ipv6_mode"), types.StringValue(enum.IPv6SVIModeDisabled.String())),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("ipv6_mode"), types.StringValue(enum.IPv6SVIModeLinkLocal.String())),
			},
		},
		"access_switch_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "IDs of Access Switch children of the Leaf Switch. Note that any Access Switch in a Redundancy Group (ESI-LAG) " +
				"must be listed here along with its redundancy partner in order to avoid state churn.",
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

//// FetchLeafRedundancyGroupID queries the Blueprint for the Leaf Switch and its Redundancy Group (if any)
//// and sets the LeafRGID field accordingly. If the Leaf Switch is not part of a Redundancy Group, LeafRGID
//// will be set to null. This function should be run only once: Early during Create()
//func (vna *VirtualNetworkAssignment) FetchLeafRedundancyGroupID(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	query := new(apstra.MatchQuery).
//		SetClient(bp.Client()).
//		SetBlueprintId(bp.Id()).
//		Match(new(apstra.PathQuery).
//			Node([]apstra.QEEAttribute{
//				apstra.NodeTypeSystem.QEEAttribute(),
//				{Key: "id", Value: apstra.QEStringVal(vna.LeafID.ValueString())},
//				{Key: "name", Value: apstra.QEStringVal("leaf")},
//			}),
//		).Optional(
//		new(apstra.PathQuery).
//			Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("leaf")}}).
//			Out([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRedundancyGroup.QEEAttribute()}).
//			Node([]apstra.QEEAttribute{
//				apstra.NodeTypeRedundancyGroup.QEEAttribute(),
//				{Key: "name", Value: apstra.QEStringVal("redundancy_group")},
//			}),
//	)
//
//	var target struct {
//		Items []struct {
//			Leaf struct {
//				Type       string `tfsdk:"type"`
//				SystemType string `tfsdk:"system_type"`
//				Role       string `tfsdk:"role"`
//				ID         string `json:"id"`
//			} `json:"leaf"`
//			RedundancyGroup *struct {
//				ID string `json:"id"`
//			} `json:"redundancy_group"`
//		} `json:"items"`
//	}
//
//	err := query.Do(ctx, &target)
//	if err != nil {
//		diags.AddError("failed while checking for leaf redundancy group", err.Error())
//		return
//	}
//
//	switch len(target.Items) {
//	case 0:
//		diags.AddError("failed while checking for leaf redundancy group", "leaf switch not found")
//	case 1: // Expected case falls through.
//	default:
//		diags.AddError("failed while checking for leaf redundancy group", "found multiple matches")
//	}
//	if diags.HasError() {
//		return
//	}
//
//	leaf := &target.Items[0].Leaf
//	group := target.Items[0].RedundancyGroup
//
//	switch {
//	case leaf.Type != apstra.NodeTypeSystem.String():
//		diags.AddError("failed while checking for leaf redundancy group", "leaf switch node has type "+leaf.Type+", expected "+apstra.NodeTypeSystem.String())
//	case leaf.SystemType != enum.SystemTypeSwitch.String():
//		diags.AddError("failed while checking for leaf redundancy group", "leaf switch node has system_type "+leaf.SystemType+", expected "+enum.SystemTypeSwitch.String())
//	case leaf.Role != enum.SystemNodeRoleLeaf.String():
//		diags.AddError("failed while checking for leaf redundancy group", "leaf switch node has role "+leaf.Role+", expected "+enum.SystemNodeRoleLeaf.String())
//	case leaf.ID == "":
//		diags.AddError("failed while checking for leaf redundancy group", "leaf switch node has empty ID")
//	case group != nil && group.ID == "":
//		diags.AddError("failed while checking for leaf redundancy group", "leaf switch redundancy group node has empty ID")
//	}
//	if diags.HasError() {
//		return
//	}
//
//	if group == nil {
//		vna.LeafRGID = types.StringNull()
//	} else {
//		vna.LeafRGID = types.StringValue(group.ID)
//	}
//}

func (vna VirtualNetworkAssignment) Request(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) apstra.VirtualNetworkBindingsRequest {
	// Check the system redundancy cache for a group ID. If one exists, use it instead of the leaf ID.
	bindTo := vna.LeafID.ValueString()
	if groupID := cache.LookupGroup(ctx, bp, bindTo, diags); groupID != nil {
		bindTo = *groupID
	}
	if diags.HasError() {
		return apstra.VirtualNetworkBindingsRequest{}
	}

	// Collect the Access Switch IDs, replacing any individual Access Switch IDs with their Redundancy Group ID if applicable.
	accessIDs := make(map[string]struct{}, len(vna.AccessIDs.Elements()))
	for _, v := range vna.AccessIDs.Elements() {
		accessID := v.(basetypes.StringValue).ValueString()
		if group := cache.LookupGroup(ctx, bp, accessID, diags); group == nil {
			accessIDs[accessID] = struct{}{}
		} else {
			accessIDs[*group] = struct{}{}
		}

	}

	// If any of the SVI attributes are set, we need to build the SVIAddressing map.
	// If none of the SVI attributes are set, we can leave the SVIAddressing map nil.
	var sviIPs map[apstra.ObjectId]*datacenter.SVIAddressing
	if utils.HasValue(vna.IPv4Mode) || !vna.IPv4Address.IsNull() ||
		utils.HasValue(vna.IPv6Mode) || !vna.IPv6Address.IsNull() {
		ipv4Mode := &enum.IPv4SVIModeDisabled
		if !vna.IPv4Mode.IsNull() {
			if ipv4Mode = enum.IPv4SVIModes.Parse(vna.IPv4Mode.ValueString()); ipv4Mode == nil {
				diags.AddError("invalid ipv4_mode", fmt.Sprintf("invalid value %q for ipv4_mode", vna.IPv4Mode.ValueString()))
				return apstra.VirtualNetworkBindingsRequest{} // This should never happen because of validation specified in the schema.
			}
		}

		ipv6Mode := &enum.IPv6SVIModeDisabled
		if !vna.IPv6Mode.IsNull() {
			if ipv6Mode = enum.IPv6SVIModes.Parse(vna.IPv6Mode.ValueString()); ipv6Mode == nil {
				diags.AddError("invalid ipv6_mode", fmt.Sprintf("invalid value %q for ipv6_mode", vna.IPv6Mode.ValueString()))
				return apstra.VirtualNetworkBindingsRequest{} // This should never happen because of validation specified in the schema.
			}
		}

		var ipv4Addr *net.IPNet
		if !vna.IPv4Address.IsNull() {
			ip, ipNet, err := net.ParseCIDR(vna.IPv4Address.ValueString())
			if err != nil {
				diags.AddError("invalid ipv4_address", fmt.Sprintf("invalid value %q for ipv4_address: %s", vna.IPv4Address.ValueString(), err.Error()))
				return apstra.VirtualNetworkBindingsRequest{} // This should never happen because of validation specified in the schema.
			}
			ipv4Addr = &net.IPNet{IP: ip, Mask: ipNet.Mask}
		}

		var ipv6Addr *net.IPNet
		if !vna.IPv6Address.IsNull() {
			ip, ipNet, err := net.ParseCIDR(vna.IPv6Address.ValueString())
			if err != nil {
				diags.AddError("invalid ipv6_address", fmt.Sprintf("invalid value %q for ipv6_address: %s", vna.IPv6Address.ValueString(), err.Error()))
				return apstra.VirtualNetworkBindingsRequest{} // This should never happen because of validation specified in the schema.
			}
			ipv6Addr = &net.IPNet{IP: ip, Mask: ipNet.Mask}
		}

		sviIPs = map[apstra.ObjectId]*datacenter.SVIAddressing{
			apstra.ObjectId(vna.LeafID.ValueString()): {
				SystemID: vna.LeafID.ValueString(),
				IPv4Addr: ipv4Addr,
				IPv4Mode: *ipv4Mode,
				IPv6Addr: ipv6Addr,
				IPv6Mode: *ipv6Mode,
			},
		}
	}

	return apstra.VirtualNetworkBindingsRequest{
		VnId:   apstra.ObjectId(vna.VNID.ValueString()),
		SviIps: sviIPs,
		VnBindings: map[apstra.ObjectId]*datacenter.VNBinding{
			apstra.ObjectId(bindTo): {
				AccessSwitchNodeIDs: slices.Collect(maps.Keys(accessIDs)),
				SystemID:            bindTo,
				VLAN:                pointer.ConvertInteger(new(uint16), vna.VLAN.ValueInt64Pointer()),
			},
		},
	}
}

// Read queries the Blueprint for the Virtual Network and its bindings, and populates the VirtualNetworkAssignment
// struct with data returned by the API. If the assigment (binding) is not found, the function returns false. If
// the assignment is found, the function returns true. If an error occurs during the API call, the function returns
// false and adds an error to the diagnostics.
func (vna *VirtualNetworkAssignment) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) bool {
	vn, err := bp.GetVirtualNetwork(ctx, vna.VNID.ValueString())
	if err != nil {
		diags.AddError("error reading virtual network", err.Error())
		return false
	}

	// If the leaf switch is part of a redundancy group, the binding will be associated with the
	// group ID rather than the leaf switch ID. Use whichever string is appropriate, depending on
	// whether the leaf switch is part of a redundancy group or not.
	boundTo := vna.LeafID.ValueString()
	if groupID := cache.LookupGroup(ctx, bp, boundTo, diags); groupID != nil {
		boundTo = *groupID
	}
	if diags.HasError() {
		return false
	}

	// Find the relevant binding from the API response.
	var binding *datacenter.VNBinding
	for _, b := range vn.Bindings {
		if boundTo == b.SystemID {
			binding = &b
			break // We found the correct binding.
		}
	}

	if binding == nil {
		return false // No binding found. False will signal Read() to remove the resource from state.
	}

	// Set the VLAN attribute.
	vna.VLAN = types.Int64PointerValue(pointer.ConvertInteger(new(int64), binding.VLAN))

	// Set the Access Switch IDs attribute.
	accessIDs := make([]string, 0, 2*len(binding.AccessSwitchNodeIDs))
	for _, accessID := range binding.AccessSwitchNodeIDs {
		systems, ok := cache.LookupSystems(ctx, bp, vna.LeafID.ValueString(), diags)
		if diags.HasError() {
			return false
		}

		if ok {
			// accessID is a group ID. Add both the ID of both member system to our slice.
			accessIDs = append(accessIDs, systems[:]...)
		} else {
			// accessID is an individual switch ID. Add it to our slice.
			accessIDs = append(accessIDs, accessID)
		}
	}
	vna.AccessIDs = value.SetOrNull(ctx, types.StringType, accessIDs, diags)

	// Set the SVI attributes.
	for _, sviAddressing := range vn.SVIIPs {
		if vna.LeafID.ValueString() != sviAddressing.SystemID {
			continue // SVI info represents some other leaf switch.
		}

		vna.IPv4Mode = types.StringValue(sviAddressing.IPv4Mode.String())
		if sviAddressing.IPv4Addr == nil {
			vna.IPv4Address = cidrtypes.NewIPv4PrefixNull()
		} else {
			vna.IPv4Address = cidrtypes.NewIPv4PrefixValue(sviAddressing.IPv4Addr.String())
		}

		vna.IPv6Mode = types.StringValue(sviAddressing.IPv6Mode.String())
		if sviAddressing.IPv6Addr == nil {
			vna.IPv6Address = cidrtypes.NewIPv6PrefixNull()
		} else {
			vna.IPv6Address = cidrtypes.NewIPv6PrefixValue(sviAddressing.IPv6Addr.String())
		}
	}

	return true
}

func (vna VirtualNetworkAssignment) Identity() VirtualNetworkAssignmentIdentity {
	return VirtualNetworkAssignmentIdentity{
		BlueprintID: vna.BlueprintID,
		VNID:        vna.VNID,
		LeafID:      vna.LeafID,
	}
}
