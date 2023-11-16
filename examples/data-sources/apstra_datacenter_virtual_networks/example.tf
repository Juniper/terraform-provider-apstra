# Without specifying no filter, a wide search is performed. The ID of all
# routing zones in the blueprint will be returned in the `ids` attribute..
data "apstra_datacenter_virtual_networks" "all" {
  blueprint_id = "05f9d3fc-671a-4efc-8e91-5ef87b2937d3"
}

# This example performs a narrow search. It will construct and
# and execute the following graph query:
# match(
#   node(type='virtual_network', name='n_virtual_network', reserved_vlan_id=is_none()),
#   node(type='virtual_network', name='n_virtual_network')
#     .in_(type='member_vns')
#     .node(type='security_zone', id='Zplm0niOFCCCfjaXkXo'),
#   node(type='virtual_network', name='n_virtual_network')
#     .out(type='instantiated_by')
#     .node(type='vn_instance', dhcp_enabled=True)
# )
data "apstra_datacenter_virtual_networks" "prod_unreserved_with_dhcp" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  filters = [
    {
      reserve_vlan         = false
      dhcp_service_enabled = true
      routing_zone_id      = "Zplm0niOFCCCfjaXkXo"
    }
  ]
}
