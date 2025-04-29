// This is a new implementation of the LoadApiData method to replace the existing one
// Copy the content of this function into datacenter_virtual_network.go

func (o *DatacenterVirtualNetwork) LoadApiData(ctx context.Context, in *apstra.VirtualNetworkData, diags *diag.Diagnostics) {
	var virtualGatewayIpv4, virtualGatewayIpv6 string
	if len(in.VirtualGatewayIpv4.To4()) == net.IPv4len {
		virtualGatewayIpv4 = in.VirtualGatewayIpv4.String()
	}
	if len(in.VirtualGatewayIpv6) == net.IPv6len {
		virtualGatewayIpv6 = in.VirtualGatewayIpv6.String()
	}

	o.Name = types.StringValue(in.Label)
	o.Description = utils.StringValueOrNull(ctx, in.Description, diags)
	o.Type = types.StringValue(in.VnType.String())
	o.RoutingZoneId = types.StringValue(in.SecurityZoneId.String())
	o.Bindings = newBindingMap(ctx, in.VnBindings, diags)
	o.Vni = utils.Int64ValueOrNull(ctx, in.VnId, diags)
	o.DhcpServiceEnabled = types.BoolValue(bool(in.DhcpService))
	o.IPv4ConnectivityEnabled = types.BoolValue(in.Ipv4Enabled)
	o.IPv6ConnectivityEnabled = types.BoolValue(in.Ipv6Enabled)
	o.ReserveVlan = types.BoolValue(in.ReservedVlanId != nil)
	if in.ReservedVlanId == nil {
		o.ReservedVlanId = types.Int64Null()
	} else {
		o.ReservedVlanId = types.Int64Value(int64(*in.ReservedVlanId))
	}
	if in.Ipv4Subnet == nil {
		o.IPv4Subnet = types.StringNull()
	} else {
		o.IPv4Subnet = types.StringValue(in.Ipv4Subnet.String())
	}
	if in.Ipv6Subnet == nil {
		o.IPv6Subnet = types.StringNull()
	} else {
		o.IPv6Subnet = types.StringValue(in.Ipv6Subnet.String())
	}
	o.IPv4GatewayEnabled = types.BoolValue(in.VirtualGatewayIpv4Enabled)
	o.IPv6GatewayEnabled = types.BoolValue(in.VirtualGatewayIpv6Enabled)
	o.IPv4Gateway = utils.StringValueOrNull(ctx, virtualGatewayIpv4, diags)
	o.IPv6Gateway = utils.StringValueOrNull(ctx, virtualGatewayIpv6, diags)
	o.L3Mtu = utils.Int64ValueOrNull(ctx, in.L3Mtu, diags)
	
	// Load SVI IPs
	o.SviIps = LoadApiSviIps(ctx, in.SviIps, diags)

	if in.RtPolicy == nil {
		o.ImportRouteTargets = types.SetNull(types.StringType)
		o.ExportRouteTargets = types.SetNull(types.StringType)
	} else {
		o.ImportRouteTargets = utils.SetValueOrNull(ctx, types.StringType, in.RtPolicy.ImportRTs, diags)
		o.ExportRouteTargets = utils.SetValueOrNull(ctx, types.StringType, in.RtPolicy.ExportRTs, diags)
	}
}