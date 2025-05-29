package blueprint

import (
	"context"
	"fmt"
	"math"
	"net/netip"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExternalGateway struct {
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
	Password          types.String        `tfsdk:"password"`
}

func (o ExternalGateway) ResourceAttributes() map[string]resourceSchema.Attribute {
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
				strings.Join(enum.RemoteGatewayRouteTypes.Values(), `", "`),
				enum.RemoteGatewayRouteTypeAll.Value),
			Optional:   true,
			Computed:   true,
			Default:    stringdefault.StaticString(enum.RemoteGatewayRouteTypeAll.Value),
			Validators: []validator.String{stringvalidator.OneOf(enum.RemoteGatewayRouteTypes.Values()...)},
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
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "BGP TCP authentication password",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			Sensitive:           true,
		},
	}
}

func (o ExternalGateway) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Blueprint in which the External Gateway should be created.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "External Gateway name",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ip_address": dataSourceSchema.StringAttribute{
			MarkdownDescription: "External Gateway IP address",
			Computed:            true,
			CustomType:          iptypes.IPv4AddressType{},
		},
		"asn": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "External Gateway AS Number",
			Computed:            true,
		},
		"ttl": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP Time To Live. Omit to use device defaults.",
			Computed:            true,
		},
		"keepalive_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP keepalive time (seconds).",
			Computed:            true,
		},
		"hold_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Computed:            true,
		},
		"evpn_route_types": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(`EVPN route types. Valid values are: ["%s"]. Default: %q`,
				strings.Join(enum.RemoteGatewayRouteTypes.Values(), `", "`),
				enum.RemoteGatewayRouteTypeAll.Value),
			Computed: true,
		},
		"local_gateway_nodes": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of IDs of switch nodes which will be configured to peer with the External Gateway",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"password": dataSourceSchema.StringAttribute{
			MarkdownDescription: "BGP TCP authentication password",
			Computed:            true,
			Sensitive:           true,
		},
	}
}

func (o ExternalGateway) DataSourceAttributesAsFilter() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "External Gateway name",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ip_address": dataSourceSchema.StringAttribute{
			MarkdownDescription: "External Gateway IP address",
			Optional:            true,
			CustomType:          iptypes.IPv4AddressType{},
		},
		"asn": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "External Gateway AS Number",
			Optional:            true,
		},
		"ttl": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP Time To Live. Omit to use device defaults.",
			Optional:            true,
		},
		"keepalive_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP keepalive time (seconds).",
			Optional:            true,
		},
		"hold_time": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "BGP hold time (seconds).",
			Optional:            true,
		},
		"evpn_route_types": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(`EVPN route types. Valid values are: ["%s"]. Default: %q`,
				strings.Join(enum.RemoteGatewayRouteTypes.Values(), `", "`),
				enum.RemoteGatewayRouteTypeAll.Value),
			Optional: true,
		},
		"local_gateway_nodes": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of IDs of switch nodes which will be configured to peer with the External Gateway",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"password": dataSourceSchema.StringAttribute{
			MarkdownDescription: "BGP TCP authentication password",
			Optional:            true,
			Sensitive:           true,
		},
	}
}

func (o *ExternalGateway) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.TwoStageL3ClosRemoteGatewayData {
	routeTypes := enum.RemoteGatewayRouteTypes.Parse(o.EvpnRouteTypes.ValueString())
	// skipping nil check because input validation should make that impossible

	var localGwNodes []apstra.ObjectId
	diags.Append(o.LocalGatewayNodes.ElementsAs(ctx, &localGwNodes, false)...)
	if diags.HasError() {
		return nil
	}

	var ttl *uint8
	if utils.HasValue(o.Ttl) {
		t := uint8(o.Ttl.ValueInt64())
		ttl = &t
	}

	var keepaliveTimer *uint16
	if utils.HasValue(o.KeepaliveTime) {
		t := uint16(o.KeepaliveTime.ValueInt64())
		keepaliveTimer = &t
	}

	var holdtimeTimer *uint16
	if utils.HasValue(o.HoldTime) {
		t := uint16(o.HoldTime.ValueInt64())
		holdtimeTimer = &t
	}

	var password *string
	if utils.HasValue(o.Password) {
		t := o.Password.ValueString()
		password = &t
	}

	gwIp, _ := netip.ParseAddr(o.IpAddress.ValueString()) // ignoring error; address already validated

	return &apstra.TwoStageL3ClosRemoteGatewayData{
		RouteTypes:     routeTypes,
		LocalGwNodes:   localGwNodes,
		GwAsn:          uint32(o.Asn.ValueInt64()),
		GwIp:           gwIp,
		Label:          o.Name.ValueString(),
		Ttl:            ttl,
		KeepaliveTimer: keepaliveTimer,
		HoldtimeTimer:  holdtimeTimer,
		Password:       password,
	}
}

