package primitives

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BgpPeeringGenericSystem struct {
	Id                 types.String `tfsdk:"id"`
	BatchId            types.String `tfsdk:"batch_id"`
	PipelineId         types.String `tfsdk:"pipeline_id"`
	Ttl                types.Int64  `tfsdk:"ttl"`
	BfdEnabled         types.Bool   `tfsdk:"bfd_enabled"`
	Password           types.String `tfsdk:"password"`
	KeepaliveTime      types.Int64  `tfsdk:"keepalive_time"`
	HoldTime           types.Int64  `tfsdk:"hold_time"`
	Ipv4AddressingType types.String `tfsdk:"ipv4_addressing_type"`
	Ipv6AddressingType types.String `tfsdk:"ipv6_addressing_type"`
	LocalAsn           types.Int64  `tfsdk:"local_asn"`
	NeighborAsnDynamic types.Bool   `tfsdk:"neighbor_asn_dynamic"`
	PeerFromLoopback   types.Bool   `tfsdk:"peer_from_loopback"`
	PeerTo             types.String `tfsdk:"peer_to"`
	RoutingPolicies    types.Map    `tfsdk:"routing_policies"`
}

func (o BgpPeeringGenericSystem) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                   types.StringType,
		"batch_id":             types.StringType,
		"pipeline_id":          types.StringType,
		"ttl":                  types.Int64Type,
		"bfd_enabled":          types.BoolType,
		"password":             types.StringType,
		"keepalive_time":       types.Int64Type,
		"hold_time":            types.Int64Type,
		"ipv4_addressing_type": types.StringType,
		"ipv6_addressing_type": types.StringType,
		"local_asn":            types.Int64Type,
		"neighbor_asn_dynamic": types.BoolType,
		"peer_from_loopback":   types.BoolType,
		"peer_to":              types.StringType,
		"routing_policies":     types.MapType{ElemType: types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}},
	}
}

func (o BgpPeeringGenericSystem) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive element",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"batch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's downstream collection",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{bgpPeeringGenericSystemBatchIdPlanModifier{}},
		},
		"pipeline_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's upstream pipeline",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"ttl": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP Time To Live. Omit to use device defaults.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(constants.TtlMin, constants.TtlMax)},
		},
		"bfd_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enable BFD.",
			Required:            true,
		},
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "Password used to secure the BGP session.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"keepalive_time": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP keepalive time (seconds).",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(constants.BgpKeepaliveMin, constants.BgpKeepaliveMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("hold_time")),
			},
		},
		"hold_time": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(constants.BgpHoldMin, constants.BgpHoldMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("keepalive_time")),
				apstravalidator.AtLeastProductOf(3, path.MatchRelative().AtParent().AtName("keepalive_time")),
			},
		},
		"ipv4_addressing_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Must be one of: \n  - %s\n", strings.Join([]string{
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv4ProtocolSessionAddressingNone),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv4ProtocolSessionAddressingAddressed),
			}, "\n  - ")),
			Required: true,
			Validators: []validator.String{stringvalidator.OneOf(
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv4ProtocolSessionAddressingNone),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv4ProtocolSessionAddressingAddressed),
			)},
		},
		"ipv6_addressing_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Must be one of: \n  - %s\n", strings.Join([]string{
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6ProtocolSessionAddressingNone),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6ProtocolSessionAddressingAddressed),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6ProtocolSessionAddressingLinkLocal),
			}, "\n  - ")),
			Required: true,
			Validators: []validator.String{stringvalidator.OneOf(
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6ProtocolSessionAddressingNone),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6ProtocolSessionAddressingAddressed),
				utils.StringersToFriendlyString(apstra.CtPrimitiveIPv6ProtocolSessionAddressingLinkLocal),
			)},
		},
		"local_asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "This feature is configured on a per-peer basis. It allows a router " +
				"to appear to be a member of a second autonomous system (AS) by prepending a local-as " +
				"AS number, in addition to its real AS number, announced to its eBGP peer, resulting " +
				"in an AS path length of two.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(1, math.MaxUint32)},
		},
		"neighbor_asn_dynamic": resourceSchema.BoolAttribute{
			MarkdownDescription: "When `true`, the BGP process will accept connections from any peer AS.",
			Required:            true,
		},
		"peer_from_loopback": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enable to peer from loopback interface. Default behavior peers from physical interface.",
			Required:            true,
		},
		"peer_to": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Must be one of: \n  - %s\n", strings.Join(utils.PeerToTypes(), "\n  - ")),
			Required:            true,
			Validators:          []validator.String{stringvalidator.OneOf(utils.PeerToTypes()...)},
		},
		"routing_policies": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Routing Policy Primitives to be used with this *Protocol Endpoint*.",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: RoutingPolicy{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
	}
}

