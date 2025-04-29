// This is a new implementation of ResourceAttributes to replace the existing one
// Copy the content of this function into datacenter_virtual_network.go

func (o DatacenterVirtualNetwork) ResourceAttributes() map[string]resourceSchema.Attribute {
	attrs := map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Name",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 30),
				stringvalidator.RegexMatches(apstraregexp.AlphaNumW2HLConstraint, apstraregexp.AlphaNumW2HLConstraintMsg),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Description",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 222),
				stringvalidator.RegexMatches(regexp.MustCompile(`^[^"<>\\?]+$`), `must not contain the following characters: ", <, >, \, ?`),
			},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Type",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(enum.VnTypeVxlan.String()),
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{
				// specifically enumerated types - SDK supports additional
				// types which do not make sense in this context.
				stringvalidator.OneOf(enum.VnTypeVlan.String(), enum.VnTypeVxlan.String()),
			},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Routing Zone ID (required when `type == %s`", enum.VnTypeVxlan),
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.RequiredWhenValueIs(
					path.MatchRelative().AtParent().AtName("type"),
					types.StringValue(enum.VnTypeVxlan.String()),
				),
				apstravalidator.RequiredWhenValueNull(
					path.MatchRelative().AtParent().AtName("type"),
				),
			},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"vni": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("EVPN Virtual Network ID to be associated with this Virtual "+
				"Network.  When omitted, Apstra chooses a VNI from the Resource Pool [allocated]"+
				"(../resources/datacenter_resource_pool_allocation) to role `%s`.",
				utils.StringersToFriendlyString(apstra.ResourceGroupNameVxlanVnIds)),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.VniMin, constants.VniMax),
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRelative().AtParent().AtName("type"),
					types.StringValue(enum.VnTypeVlan.String()),
				),
			},
		},
		"had_prior_vni_config": resourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vni` has been removed from the configuration.",
			Computed:            true,
		},
		"reserve_vlan": resourceSchema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("For use only with `%s` type Virtual networks when all "+
				"`bindings` use the same VLAN ID. This option reserves the VLAN fabric-wide, even on "+
				"switches to which the Virtual Network has not yet been deployed.", enum.VnTypeVxlan.String()),
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(
					types.BoolValue(true),
					apstravalidator.ForbiddenWhenValueIs(
						path.MatchRelative().AtParent().AtName("type"),
						types.StringValue(enum.VnTypeVlan.String()),
					),
				),
				apstravalidator.AlsoRequiresNOf(1,
					path.MatchRoot("bindings"),
					path.MatchRoot("reserved_vlan_id"),
				),
			},
		},
		"reserved_vlan_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "Used to specify the reserved VLAN ID without specifying any *bindings*.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("reserve_vlan"), types.BoolNull()),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("reserve_vlan"), types.BoolValue(false)),
				int64validator.ConflictsWith(path.MatchRoot("bindings")),
				int64validator.Between(design.VlanMin, design.VlanMax),
			},
		},
		"bindings": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Bindings make a Virtual Network available on Leaf Switches and Access Switches. " +
				"At least one binding entry is required with Apstra 4.x. With Apstra 5.x, a Virtual Network with " +
				"no bindings can be created by omitting (or setting `null`) this attribute. The value is a map " +
				"keyed by graph db node IDs of *either* Leaf Switches (non-redundant Leaf Switches) or Leaf Switch " +
				"redundancy groups (redundant Leaf Switches). Practitioners are encouraged to consider using the " +
				"[`apstra_datacenter_virtual_network_binding_constructor`]" +
				"(../data-sources/datacenter_virtual_network_binding_constructor) data source to populate " +
				"this map.",
			Optional: true,
			Validators: []validator.Map{
				mapvalidator.SizeAtLeast(1),
				apstravalidator.WhenValueAtMustBeMap(
					path.MatchRelative().AtParent().AtName("type"),
					types.StringValue(enum.VnTypeVlan.String()),
					mapvalidator.SizeAtMost(1),
				),
			},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: VnBinding{}.ResourceAttributes(),
			},
		},
		"dhcp_service_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables a DHCP relay agent.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(types.BoolValue(true),
					apstravalidator.AlsoRequiresNOf(1,
						path.MatchRelative().AtParent().AtName("ipv4_connectivity_enabled"),
						path.MatchRelative().AtParent().AtName("ipv6_connectivity_enabled"),
					),
				),
			},
		},
		"ipv4_connectivity_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv4 within the Virtual Network. Default: true",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
		},
		"ipv6_connectivity_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables IPv6 within the Virtual Network. Default: false",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"ipv4_subnet": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("IPv4 subnet associated with the "+
				"Virtual Network. When not specified, a prefix from within the IPv4 "+
				"Resource Pool assigned to the `%s` role will be automatically a"+
				"ssigned by Apstra.", apstra.ResourceGroupNameVirtualNetworkSviIpv4),
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				apstravalidator.ParseCidr(true, false),
				apstravalidator.WhenValueSetString(
					apstravalidator.ValueAtMustBeString(
						path.MatchRelative().AtParent().AtName("ipv4_connectivity_enabled"),
						types.BoolValue(true), false,
					),
				),
			},
		},
		"ipv6_subnet": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("IPv6 subnet associated with the "+
				"Virtual Network. When not specified, a prefix from within the IPv6 "+
				"Resource Pool assigned to the `%s` role will be automatically a"+
				"ssigned by Apstra.", apstra.ResourceGroupNameVirtualNetworkSviIpv6),
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				apstravalidator.ParseCidr(false, true),
				apstravalidator.WhenValueSetString(
					apstravalidator.ValueAtMustBeString(
						path.MatchRelative().AtParent().AtName("ipv6_connectivity_enabled"),
						types.BoolValue(true), false,
					),
				),
			},
		},
		"ipv4_virtual_gateway_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Controls and indicates whether the IPv4 gateway within the " +
				"Virtual Network is enabled. Requires `ipv4_connectivity_enabled` to be `true`",
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(
					types.BoolValue(true),
					apstravalidator.ValueAtMustBeBool(
						path.MatchRelative().AtParent().AtName("ipv4_connectivity_enabled"),
						types.BoolValue(true),
						false,
					),
				),
			},
		},
		"ipv6_virtual_gateway_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Controls and indicates whether the IPv6 gateway within the " +
				"Virtual Network is enabled. Requires `ipv6_connectivity_enabled` to be `true`",
			Optional: true,
			Computed: true,
			Validators: []validator.Bool{
				apstravalidator.WhenValueIsBool(
					types.BoolValue(true),
					apstravalidator.ValueAtMustBeBool(
						path.MatchRelative().AtParent().AtName("ipv6_connectivity_enabled"),
						types.BoolValue(true),
						false,
					),
				),
			},
		},
		"ipv4_virtual_gateway": resourceSchema.StringAttribute{
			MarkdownDescription: "Specifies the IPv4 virtual gateway address within the " +
				"Virtual Network. The configured value must be a valid IPv4 host address " +
				"configured value within range specified by `ipv4_subnet`",
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				apstravalidator.ParseIp(true, false),
				apstravalidator.FallsWithinCidr(
					path.MatchRelative().AtParent().AtName("ipv4_subnet"),
					false, false),
			},
		},
		"ipv6_virtual_gateway": resourceSchema.StringAttribute{
			MarkdownDescription: "Specifies the IPv6 virtual gateway address within the " +
				"Virtual Network. The configured value must be a valid IPv6 host address " +
				"configured value within range specified by `ipv6_subnet`",
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				apstravalidator.ParseIp(false, true),
				apstravalidator.FallsWithinCidr(
					path.MatchRelative().AtParent().AtName("ipv6_subnet"),
					true, true),
			},
		},
		"l3_mtu": resourceSchema.Int64Attribute{
			MarkdownDescription: fmt.Sprintf("L3 MTU used by the L3 switch interfaces participating in the"+
				" Virtual Network. Must be an even number between %d and %d. Requires Apstra %s or later.",
				constants.L3MtuMin, constants.L3MtuMax, apiversions.Apstra420),
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.Between(constants.L3MtuMin, constants.L3MtuMax),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"import_route_targets": resourceSchema.SetAttribute{
			MarkdownDescription: "Import RTs for this Virtual Network.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseRT()),
			},
		},
		"export_route_targets": resourceSchema.SetAttribute{
			MarkdownDescription: "Export RTs for this Virtual Network.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseRT()),
			},
		},
		"svi_ips": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "SVI IP assignments for switches in the virtual network. This allows explicit " +
				"control over the secondary virtual interface IPs assigned to switches, preventing overlaps " +
				"when identical virtual networks are created in multiple blueprints.",
			Optional: true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: SviIp{}.ResourceAttributes(),
			},
		},
	}
	return attrs
}