package primitives

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DynamicBgpPeering struct {
	Id              types.String         `tfsdk:"id"`
	BatchId         types.String         `tfsdk:"batch_id"`
	PipelineId      types.String         `tfsdk:"pipeline_id"`
	Ttl             types.Int64          `tfsdk:"ttl"`
	BfdEnabled      types.Bool           `tfsdk:"bfd_enabled"`
	Password        types.String         `tfsdk:"password"`
	KeepaliveTime   types.Int64          `tfsdk:"keepalive_time"`
	HoldTime        types.Int64          `tfsdk:"hold_time"`
	Ipv4Enabled     types.Bool           `tfsdk:"ipv4_enabled"`
	Ipv6Enabled     types.Bool           `tfsdk:"ipv6_enabled"`
	LocalAsn        types.Int64          `tfsdk:"local_asn"`
	Ipv4PeerPrefix  cidrtypes.IPv4Prefix `tfsdk:"ipv4_peer_prefix"`
	Ipv6PeerPrefix  cidrtypes.IPv6Prefix `tfsdk:"ipv6_peer_prefix"`
	RoutingPolicies types.Map            `tfsdk:"routing_policies"`
}

func (o DynamicBgpPeering) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"batch_id":         types.StringType,
		"pipeline_id":      types.StringType,
		"ttl":              types.Int64Type,
		"bfd_enabled":      types.BoolType,
		"password":         types.StringType,
		"keepalive_time":   types.Int64Type,
		"hold_time":        types.Int64Type,
		"ipv4_enabled":     types.BoolType,
		"ipv6_enabled":     types.BoolType,
		"local_asn":        types.Int64Type,
		"ipv4_peer_prefix": cidrtypes.IPv4PrefixType{},
		"ipv6_peer_prefix": cidrtypes.IPv6PrefixType{},
		"routing_policies": types.MapType{ElemType: types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}},
	}
}

func (o DynamicBgpPeering) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive element",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"batch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's downstream collection",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{dynamicBgpPeeringBatchIdPlanModifier{}},
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
		"ipv4_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables peering with IPv4 neighbors.",
			Required:            true,
		},
		"ipv6_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables peering with IPv6 neighbors.",
			Required:            true,
		},
		"local_asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "This feature is configured on a per-peer basis. It allows a router " +
				"to appear to be a member of a second autonomous system (AS) by prepending a local-as " +
				"AS number, in addition to its real AS number, announced to its eBGP peer, resulting " +
				"in an AS path length of two.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(1, math.MaxUint32)},
		},
		"ipv4_peer_prefix": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 Subnet for BGP Prefix Dynamic Neighbors. Leave blank to derive subnet from application point.",
			CustomType:          cidrtypes.IPv4PrefixType{},
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("ipv4_enabled"), types.BoolNull()),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("ipv4_enabled"), types.BoolValue(false)),
			},
		},
		"ipv6_peer_prefix": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 Subnet for BGP Prefix Dynamic Neighbors. Leave blank to derive subnet from application point.",
			CustomType:          cidrtypes.IPv6PrefixType{},
			Optional:            true,
			Validators: []validator.String{
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("ipv6_enabled"), types.BoolNull()),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRelative().AtParent().AtName("ipv6_enabled"), types.BoolValue(false)),
			},
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

func (o DynamicBgpPeering) ValidateConfig(_ context.Context, path path.Path, diags *diag.Diagnostics) {
	if !o.Ipv4Enabled.ValueBool() && !o.Ipv6Enabled.ValueBool() {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			path, "at least one of 'ipv4_enabled' and 'ipv6_enabled' must be true.",
		))
	}
}

