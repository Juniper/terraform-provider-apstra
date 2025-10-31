package blueprint

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoutingZoneLoopbacks struct {
	BlueprintId   types.String `tfsdk:"blueprint_id"`
	RoutingZoneId types.String `tfsdk:"routing_zone_id"`
	Loopbacks     types.Map    `tfsdk:"loopbacks"`
	LoopbackIds   types.Map    `tfsdk:"loopback_ids"`
}

func (o RoutingZoneLoopbacks) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Routing Zone ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"loopbacks": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Loopback IPv4 and IPv6 addresses, keyed by System (switch) Node ID.",
			Required:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: RoutingZoneLoopback{}.ResourceAttributes(),
			},
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
		"loopback_ids": resourceSchema.MapAttribute{
			MarkdownDescription: "Map of Loopback interface Node IDs configured by this resource, keyed by System (switch) Node ID.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o *RoutingZoneLoopbacks) Request(ctx context.Context, bp *apstra.TwoStageL3ClosClient, previousLoopbackMap private.ResourceDatacenterRoutingZoneLoopbackAddresses, diags *diag.Diagnostics) (map[apstra.ObjectId]apstra.SecurityZoneLoopback, *private.ResourceDatacenterRoutingZoneLoopbackAddresses) {
	// API response will allow us to determine interface IDs from system IDs
	szInfo, err := bp.GetSecurityZoneInfo(ctx, apstra.ObjectId(o.RoutingZoneId.ValueString()))
	if err != nil {
		diags.AddError("failed querying for security zone", err.Error())
		return nil, nil
	}

	// convert API response to map (switchId -> loopbackId) for easy lookups
	loopbackIds := make(map[string]apstra.ObjectId)
	for _, memberInterface := range szInfo.MemberInterfaces {
		for _, loopback := range memberInterface.Loopbacks {
			// this is a loop, but only one element should exist. See: https://apstra-eng.slack.com/archives/CQBUYMZ39/p1738920079268339
			loopbackIds[memberInterface.HostingSystem.Id.String()] = loopback.Id
		}
	}
	o.LoopbackIds = utils.MapValueOrNull(ctx, types.StringType, loopbackIds, diags)

	// extract the planned loopbacks
	var planLoopbackMap map[string]RoutingZoneLoopback
	diags.Append(o.Loopbacks.ElementsAs(ctx, &planLoopbackMap, false)...)
	if diags.HasError() {
		return nil, nil
	}

	// we return these two maps
	resultMap := make(map[apstra.ObjectId]apstra.SecurityZoneLoopback, len(planLoopbackMap))
	resultPrivate := make(private.ResourceDatacenterRoutingZoneLoopbackAddresses, len(planLoopbackMap))

	for sysId, loopback := range planLoopbackMap {
		// ensure the specified system ID exists in the RZ-specific map we got from the API
		loopbackId, ok := loopbackIds[sysId]
		if !ok {
			diags.AddError(
				"System not participating in Routing Zone",
				fmt.Sprintf("System %s not participating in routing zone %s", sysId, o.RoutingZoneId),
			)
			return nil, nil
		}

		var szl apstra.SecurityZoneLoopback // this will be a resultMap entry
		var p struct {                      // this will be a resultPrivate entry
			HasIpv4 bool `json:"has_ipv4"`
			HasIpv6 bool `json:"has_ipv6"`
		}

		if !loopback.Ipv4Addr.IsNull() {
			szl.IPv4Addr = pointer.To(netip.MustParsePrefix(loopback.Ipv4Addr.ValueString()))
			p.HasIpv4 = true
		} else {
			if previousLoopbackMap[sysId].HasIpv4 {
				szl.IPv4Addr = new(netip.Prefix) // signals to remove previous value
			}
		}

		if !loopback.Ipv6Addr.IsNull() {
			szl.IPv6Addr = pointer.To(netip.MustParsePrefix(loopback.Ipv6Addr.ValueString()))
			p.HasIpv6 = true
		} else {
			if previousLoopbackMap[sysId].HasIpv6 {
				szl.IPv6Addr = new(netip.Prefix) // signals to remove previous value
			}
		}

		// previous addresses (if any) are no longer of any interest
		delete(previousLoopbackMap, sysId)

		resultPrivate[sysId] = p
		resultMap[loopbackId] = szl
	}

	// loop over remaining previous IP assignments; clear them as necessary
	for sysId, previous := range previousLoopbackMap {
		ifId, ok := loopbackIds[sysId]
		if !ok {
			continue // system no longer exists
		}

		var szl apstra.SecurityZoneLoopback
		if previous.HasIpv4 {
			szl.IPv4Addr = new(netip.Prefix) // bogus prefix clears entry from API
		}
		if previous.HasIpv6 {
			szl.IPv6Addr = new(netip.Prefix) // bogus prefix clears entry from API
		}
		resultMap[ifId] = szl
	}

	return resultMap, &resultPrivate
}

func (o *RoutingZoneLoopbacks) LoadApiData(ctx context.Context, info *apstra.TwoStageL3ClosSecurityZoneInfo, ps private.State, diags *diag.Diagnostics) {
	// extract private state (previously configured loopbacks)
	var previousLoopbackMap private.ResourceDatacenterRoutingZoneLoopbackAddresses
	if ps != nil {
		previousLoopbackMap.LoadPrivateState(ctx, ps, diags)
		if diags.HasError() {
			return
		}
	}

	loopbacks := make(map[string]RoutingZoneLoopback, len(previousLoopbackMap))
	for _, memberInterface := range info.MemberInterfaces {
		switch len(memberInterface.Loopbacks) {
		case 0: // weird, but whatever
			continue
		case 1: // expected case handled below
		default: // error condition
			diags.AddError(
				"invalid API response",
				fmt.Sprintf("System %q has %d loopback interfaces in Routing Zone %s, expected 0 or 1",
					memberInterface.HostingSystem.Id, len(memberInterface.Loopbacks), o.RoutingZoneId),
			)
			return
		}

		previous, ok := previousLoopbackMap[memberInterface.HostingSystem.Id.String()]
		if !ok {
			continue
		}

		var ipv4Addr cidrtypes.IPv4Prefix
		if memberInterface.Loopbacks[0].Ipv4Addr != nil && previous.HasIpv4 {
			ipv4Addr = cidrtypes.NewIPv4PrefixValue(memberInterface.Loopbacks[0].Ipv4Addr.String())
		}

		var ipv6Addr cidrtypes.IPv6Prefix
		if memberInterface.Loopbacks[0].Ipv6Addr != nil && previous.HasIpv6 {
			ipv6Addr = cidrtypes.NewIPv6PrefixValue(memberInterface.Loopbacks[0].Ipv6Addr.String())
		}

		loopbacks[memberInterface.HostingSystem.Id.String()] = RoutingZoneLoopback{
			Ipv4Addr: ipv4Addr,
			Ipv6Addr: ipv6Addr,
		}
	}

	o.Loopbacks = utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: RoutingZoneLoopback{}.AttrTypes()}, loopbacks, diags)
}
