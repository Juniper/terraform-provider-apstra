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
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
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

type BgpPeeringIpEndpoint struct {
	Id              types.String        `tfsdk:"id"`
	BatchId         types.String        `tfsdk:"batch_id"`
	PipelineId      types.String        `tfsdk:"pipeline_id"`
	NeighborAsn     types.Int64         `tfsdk:"neighbor_asn"`
	Ttl             types.Int64         `tfsdk:"ttl"`
	BfdEnabled      types.Bool          `tfsdk:"bfd_enabled"`
	Password        types.String        `tfsdk:"password"`
	KeepaliveTime   types.Int64         `tfsdk:"keepalive_time"`
	HoldTime        types.Int64         `tfsdk:"hold_time"`
	LocalAsn        types.Int64         `tfsdk:"local_asn"`
	Ipv4Address     iptypes.IPv4Address `tfsdk:"ipv4_address"`
	Ipv6Address     iptypes.IPv6Address `tfsdk:"ipv6_address"`
	RoutingPolicies types.Map           `tfsdk:"routing_policies"`
}

func (o BgpPeeringIpEndpoint) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"batch_id":         types.StringType,
		"pipeline_id":      types.StringType,
		"neighbor_asn":     types.Int64Type,
		"ttl":              types.Int64Type,
		"bfd_enabled":      types.BoolType,
		"password":         types.StringType,
		"keepalive_time":   types.Int64Type,
		"hold_time":        types.Int64Type,
		"local_asn":        types.Int64Type,
		"ipv4_address":     iptypes.IPv4AddressType{},
		"ipv6_address":     iptypes.IPv6AddressType{},
		"routing_policies": types.MapType{ElemType: types.ObjectType{AttrTypes: RoutingPolicy{}.AttrTypes()}},
	}
}

func (o BgpPeeringIpEndpoint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive element",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"batch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's downstream collection",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{bgpPeeringIpEndpointBatchIdPlanModifier{}},
		},
		"pipeline_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's upstream pipeline",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"neighbor_asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "Neighbor ASN. Omit for *Neighbor ASN Type Dynamic*.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(constants.AsnMin, constants.AsnMax)},
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
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("hold_time").Resolve()),
			},
		},
		"hold_time": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(constants.BgpHoldMin, constants.BgpHoldMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("keepalive_time").Resolve()),
				apstravalidator.AtLeastProductOf(3, path.MatchRelative().AtParent().AtName("keepalive_time").Resolve()),
			},
		},
		"local_asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "This feature is configured on a per-peer basis. It allows a router " +
				"to appear to be a member of a second autonomous system (AS) by prepending a local-as " +
				"AS number, in addition to its real AS number, announced to its eBGP peer, resulting " +
				"in an AS path length of two.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(1, math.MaxUint32)},
		},
		"ipv4_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address of peer.",
			CustomType:          iptypes.IPv4AddressType{},
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.AtLeastOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRelative().AtParent().AtName("ipv6_address").Resolve(),
				}...),
			},
		},
		"ipv6_address": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address of peer.",
			CustomType:          iptypes.IPv6AddressType{},
			Optional:            true,
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

func (o BgpPeeringIpEndpoint) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt {
	var neighborAsn *uint32
	if !o.NeighborAsn.IsNull() {
		neighborAsn = utils.ToPtr(uint32(o.NeighborAsn.ValueInt64()))
	}

	var holdTime *uint16
	if !o.HoldTime.IsNull() {
		holdTime = utils.ToPtr(uint16(o.HoldTime.ValueInt64()))
	}

	var ipv4Addr net.IP
	if !o.Ipv4Address.IsNull() {
		ipv4Addr = net.ParseIP(o.Ipv4Address.ValueString())
	}

	var ipv6Addr net.IP
	if !o.Ipv6Address.IsNull() {
		ipv6Addr = net.ParseIP(o.Ipv6Address.ValueString())
	}

	var keepaliveTime *uint16
	if !o.KeepaliveTime.IsNull() {
		keepaliveTime = utils.ToPtr(uint16(o.KeepaliveTime.ValueInt64()))
	}

	var localAsn *uint32
	if !o.LocalAsn.IsNull() {
		localAsn = utils.ToPtr(uint32(o.LocalAsn.ValueInt64()))
	}

	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt{
		Asn:                neighborAsn,
		Bfd:                o.BfdEnabled.ValueBool(),
		Holdtime:           holdTime,
		Ipv4Addr:           ipv4Addr,
		Ipv6Addr:           ipv6Addr,
		Ipv4Safi:           !o.Ipv4Address.IsNull(),
		Ipv6Safi:           !o.Ipv6Address.IsNull(),
		Keepalive:          keepaliveTime,
		LocalAsn:           localAsn,
		NeighborAsnDynamic: o.NeighborAsn.IsNull(),
		Password:           o.Password.ValueStringPointer(),
		Ttl:                uint8(o.Ttl.ValueInt64()), // okay if null, then we get zero value
	}
}

