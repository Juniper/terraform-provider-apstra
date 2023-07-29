# Each apstra_datacenter_ct_* data source represents a Connectivity Template
# Primitive. They're stand-ins for the Primitives found in the Web UI's CT
# builder interface.
#
# These data sources do not interact with the Apstra API. Instead, they assemble
# their input fields into a JSON string presented at the `primitive` attribute
# key.
#
# Use the `primitive` output anywhere you need a primitive represented as JSON:
# - at the root of a Connectivity Template
# - as a child of another Primitive (as constrained by the accepts/produces
#   relationship between Primitives)

# Look up the details of the desired routing policy
data "apstra_datacenter_routing_policy" "default" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  name         = "Default_immutable"
}

# Declare a "Routing Policy" Connectivity Template Primitive:
data "apstra_datacenter_ct_routing_policy" "default" {
  routing_policy_id = data.apstra_datacenter_ct_routing_policy.default.id
}

# This data source's `primitive` attribute produces JSON like this:
# {
#   "type": "AttachExistingRoutingPolicy",
#   "data": {
#     "routing_policy_id": "Xd5Uoo8qUjCqhihGafQ"
#   }
# }

# Declare a "BGP Peering (Generic System)" Connectivity Template Primitive
# which uses the "Routing Policy" primitive:
data "apstra_datacenter_ct_bgp_peering_generic_system" "bgp_server" {
  ipv4_afi_enabled     = true
  ipv6_afi_enabled     = true
  ipv4_addressing_type = "addressed"
  ipv6_addressing_type = "link_local"
  bfd_enabled          = true
  ttl                  = 1
  password             = "big secret"
  children = [
    data.apstra_datacenter_routing_policy.default.primitive
  ]
}

# The BGP Peering data source's `primitive` field has the routing policy
# data source (child primitive) as an embedded string:
# {
#   "type": "AttachBgpOverSubinterfacesOrSvi",
#   "data": {
#     "ipv4_afi_enabled": true,
#     "ipv6_afi_enabled": true,
#     "ttl": 1,
#     "bfd_enabled": true,
#     "password": "big secret",
#     "keepalive_time": null,
#     "hold_time": null,
#     "ipv4_addressing_type": "addressed",
#     "ipv6_addressing_type": "link_local",
#     "local_asn": null,
#     "neighbor_asn_dynamic": false,
#     "peer_from_loopback": false,
#     "peer_to": "interface_or_ip_endpoint",
#     "children": [
#       "{\"type\":\"AttachExistingRoutingPolicy\",\"data\":{\"routing_policy_id\":\"Xd5Uoo8qUjCqhihGafQ\"}}"
#     ]
#   }
# }

# Use the `primitive` element (JSON string) when declaring a parent primitive:
data "apstra_datacenter_ct_ip_link" "ip_link_with_bgp" {
  routing_zone_id      = "Zplm0niOFCCCfjaXkXo"
  vlan_id              = 3
  ipv4_addressing_type = "numbered"
  ipv6_addressing_type = "link_local"
  children = [
    data.apstra_datacenter_ct_bgp_peering_generic_system.bgp_server.primitive,
  ]
}

# The IP Link data source's `primitive` field has the primitive the BGP data
# source (child primitive) as an embedded string:
# {
#   "type": "AttachLogicalLink",
#   "data": {
#     "routing_zone_id": "Zplm0niOFCCCfjaXkXo",
#     "tagged": true,
#     "vlan_id": 3,
#     "ipv4_addressing_type": "numbered",
#     "ipv6_addressing_type": "link_local",
#     "children": [
#       "{\"type\":\"AttachBgpOverSubinterfacesOrSvi\",\"data\":{\"ipv4_afi_enabled\":true,\"ipv6_afi_enabled\":true,\"ttl\":1,\"bfd_enabled\":true,\"password\":\"big secret\",\"keepalive_time\":null,\"hold_time\":null,\"ipv4_addressing_type\":\"addressed\",\"ipv6_addressing_type\":\"link_local\",\"local_asn\":null,\"neighbor_asn_dynamic\":false,\"peer_from_loopback\":false,\"peer_to\":\"interface_or_ip_endpoint\",\"children\":[\"{\\\"type\\\":\\\"AttachExistingRoutingPolicy\\\",\\\"data\\\":{\\\"routing_policy_id\\\":\\\"Xd5Uoo8qUjCqhihGafQ\\\"}}\"]}}"
#     ]
#   }
# }

# Finally, use the IP Link's `primitive` element in a Connectivity Template:
resource "apstra_datacenter_connectivity_template" "t" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  name         = "test-net-handoff"
  description  = "ip handoff with static routes to test nets"
  tags = [
    "test",
    "terraform",
  ]
  primitives = [
    data.apstra_datacenter_ct_ip_link.ip_link_with_static_routes.primitive
  ]
}
