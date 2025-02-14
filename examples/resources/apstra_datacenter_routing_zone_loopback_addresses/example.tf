# This example sets the IPv4 and IPv6 loopback addresses for two leaf switches
# in the prescribed blueprint and routing zone.
#
# Note that in this case, the map keys are references, so they must be
# wrapped in parentheses.
resource "apstra_datacenter_routing_zone_loopback_addresses" "example" {
  blueprint_id    = local.blueprint_id
  routing_zone_id = local.routing_zone_id
  loopbacks = {
    (local.leaf_1_id) = {
      ipv4_addr = "192.0.2.1/32"
      ipv6_addr = "3fff::1/128"
    }
    (local.leaf_2_id) = {
      ipv4_addr = "192.0.2.2/32"
      ipv6_addr = "3fff::2/128"
    }
  }
}
