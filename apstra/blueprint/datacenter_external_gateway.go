package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math"
	"net"
	"strings"
)

type DatacenterExternalGateway struct {
	Id                types.String        `tfsdk:"id"`
	BlueprintId       types.String        `tfsdk:"blueprint_id"`
	Name              types.String        `tfsdk:"name"`
	IpAddress         iptypes.IPv4Address `tfsdk:"ip_address"`
	Asn               types.Int64         `tfsdk:"asn"`
	Ttl               types.Int64         `tfsdk:"ttl"`
	KeepaliveTime     types.Int64         `tfsdk:"keepalive_time"`
	HoldTime          types.Int64         `tfsdk:"hold_time"`
	EvpnRouteTypes    types.String        `tfsdk:"evpn_route_types"`
	LocalGatewayNodes types.Set           `tfsdk:"local_gateway_nodes"`
}

func (o DatacenterExternalGateway) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Blueprint in which the External Gateway should be created.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "External Gateway name",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ip_address": resourceSchema.StringAttribute{
			MarkdownDescription: "External Gateway IP address",
			Required:            true,
			CustomType:          iptypes.IPv4AddressType{},
		},
		"asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "External Gateway AS Number",
			Required:            true,
			Validators:          []validator.Int64{int64validator.Between(1, int64(math.MaxUint32))},
		},
		"ttl": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP Time To Live. Omit to use device defaults.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.Int64{int64validator.Between(2, int64(math.MaxUint8))},
		},
		"keepalive_time": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP keepalive time (seconds).",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.Int64{int64validator.Between(1, int64(math.MaxUint16))},
		},
		"hold_time": resourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.Int64{int64validator.Between(3, int64(math.MaxUint16))},
		},
		"evpn_route_types": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(`EVPN route types. Valid values are: ["%s"]. Default: %q`,
				strings.Join(apstra.RemoteGatewayRouteTypesEnum.Values(), `", "`),
				apstra.RemoteGatewayRouteTypesAll.Value),
			Optional:   true,
			Computed:   true,
			Default:    stringdefault.StaticString(apstra.RemoteGatewayRouteTypesAll.Value),
			Validators: []validator.String{stringvalidator.OneOf(apstra.RemoteGatewayRouteTypesEnum.Values()...)},
		},
		"local_gateway_nodes": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of IDs of switch nodes which will be configured to peer with the External Gateway",
			Required:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *DatacenterExternalGateway) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.RemoteGatewayData {
	routeTypes := apstra.RemoteGatewayRouteTypesEnum.Parse(o.EvpnRouteTypes.ValueString())
	// skipping nil check because input validation should make that impossible

	var localGwNodes []apstra.ObjectId
	diags.Append(o.LocalGatewayNodes.ElementsAs(ctx, &localGwNodes, false)...)
	if diags.HasError() {
		return nil
	}

	var ttl *uint8
	if utils.Known(o.Ttl) {
		t := uint8(o.Ttl.ValueInt64())
		ttl = &t
	}

	var keepaliveTimer *uint16
	if utils.Known(o.KeepaliveTime) {
		t := uint16(o.KeepaliveTime.ValueInt64())
		keepaliveTimer = &t
	}

	var holdtimeTimer *uint16
	if utils.Known(o.HoldTime) {
		t := uint16(o.HoldTime.ValueInt64())
		holdtimeTimer = &t
	}

	return &apstra.RemoteGatewayData{
		RouteTypes:     *routeTypes,
		LocalGwNodes:   localGwNodes,
		GwAsn:          uint32(o.Asn.ValueInt64()),
		GwIp:           net.ParseIP(o.IpAddress.ValueString()), // skipping nil check because input
		GwName:         o.Name.ValueString(),                   // validation should make that impossible
		Ttl:            ttl,
		KeepaliveTimer: keepaliveTimer,
		HoldtimeTimer:  holdtimeTimer,
	}
}

func (o *DatacenterExternalGateway) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	remoteGateway, err := bp.GetRemoteGateway(ctx, apstra.ObjectId(o.Id.ValueString()))
	if err != nil {
		diags.AddError("failed to fetch remote gateway", err.Error())
		return
	}

	o.loadApiData(ctx, remoteGateway.Data, diags)
	if diags.HasError() {
		return
	}
}

func (o *DatacenterExternalGateway) loadApiData(_ context.Context, in *apstra.RemoteGatewayData, _ *diag.Diagnostics) {
	ttl := types.Int64Null()
	if in.Ttl != nil {
		ttl = types.Int64Value(int64(*in.Ttl))
	}

	keepaliveTime := types.Int64Null()
	if in.KeepaliveTimer != nil {
		keepaliveTime = types.Int64Value(int64(*in.KeepaliveTimer))
	}

	holdTime := types.Int64Null()
	if in.HoldtimeTimer != nil {
		holdTime = types.Int64Value(int64(*in.HoldtimeTimer))
	}

	localGatewayNodes := make([]attr.Value, len(in.LocalGwNodes))
	for i, localGatewayNode := range in.LocalGwNodes {
		localGatewayNodes[i] = types.StringValue(localGatewayNode.String())
	}

	o.Name = types.StringValue(in.GwName)
	o.IpAddress = iptypes.NewIPv4AddressValue(in.GwIp.String())
	o.Asn = types.Int64Value(int64(in.GwAsn))
	o.Ttl = ttl
	o.KeepaliveTime = keepaliveTime
	o.HoldTime = holdTime
	o.EvpnRouteTypes = types.StringValue(in.RouteTypes.Value)
	o.LocalGatewayNodes = types.SetValueMust(types.StringType, localGatewayNodes)
}
