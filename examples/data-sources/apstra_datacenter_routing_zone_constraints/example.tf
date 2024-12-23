# This example uses filters to find the ID of every Routing Zone
# Constraint which ether allows the routing zone named "dev-1"
# or allows the routing zone named "dev-2"

data "apstra_datacenter_routing_zone" "dev-1" {
  blueprint_id = "372eca0d-41de-47cc-a17d-65f27960ca3f"
  name = "dev-1"
}

data "apstra_datacenter_routing_zone" "dev-2" {
  blueprint_id = "372eca0d-41de-47cc-a17d-65f27960ca3f"
  name = "dev-2"
}

data "apstra_datacenter_routing_zone_constraints" "allow_dev_1_or_dev_2" {
  blueprint_id = "372eca0d-41de-47cc-a17d-65f27960ca3f"
  filters = [
    {
      routing_zones_list_constraint = "allow"
      constraints = [data.apstra_datacenter_routing_zone.dev-1.id]
    },
    {
      routing_zones_list_constraint = "allow"
      constraints = [data.apstra_datacenter_routing_zone.dev-2.id]
    },
  ]
}

output "constraint_allowing_dev_1_or_dev_2" {
  value = data.apstra_datacenter_routing_zone_constraints.allow_dev_1_or_dev_2
}

# The output looks like this:
# constraint_allowing_dev_1_or_dev_2 = {
#   "blueprint_id" = "372eca0d-41de-47cc-a17d-65f27960ca3f"
#   "filters" = tolist([
#     {
#       "blueprint_id" = tostring(null)
#       "constraints" = toset([
#         "a8cU-tv0eNwj-KG-wg",
#       ])
#       "id" = tostring(null)
#       "max_count_constraint" = tonumber(null)
#       "name" = tostring(null)
#       "routing_zones_list_constraint" = "allow"
#     },
#     {
#       "blueprint_id" = tostring(null)
#       "constraints" = toset([
#         "6uEL07avVGEjxXYiZQ",
#       ])
#       "id" = tostring(null)
#       "max_count_constraint" = tonumber(null)
#       "name" = tostring(null)
#       "routing_zones_list_constraint" = "allow"
#     },
#   ])
#   "graph_queries" = tolist([
#     "match(node(name='n_routing_zone_constraint',type='routing_zone_constraint',routing_zones_list_constraint='allow'),node(name='n_routing_zone_constraint').out(type='constraint').node(type='security_zone',id='a8cU-tv0eNwj-KG-wg'))",
#     "match(node(name='n_routing_zone_constraint',type='routing_zone_constraint',routing_zones_list_constraint='allow'),node(name='n_routing_zone_constraint').out(type='constraint').node(type='security_zone',id='6uEL07avVGEjxXYiZQ'))",
#   ])
#   "ids" = toset([
#     "nbe8Ly6zUwXWWdGMjQ",
#     "qEH5mRPjsxhuyDovLg",
#   ])
# }
