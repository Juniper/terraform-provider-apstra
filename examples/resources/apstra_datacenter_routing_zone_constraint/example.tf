# This example creates a Routing Zone Constraint which permits exactly one "dev"
# Routing Zone anywhere it is applied.

# First, collect all routing zone IDs in the blueprint
data "apstra_datacenter_routing_zones" "all" {
  blueprint_id = local.blueprint_id
}

# Second, collect details about each of those routing zones
data "apstra_datacenter_routing_zone" "all" {
  for_each     = data.apstra_datacenter_routing_zones.all.ids
  blueprint_id = local.blueprint_id
  id           = each.key
}

# Finally, create the Routing Zone Constraint
resource "apstra_datacenter_routing_zone_constraint" "example" {
  blueprint_id                  = local.blueprint_id
  name                          = "Permit 1 dev RZ"
  max_count_constraint          = 1
  routing_zones_list_constraint = "allow"
  # Constraints is created as a list comprehension by iterating over
  # details of each RZ in data.apstra_datacenter_routing_zone.all
  constraints = [
    for rz in data.apstra_datacenter_routing_zone.all : rz.id
    if strcontains(rz.name, "dev") // select those with "dev" in their name
  ]
}
