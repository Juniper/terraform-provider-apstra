# The following example creates a VXLAN Virtual Network and makes it
# available on 3 switches in rack "a" and one switch in rack "b".
#
# Note that computing the 'bindings' data requires knowledge of the
# topological relationship between leaf and access switches, ESI/MLAG
# redundancy group relationships and graph db node IDs of switch and
# group nodes.
#
# Users are encouraged to use the `apstra_datacenter_systems`
# and `apstra_datacenter_virtual_network_binding_constructor` data sources to
# produce the 'bindings' data.

resource "apstra_datacenter_virtual_network" "test" {
  name                         = "test"
  blueprint_id                 = "blueprint-id"
  type                         = "vxlan"
  routing_zone_id              = "routing-zone-id"
  ipv4_connectivity_enabled    = true
  ipv4_virtual_gateway_enabled = true
  ipv4_virtual_gateway         = "192.168.10.1"
  ipv4_subnet                  = "192.168.10.0/24"
  bindings = {
    "leaf-a" = {
      "access_ids" = [ "access-a-1", "access-a-2" ]
    },
    "leaf-b" = {
      "access_ids" = []
    }
  }
}
