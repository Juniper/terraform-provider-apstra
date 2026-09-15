package blueprint

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	AccessRGIDs types.Map          `tfsdk:"access_rg_ids"`
}

func (vna *VirtualNetworkAssignment) DatasourceAttributes() map[string]datasourceSchema.Attribute {
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
		"leaf_redundancy_group_id": datasourceSchema.StringAttribute{
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
		"access_ids": datasourceSchema.SetAttribute{
			MarkdownDescription: "",
			Computed:            true,
		},
		"access_rg_ids": datasourceSchema.MapAttribute{
			MarkdownDescription: "",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

func (vna *VirtualNetworkAssignment) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual network ID",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"leaf_switch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Leaf Switch node ID",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"leaf_redundancy_group_id": resourceSchema.StringAttribute{
			MarkdownDescription: "If the Leaf Switch is part of a Redundancy Group, this attribute will contain the Redundancy Group ID. Otherwise `null`.",
			Computed:            true,
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
		"access_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "IDs of Access Switch children of the Leaf Switch. Note that any Access Switch in a Redundancy Group (ESI-LAG) " +
				"must be listed here along with its redundancy partner in order to avoid state churn.",
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"access_rg_ids": resourceSchema.MapAttribute{},
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
