// This is a new implementation of the Request method to replace the existing one
// Copy the content of this function into datacenter_virtual_network.go

func (o *DatacenterVirtualNetwork) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.VirtualNetworkData {
	var vnType enum.VnType
	err := vnType.FromString(o.Type.ValueString())
	if err != nil {
		diags.Append(
			validatordiag.BugInProviderDiagnostic(
				fmt.Sprintf("error parsing virtual network type %q - %s", o.Type.String(), err.Error())))
		return nil
	}

	b := make(map[string]VnBinding)
	diags.Append(o.Bindings.ElementsAs(ctx, &b, false)...)
	if diags.HasError() {
		return nil
	}
	vnBindings := make([]apstra.VnBinding, len(b))
	var i int
	for leafId, binding := range b {
		vnBindings[i] = *binding.Request(ctx, leafId, diags)
		i++
	}
	if diags.HasError() {
		return nil
	}

	var vnId *apstra.VNI
	if utils.HasValue(o.Vni) {
		v := apstra.VNI(o.Vni.ValueInt64())
		vnId = &v
	}

	if o.Type.ValueString() == enum.VnTypeVlan.String() {
		// Maximum of one binding is required when type==vlan.
		// Apstra requires vlan == vni when creating a "vlan" type VN.
		// VNI attribute is forbidden when type == VLAN
		if len(vnBindings) > 0 && vnBindings[0].VlanId != nil {
			v := apstra.VNI(*vnBindings[0].VlanId)
			vnId = &v
		}
	}

	var reservedVlanId *apstra.Vlan
	if o.ReserveVlan.ValueBool() {
		if utils.HasValue(o.ReservedVlanId) {
			reservedVlanId = utils.ToPtr(apstra.Vlan(o.ReservedVlanId.ValueInt64()))
		} else {
			reservedVlanId = vnBindings[0].VlanId
		}
	}

	var ipv4Subnet, ipv6Subnet *net.IPNet
	if utils.HasValue(o.IPv4Subnet) {
		_, ipv4Subnet, err = net.ParseCIDR(o.IPv4Subnet.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing attribute ipv4_subnet value %q", o.IPv4Subnet.ValueString()), err.Error())
		}
	}
	if utils.HasValue(o.IPv6Subnet) {
		_, ipv6Subnet, err = net.ParseCIDR(o.IPv6Subnet.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing attribute ipv6_subnet value %q", o.IPv6Subnet.ValueString()), err.Error())
		}
	}

	var ipv4Gateway, ipv6Gateway net.IP
	if utils.HasValue(o.IPv4Gateway) {
		ipv4Gateway = net.ParseIP(o.IPv4Gateway.ValueString())
	}
	if utils.HasValue(o.IPv6Gateway) {
		ipv6Gateway = net.ParseIP(o.IPv6Gateway.ValueString())
	}

	var l3Mtu *int
	if utils.HasValue(o.L3Mtu) {
		i := int(o.L3Mtu.ValueInt64())
		l3Mtu = &i
	}

	var rtPolicy *apstra.RtPolicy
	if !o.ImportRouteTargets.IsNull() || !o.ExportRouteTargets.IsNull() {
		rtPolicy = new(apstra.RtPolicy)
		if !o.ImportRouteTargets.IsNull() {
			diags.Append(o.ImportRouteTargets.ElementsAs(ctx, &rtPolicy.ImportRTs, false)...)
		}
		if !o.ExportRouteTargets.IsNull() {
			diags.Append(o.ExportRouteTargets.ElementsAs(ctx, &rtPolicy.ExportRTs, false)...)
		}
	}

	// Create the request object
	result := &apstra.VirtualNetworkData{
		Description:               o.Description.ValueString(),
		DhcpService:               apstra.DhcpServiceEnabled(o.DhcpServiceEnabled.ValueBool()),
		Ipv4Enabled:               o.IPv4ConnectivityEnabled.ValueBool(),
		Ipv4Subnet:                ipv4Subnet,
		Ipv6Enabled:               o.IPv6ConnectivityEnabled.ValueBool(),
		Ipv6Subnet:                ipv6Subnet,
		L3Mtu:                     l3Mtu,
		Label:                     o.Name.ValueString(),
		ReservedVlanId:            reservedVlanId,
		RouteTarget:               "",
		RtPolicy:                  rtPolicy,
		SecurityZoneId:            apstra.ObjectId(o.RoutingZoneId.ValueString()),
		SviIps:                    nil,
		VirtualGatewayIpv4:        ipv4Gateway,
		VirtualGatewayIpv6:        ipv6Gateway,
		VirtualGatewayIpv4Enabled: o.IPv4GatewayEnabled.ValueBool(),
		VirtualGatewayIpv6Enabled: o.IPv6GatewayEnabled.ValueBool(),
		VnBindings:                vnBindings,
		VnId:                      vnId,
		VnType:                    vnType,
		VirtualMac:                nil,
	}
	
	// Add SviIps to the request if provided
	if !o.SviIps.IsNull() {
		var sviIpsSlice []SviIp
		diags.Append(o.SviIps.ElementsAs(ctx, &sviIpsSlice, false)...)
		if diags.HasError() {
			return nil
		}

		apiSviIps := make([]apstra.SviIp, len(sviIpsSlice))
		for i, sviIp := range sviIpsSlice {
			apiSviIps[i] = *sviIp.Request(ctx, diags)
			if diags.HasError() {
				return nil
			}
		}
		
		result.SviIps = apiSviIps
	}
	
	return result
}