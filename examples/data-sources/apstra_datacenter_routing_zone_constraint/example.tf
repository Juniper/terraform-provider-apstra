# This example pulls the details of a Routing Zone Constraint
# using a "by name" lookup. Lookup by ID is also supported.

data "apstra_datacenter_routing_zone_constraint" "vasili" {
  blueprint_id = "372eca0d-41de-47cc-a17d-65f27960ca3f"
  name         = "one_zone_only"
}

output "routing_zone_constraint" {
  value = data.apstra_datacenter_routing_zone_constraint.vasili
}

# The output looks like this:

# routing_zone_constraint = {
#   "blueprint_id" = "372eca0d-41de-47cc-a17d-65f27960ca3f"
#   "constraints" = toset([
#     "6uEL07avVGEjxXYiZQ",
#     "J7ApJRAmqWOIjVCV4A",
#     "a8cU-tv0eNwj-KG-wg",
#   ])
#   "id" = "qEH5mRPjsxhuyDovLg"
#   "max_count_constraint" = 1
#   "name" = "one_zone_only"
#   "routing_zones_list_constraint" = "allow"
# }