func (o DynamicBgpPeering) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface {
	var holdTime *uint16
	if !o.HoldTime.IsNull() {
		holdTime = utils.ToPtr(uint16(o.HoldTime.ValueInt64()))
	}

	var ipv4PeerPrefix *net.IPNet
	if !o.Ipv4PeerPrefix.IsNull() {
		_, ipv4PeerPrefix, _ = net.ParseCIDR(o.Ipv4PeerPrefix.ValueString())
	}

	var ipv6PeerPrefix *net.IPNet
	if !o.Ipv6PeerPrefix.IsNull() {
		_, ipv6PeerPrefix, _ = net.ParseCIDR(o.Ipv6PeerPrefix.ValueString())
	}

	var keepaliveTime *uint16
	if !o.KeepaliveTime.IsNull() {
		keepaliveTime = utils.ToPtr(uint16(o.KeepaliveTime.ValueInt64()))
	}

	var localAsn *uint32
	if !o.LocalAsn.IsNull() {
		localAsn = utils.ToPtr(uint32(o.LocalAsn.ValueInt64()))
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface{
		Bfd:                   o.BfdEnabled.ValueBool(),
		Holdtime:              holdTime,
		Ipv4Safi:              o.Ipv4Enabled.ValueBool(),
		Ipv6Safi:              o.Ipv6Enabled.ValueBool(),
		Keepalive:             keepaliveTime,
		LocalAsn:              localAsn,
		Password:              o.Password.ValueStringPointer(),
		PrefixNeighborIpv4:    ipv4PeerPrefix,
		PrefixNeighborIpv6:    ipv6PeerPrefix,
		SessionAddressingIpv4: o.Ipv4Enabled.ValueBool(),
		SessionAddressingIpv6: o.Ipv6Enabled.ValueBool(),
		Ttl:                   uint8(o.Ttl.ValueInt64()), // okay if null, then we get zero value
	}
}

func (o DynamicBgpPeering) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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

func DynamicBgpPeeringSubpolicies(ctx context.Context, dynamicBgpPeeringMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var dynamicBgpPeerings map[string]DynamicBgpPeering
	diags.Append(dynamicBgpPeeringMap.ElementsAs(ctx, &dynamicBgpPeerings, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(dynamicBgpPeerings))
	i := 0
	for k, v := range dynamicBgpPeerings {
		subpolicies[i] = v.primitive(ctx, diags)
		if diags.HasError() {
			return nil
		}
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newDynamicBgpPeering(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface, _ *diag.Diagnostics) DynamicBgpPeering {
	result := DynamicBgpPeering{
		// Name:        // handled by caller
		// Ttl:         // handled below due to 0 = null logic
		BfdEnabled:     types.BoolValue(in.Bfd),
		Password:       types.StringPointerValue(in.Password),
		KeepaliveTime:  utils.Int64PointerValue(in.Keepalive),
		HoldTime:       utils.Int64PointerValue(in.Holdtime),
		Ipv4Enabled:    types.BoolValue(in.SessionAddressingIpv4),
		Ipv6Enabled:    types.BoolValue(in.SessionAddressingIpv6),
		LocalAsn:       utils.Int64PointerValue(in.LocalAsn),
		Ipv4PeerPrefix: utils.Ipv4PrefixPointerValue(in.PrefixNeighborIpv4),
		Ipv6PeerPrefix: utils.Ipv6PrefixPointerValue(in.PrefixNeighborIpv6),
		// RoutingPolicies: types.Set{}, // set after this function is invoked
	}

	if in.Ttl > 0 {
		result.Ttl = types.Int64Value(int64(in.Ttl))
	}

	return result
}

func DynamicBgpPeeringPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]DynamicBgpPeering)

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting DynamicBgpPeering primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newDynamicBgpPeering(ctx, p, diags)
			newPrimitive.PipelineId = types.StringPointerValue((*string)(subpolicy.PipelineId))
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			newPrimitive.BatchId = types.StringPointerValue((*string)(subpolicy.BatchId))
			newPrimitive.RoutingPolicies = RoutingPolicyPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: DynamicBgpPeering{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: DynamicBgpPeering{}.AttrTypes()}, result, diags)
}

func LoadIDsIntoDynamicBgpPeeringMap(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, inMap types.Map, diags *diag.Diagnostics) types.Map {
	result := make(map[string]DynamicBgpPeering, len(inMap.Elements()))
	inMap.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: DynamicBgpPeering{}.AttrTypes()})
	}

	for _, p := range subpolicies {
		if _, ok := p.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachBgpWithPrefixPeeringForSviOrSubinterface); !ok {
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

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: DynamicBgpPeering{}.AttrTypes()}, result, diags)
}

var _ planmodifier.String = (*dynamicBgpPeeringBatchIdPlanModifier)(nil)

type dynamicBgpPeeringBatchIdPlanModifier struct{}

func (o dynamicBgpPeeringBatchIdPlanModifier) Description(_ context.Context) string {
	return "preserves the the state value unless we're transitioning between zero and non-zero child primitives, in which case null or unknown is planned"
}

func (o dynamicBgpPeeringBatchIdPlanModifier) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o dynamicBgpPeeringBatchIdPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var plan, state DynamicBgpPeering

	// unpacking the parent object's planned value should always work
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, req.Path.ParentPath(), &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// attempting to unpack the parent object's state indicates whether state *exists*
	d := req.State.GetAttribute(ctx, req.Path.ParentPath(), &state)
	stateDoesNotExist := d.HasError()

	planHasChildren := len(plan.RoutingPolicies.Elements()) > 0

	// are we a new object?
	if stateDoesNotExist {
		if planHasChildren {
			resp.PlanValue = types.StringUnknown()
		} else {
			resp.PlanValue = types.StringNull()
		}
		return
	}

	stateHasChildren := len(state.RoutingPolicies.Elements()) > 0

	if planHasChildren == stateHasChildren {
		// state and plan agree about whether a batch ID is required. Reuse the old value.
		resp.PlanValue = req.StateValue
		return
	}

	// We've either gained our first, or lost our last child primitive. Set the plan value accordingly.
	if planHasChildren {
		resp.PlanValue = types.StringUnknown()
	} else {
		resp.PlanValue = types.StringNull()
	}
}
