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

# Declare a "Virtual Network (Single)" Connectivity Template Primitive:
data "apstra_datacenter_ct_virtual_network_single" "application_a" {
  vn_id  = "xSKiJrPElmZh9_Dmnho"
  tagged = false
}

# The `primitive` output of this data source is the following JSON structure:
# {
#   "type": "AttachSingleVLAN",
#   "data": {
#     "vn_id": "xSKiJrPElmZh9_Dmnho",
#     "tagged": false
#   }
# }

# Use the `primitive` JSON when creating a Connectivity Template:
resource "apstra_datacenter_connectivity_template" "t" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  name         = "application_a"
  description  = "Application A LAN without VLAN tags"
  tags         = [
    "prod",
    "app_a",
  ]
  primitives   = [
    data.apstra_datacenter_ct_virtual_network_single.application_a.primitive
  ]
}


