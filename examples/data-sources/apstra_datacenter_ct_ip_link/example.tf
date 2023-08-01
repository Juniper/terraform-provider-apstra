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

# Declare a couple of "Static Route" Connectivity Template Primitives:
data "apstra_datacenter_ct_static_route" "test-net-1" {
  network           = "192.0.2.0/24"
}

data "apstra_datacenter_ct_static_route" "test-net-2" {
  network           = "198.51.100.0/24"
}

# Each of these data source's `primitive` attribute produces JSON like this:
# {
#   "type": "AttachStaticRoute",
#   "data": {
#     "network": "192.0.2.0/24",
#     "share_ip_endpoint": false
#   }
# }

# Use those `primitive` elements when declaring a parent Primitive:
data "apstra_datacenter_ct_ip_link" "ip_link_with_static_routes" {
  routing_zone_id      = "Zplm0niOFCCCfjaXkXo"
  vlan_id              = 3
  ipv4_addressing_type = "numbered"
  ipv6_addressing_type = "link_local"
  child_primitives = [
    data.apstra_datacenter_ct_static_route.test-net-1.primitive,
    data.apstra_datacenter_ct_static_route.test-net-2.primitive,
  ]
}

# The IP Link data source's `primitive` field has the primitives of two static
# routes (child_primitives) as embedded strings:
# {
#   "type": "AttachLogicalLink",
#   "data": {
#     "routing_zone_id": "Zplm0niOFCCCfjaXkXo",
#     "tagged": true,
#     "vlan_id": 3,
#     "ipv4_addressing_type": "numbered",
#     "ipv6_addressing_type": "link_local",
#     "child_primitives": [
#       "{\"type\":\"AttachStaticRoute\",\"data\":{\"network\":\"192.0.2.0/24\",\"share_ip_endpoint\":false}}",
#       "{\"type\":\"AttachStaticRoute\",\"data\":{\"network\":\"198.51.100.0/24\",\"share_ip_endpoint\":false}}"
#     ]
#   }
# }

# Finally, use the IP Link's `primitive` element in a Connectivity Template:
resource "apstra_datacenter_connectivity_template" "t" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  name         = "test-net-handoff"
  description  = "ip handoff with static routes to test nets"
  tags         = [
    "test",
    "terraform",
  ]
  primitives   = [
    data.apstra_datacenter_ct_ip_link.ip_link_with_static_routes.primitive
  ]
}
