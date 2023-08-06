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

locals {
  blueprint_id               = "b726704d-f80e-4733-9103-abd6ccd8752c"
  routing_zone_constraint_id = "FxCJlvhPloupLTQK3Z8"
}

# Declare a "Routing Zone Constraint" Connectivity Template Primitive:
data "apstra_datacenter_ct_routing_zone_constraint" "default" {
  routing_zone_constraint_id = local.routing_zone_constraint_id
}

# This data source's `primitive` attribute produces JSON like this:
# {
#   "type": "AttachRoutingZoneConstraint",
#   "data": {
#     "routing_zone_constraint_id": "FxCJlvhPloupLTQK3Z8"
#   }
# }

# Use the Routing Zone Constraint Primitive in a Connectivity Template:
resource "apstra_datacenter_connectivity_template" "t" {
  blueprint_id = local.blueprint_id
  name         = "interface constrained to default RZ"
  tags = [
    "test",
    "terraform",
  ]
  primitives = [
    data.apstra_datacenter_ct_routing_zone_constraint.default.primitive
  ]
}