func (o BgpPeeringGenericSystem) attributes(_ context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpOverSubinterfacesOrSvi {
	var holdTime *uint16
	if !o.HoldTime.IsNull() {
		holdTime = utils.ToPtr(uint16(o.HoldTime.ValueInt64()))
	}

	var keepaliveTime *uint16
	if !o.KeepaliveTime.IsNull() {
		keepaliveTime = utils.ToPtr(uint16(o.KeepaliveTime.ValueInt64()))
	}

	var localAsn *uint32
	if !o.LocalAsn.IsNull() {
		localAsn = utils.ToPtr(uint32(o.LocalAsn.ValueInt64()))
	}

	var err error

	var peerTo apstra.CtPrimitiveBgpPeerTo
	err = utils.ApiStringerFromFriendlyString(&peerTo, o.PeerTo.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse peer_to value %s", o.PeerTo), err.Error())
		return nil
	}

	var sessionAddressingIpv4 apstra.CtPrimitiveIPv4ProtocolSessionAddressing
	err = utils.ApiStringerFromFriendlyString(&sessionAddressingIpv4, o.Ipv4AddressingType.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse peer_to value %s", o.Ipv4AddressingType), err.Error())
		return nil
	}

	var sessionAddressingIpv6 apstra.CtPrimitiveIPv6ProtocolSessionAddressing
	err = utils.ApiStringerFromFriendlyString(&sessionAddressingIpv6, o.Ipv6AddressingType.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse peer_to value %s", o.Ipv6AddressingType), err.Error())
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpOverSubinterfacesOrSvi{
		Bfd:                   o.BfdEnabled.ValueBool(),
		Holdtime:              holdTime,
		Ipv4Safi:              o.Ipv4AddressingType.ValueString() != utils.StringersToFriendlyString(enum.InterfaceNumberingIpv4TypeNone),
		Ipv6Safi:              o.Ipv6AddressingType.ValueString() != utils.StringersToFriendlyString(enum.InterfaceNumberingIpv6TypeNone),
		Keepalive:             keepaliveTime,
		LocalAsn:              localAsn,
		NeighborAsnDynamic:    o.NeighborAsnDynamic.ValueBool(),
		Password:              o.Password.ValueStringPointer(),
		PeerFromLoopback:      o.PeerFromLoopback.ValueBool(),
		PeerTo:                peerTo,
		SessionAddressingIpv4: sessionAddressingIpv4,
		SessionAddressingIpv6: sessionAddressingIpv6,
		Ttl:                   uint8(o.Ttl.ValueInt64()), // okay if null, then we get zero value
	}
}

func (o BgpPeeringGenericSystem) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	result := apstra.ConnectivityTemplatePrimitive{Attributes: o.attributes(ctx, diags)}

	if !o.PipelineId.IsUnknown() {
		result.PipelineId = (*apstra.ObjectId)(o.PipelineId.ValueStringPointer()) // nil when null
	}
	if !o.Id.IsUnknown() {
		result.Id = (*apstra.ObjectId)(o.Id.ValueStringPointer()) // nil when null
	}
	if !o.BatchId.IsUnknown() {
		result.BatchId = (*apstra.ObjectId)(o.BatchId.ValueStringPointer()) // nil when null
	}

	result.Subpolicies = append(result.Subpolicies, RoutingPolicySubpolicies(ctx, o.RoutingPolicies, diags)...)

	return &result
}

func BgpPeeringGenericSystemSubpolicies(ctx context.Context, bgpPeeringGenericSystemMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var bgpPeeringGenericSystems map[string]BgpPeeringGenericSystem
	diags.Append(bgpPeeringGenericSystemMap.ElementsAs(ctx, &bgpPeeringGenericSystems, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(bgpPeeringGenericSystems))
	i := 0
	for k, v := range bgpPeeringGenericSystems {
		subpolicies[i] = v.primitive(ctx, diags)
		if diags.HasError() {
			return nil
		}
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newBgpPeeringGenericSystem(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpOverSubinterfacesOrSvi, _ *diag.Diagnostics) BgpPeeringGenericSystem {
	result := BgpPeeringGenericSystem{
		// Name:            // handled by caller
		// Ttl:             // handled below due to 0 = null logic
		BfdEnabled:         types.BoolValue(in.Bfd),
		Password:           types.StringPointerValue(in.Password),
		KeepaliveTime:      utils.Int64PointerValue(in.Keepalive),
		HoldTime:           utils.Int64PointerValue(in.Holdtime),
		Ipv4AddressingType: types.StringValue(utils.StringersToFriendlyString(in.SessionAddressingIpv4)),
		Ipv6AddressingType: types.StringValue(utils.StringersToFriendlyString(in.SessionAddressingIpv6)),
		// LocalAsn:        // handled below
		NeighborAsnDynamic: types.BoolValue(in.NeighborAsnDynamic),
		PeerFromLoopback:   types.BoolValue(in.PeerFromLoopback),
		PeerTo:             types.StringValue(utils.StringersToFriendlyString(in.PeerTo)),
		// RoutingPolicies: handled by caller
	}

	if in.Ttl > 0 {
		result.Ttl = types.Int64Value(int64(in.Ttl))
	}

	if in.LocalAsn != nil {
		result.LocalAsn = types.Int64Value(int64(*in.LocalAsn))
	}

	return result
}

func BgpPeeringGenericSystemPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]BgpPeeringGenericSystem)

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpOverSubinterfacesOrSvi); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting BgpPeeringGenericSystem primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newBgpPeeringGenericSystem(ctx, p, diags)
			newPrimitive.PipelineId = types.StringPointerValue((*string)(subpolicy.PipelineId))
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			newPrimitive.BatchId = types.StringPointerValue((*string)(subpolicy.BatchId))
			newPrimitive.RoutingPolicies = RoutingPolicyPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()}, result, diags)
}

