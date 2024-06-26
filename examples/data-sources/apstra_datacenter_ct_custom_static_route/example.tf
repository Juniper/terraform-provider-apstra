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
data "apstra_datacenter_ct_custom_static_route" "test-net-1" {
  routing_zone_id = "Zplm0niOFCCCfjaXkXo"
  network         = "192.0.2.0/24"
  next_hop        = "10.0.0.1"
}

data "apstra_datacenter_ct_custom_static_route" "test-net-2" {
  routing_zone_id = "Zplm0niOFCCCfjaXkXo"
  network = "198.51.100.0/24"
  next_hop        = "10.0.0.1"
}


# Each of these data source's `primitive` attribute produces JSON like this:
# {
#   "type": "AttachCustomStaticRoute",
#   "data": {
#     "routing_zone_id": "Zplm0niOFCCCfjaXkXo",
#     "network": "192.0.2.0/24",
#     "next_hop_ip_address": "10.0.0.1"
#   }
# }

# Use those `primitive` elements when declaring a parent Connectivity Template:
resource "apstra_datacenter_connectivity_template" "t" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  name         = "test-net-handoff"
  description  = "system with custom static routes to test nets"
  tags = [
    "test",
    "terraform",
  ]
  primitives = [
    data.apstra_datacenter_ct_custom_static_route.test-net-1.primitive,
    data.apstra_datacenter_ct_custom_static_route.test-net-2.primitive,
  ]
}