func (o BgpPeeringIpEndpoint) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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

func BgpPeeringIpEndpointSubpolicies(ctx context.Context, bgpPeeringIpEndpointMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var bgpPeeringIpEndpoints map[string]BgpPeeringIpEndpoint
	diags.Append(bgpPeeringIpEndpointMap.ElementsAs(ctx, &bgpPeeringIpEndpoints, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(bgpPeeringIpEndpoints))
	i := 0
	for k, v := range bgpPeeringIpEndpoints {
		subpolicies[i] = v.primitive(ctx, diags)
		if diags.HasError() {
			return nil
		}
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newBgpPeeringIpEndpoint(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt, _ *diag.Diagnostics) BgpPeeringIpEndpoint {
	result := BgpPeeringIpEndpoint{
		// Name:       // handled by caller
		// Ttl:        // handled below due to 0 = null logic
		NeighborAsn:   utils.Int64PointerValue(in.Asn),
		BfdEnabled:    types.BoolValue(in.Bfd),
		Password:      types.StringPointerValue(in.Password),
		KeepaliveTime: utils.Int64PointerValue(in.Keepalive),
		HoldTime:      utils.Int64PointerValue(in.Holdtime),
		LocalAsn:      utils.Int64PointerValue(in.LocalAsn),
		Ipv4Address:   utils.Ipv4AddrValue(in.Ipv4Addr),
		Ipv6Address:   utils.Ipv6AddrValue(in.Ipv6Addr),
		// RoutingPolicies: handled by caller
	}

	if in.Ttl > 0 {
		result.Ttl = types.Int64Value(int64(in.Ttl))
	}

	return result
}

func BgpPeeringIpEndpointPrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]BgpPeeringIpEndpoint)

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting BgpPeeringIpEndpoint primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newBgpPeeringIpEndpoint(ctx, p, diags)
			newPrimitive.PipelineId = types.StringPointerValue((*string)(subpolicy.PipelineId))
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			newPrimitive.BatchId = types.StringPointerValue((*string)(subpolicy.BatchId))
			newPrimitive.RoutingPolicies = RoutingPolicyPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: BgpPeeringIpEndpoint{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: BgpPeeringIpEndpoint{}.AttrTypes()}, result, diags)
}

func LoadIDsIntoBgpPeeringIpEndpointMap(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, inMap types.Map, diags *diag.Diagnostics) types.Map {
	result := make(map[string]BgpPeeringIpEndpoint, len(inMap.Elements()))
	inMap.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: BgpPeeringIpEndpoint{}.AttrTypes()})
	}

	for _, p := range subpolicies {
		if _, ok := p.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachIpEndpointWithBgpNsxt); !ok {
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

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: BgpPeeringIpEndpoint{}.AttrTypes()}, result, diags)
}

var _ planmodifier.String = (*bgpPeeringIpEndpointBatchIdPlanModifier)(nil)

type bgpPeeringIpEndpointBatchIdPlanModifier struct{}

func (o bgpPeeringIpEndpointBatchIdPlanModifier) Description(_ context.Context) string {
	return "preserves the the state value unless all child primitives have been removed, in which case null is planned"
}

func (o bgpPeeringIpEndpointBatchIdPlanModifier) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o bgpPeeringIpEndpointBatchIdPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var plan, state BgpPeeringIpEndpoint

	// unpacking the parent object's plan should always work
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, req.Path.ParentPath(), &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// attempting to unpack the parent object's state indicates whether state *exists*
	d := req.State.GetAttribute(ctx, req.Path.ParentPath(), &state)
	stateDoesNotExist := d.HasError()

	// do we have zero children?
	if len(plan.RoutingPolicies.Elements()) == 0 {
		resp.PlanValue = types.StringNull() // with no children the batch id should be null
		return
	}

	// are we a new object?
	if stateDoesNotExist {
		resp.PlanValue = types.StringUnknown() // we are a new object. the batch id is not knowable
		return
	}

	// we're not new, and we have children. use the old value
	resp.PlanValue = req.StateValue
}