func LoadIDsIntoBgpPeeringGenericSystemMap(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, inMap types.Map, diags *diag.Diagnostics) types.Map {
	result := make(map[string]BgpPeeringGenericSystem, len(inMap.Elements()))
	inMap.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()})
	}

	for _, p := range subpolicies {
		if _, ok := p.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpOverSubinterfacesOrSvi); !ok {
			continue // wrong type and nil value both wind up getting skipped
		}

		if v, ok := result[p.Label]; ok {
			v.PipelineId = types.StringPointerValue((*string)(p.PipelineId))
			v.Id = types.StringPointerValue((*string)(p.Id))
			v.BatchId = types.StringPointerValue((*string)(p.BatchId))
			v.RoutingPolicies = LoadIDsIntoRoutingPolicyMap(ctx, p.Subpolicies, v.RoutingPolicies, diags)
			result[p.Label] = v
		}
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()}, result, diags)
}

var _ planmodifier.String = (*bgpPeeringGenericSystemBatchIdPlanModifier)(nil)

type bgpPeeringGenericSystemBatchIdPlanModifier struct{}

func (o bgpPeeringGenericSystemBatchIdPlanModifier) Description(_ context.Context) string {
	return "preserves the the state value unless we're transitioning between zero and non-zero child primitives, in which case null or unknown is planned"
}

func (o bgpPeeringGenericSystemBatchIdPlanModifier) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o bgpPeeringGenericSystemBatchIdPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var plan, state BgpPeeringGenericSystem

	// unpacking the parent object's planned value should always work
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, req.Path.ParentPath(), &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// attempting to unpack the parent object's state indicates whether state *exists*
	d := req.State.GetAttribute(ctx, req.Path.ParentPath(), &state)
	stateDoesNotExist := d.HasError()

	planHasChildren := len(plan.RoutingPolicies.Elements()) > 0

	planChildrenUnknown := plan.RoutingPolicies.IsUnknown()

	// are we a new object?
	if stateDoesNotExist {
		if planHasChildren || planChildrenUnknown {
			resp.PlanValue = types.StringUnknown()
		} else {
			resp.PlanValue = types.StringNull()
		}
		return
	}

	stateHasChildren := len(state.RoutingPolicies.Elements()) > 0

	if (planHasChildren || planChildrenUnknown) == stateHasChildren {
		// state and plan agree about whether a batch ID is required. Reuse the old value.
		resp.PlanValue = req.StateValue
		return
	}

	// We've either gained our first, or lost our last child primitive. Set the plan value accordingly.
	if planHasChildren || planChildrenUnknown {
		resp.PlanValue = types.StringUnknown()
	} else {
		resp.PlanValue = types.StringNull()
	}
}