func (o *ExternalGateway) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) error {
	var err error
	var api *apstra.TwoStageL3ClosRemoteGateway

	if o.Id.IsNull() {
		api, err = bp.GetRemoteGatewayByName(ctx, o.Name.ValueString())
		if err != nil {
			return err
		}
		if api.Data.EvpnInterconnectGroupId != nil {
			diags.AddError("object has wrong type", fmt.Sprintf("remote gateway %q is an Interconnect Domain Gateway, not an External Gateway", o.Name.ValueString()))
			return nil
		}
		o.Id = types.StringValue(api.Id.String())
	} else {
		api, err = bp.GetRemoteGateway(ctx, apstra.ObjectId(o.Id.ValueString()))
		if err != nil {
			return err
		}
		if api.Data.EvpnInterconnectGroupId != nil {
			diags.AddError("object has wrong type", fmt.Sprintf("remote gateway %q is an Interconnect Domain Gateway, not an External Gateway", o.Id.ValueString()))
			return nil
		}
	}

	o.LoadApiData(ctx, api.Data, diags)
	if diags.HasError() {
		return nil
	}

	o.ReadProtocolPassword(ctx, bp, diags)
	if diags.HasError() {
		return nil
	}

	return nil
}

func (o *ExternalGateway) ReadProtocolPassword(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	query := new(apstra.PathQuery).
		SetClient(bp.Client()).
		SetBlueprintId(bp.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeProtocol.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeProtocol.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_protocol")},
		})

	var queryResponse struct {
		Items []struct {
			Protocol struct {
				Password *string `json:"password"`
			} `json:"n_protocol"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResponse)
	if err != nil {
		diags.AddError("failed while performing graph query",
			fmt.Sprintf("error: %q\nquery: %q\n", err.Error(), query.String()))
		return
	}

	// count usage of each discovered password (there should only be one password, used everywhere)
	pwUsageCounts := make(map[string]int)
	var password string
	for _, item := range queryResponse.Items {
		if item.Protocol.Password == nil {
			continue
		}

		password = *item.Protocol.Password // save the (only?) password outside the map
		pwUsageCounts[password]++          // increment the password use counter
	}

	// how many passwords discovered?
	switch len(pwUsageCounts) {
	case 0:
		o.Password = types.StringNull() // no passwords found - this is fine!
		return
	case 1: // expected case (only one password found) handled below
	default:
		diags.AddError("multiple protocol passwords found",
			fmt.Sprintf("remote gateway node %s sessions use mismatched passwords", o.Id))
		return
	}

	// if we got here, only one password is in use. That's good, but is it in use on *every* protocol session?
	if len(queryResponse.Items) > pwUsageCounts[password] {
		diags.AddError("protocol password not used uniformly",
			fmt.Sprintf("remote gateway node %s has %d protocol sessions, but only %d of them use a password",
				o.Id, len(queryResponse.Items), pwUsageCounts[password]))
		return
	}

	if len(queryResponse.Items) < pwUsageCounts[password] {
		diags.AddWarning(constants.ErrProviderBug,
			"graph query found more protocol session passwords than sessions - this should be impossible")
		return
	}

	o.Password = types.StringValue(password)
}

func (o *ExternalGateway) LoadApiData(_ context.Context, in *apstra.TwoStageL3ClosRemoteGatewayData, _ *diag.Diagnostics) {
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

	o.Name = types.StringValue(in.Label)
	o.IpAddress = iptypes.NewIPv4AddressValue(in.GwIp.String())
	o.Asn = types.Int64Value(int64(in.GwAsn))
	o.Ttl = ttl
	o.KeepaliveTime = keepaliveTime
	o.HoldTime = holdTime
	o.EvpnRouteTypes = types.StringValue(in.RouteTypes.Value)
	o.LocalGatewayNodes = types.SetValueMust(types.StringType, localGatewayNodes)
}

func (o ExternalGateway) FilterMatch(_ context.Context, in *ExternalGateway, _ *diag.Diagnostics) bool {
	if !o.Id.IsNull() && !o.Id.Equal(in.Id) {
		return false
	}

	if !o.Name.IsNull() && !o.Name.Equal(in.Name) {
		return false
	}

	if !o.IpAddress.IsNull() && !o.IpAddress.Equal(in.IpAddress) {
		return false
	}

	if !o.Asn.IsNull() && !o.Asn.Equal(in.Asn) {
		return false
	}

	if !o.Ttl.IsNull() && !o.Ttl.Equal(in.Ttl) {
		return false
	}

	if !o.KeepaliveTime.IsNull() && !o.KeepaliveTime.Equal(in.KeepaliveTime) {
		return false
	}

	if !o.HoldTime.IsNull() && !o.HoldTime.Equal(in.HoldTime) {
		return false
	}

	if !o.EvpnRouteTypes.IsNull() && !o.EvpnRouteTypes.Equal(in.EvpnRouteTypes) {
		return false
	}

	if !o.Password.IsNull() && !o.Password.Equal(in.Password) {
		return false
	}

	if !o.LocalGatewayNodes.IsNull() {
		// extract the candidate localGatewayNodes as a map for quick lookups
		actualLGWs := make(map[string]bool, len(in.LocalGatewayNodes.Elements()))
		for _, item := range in.LocalGatewayNodes.Elements() {
			actualLGWs[item.(types.String).ValueString()] = true
		}

		// fail if any required item is missing from candidate items
		for _, requiredLGW := range o.LocalGatewayNodes.Elements() {
			if !actualLGWs[requiredLGW.(types.String).ValueString()] {
				return false
			}
		}
	}

	return true
}
