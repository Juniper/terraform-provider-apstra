---
page_title: "apstra_datacenter_routing_zone_constraints Data Source - terraform-provider-apstra"
subcategory: "Reference Design: Datacenter"
description: |-
  This data source returns the IDs of Routing Zone Constraints within the specified Blueprint. All of the filter attributes are optional.
---

# apstra_datacenter_routing_zone_constraints (Data Source)

This data source returns the IDs of Routing Zone Constraints within the specified Blueprint. All of the `filter` attributes are optional.


## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID.

### Optional

- `filters` (Attributes List) List of filters used to select only desired node IDs. For a node to match a filter, all specified attributes must match (each attribute within a filter is AND-ed together). The returned node IDs represent the nodes matched by all of the filters together (filters are OR-ed together). (see [below for nested schema](#nestedatt--filters))

### Read-Only

- `graph_queries` (List of String) Graph datastore queries which performed the lookup based on supplied filters.
- `ids` (Set of String) Set of Routing Zone Constraint IDs

<a id="nestedatt--filters"></a>
### Nested Schema for `filters`

Optional:

- `constraints` (Set of String) Set of Routing Zone IDs. All Routing Zones supplied here are used to match the Routing Zone Constraint, but a matching Routing Zone Constraintmay have additional Security Zones not enumerated in this set.
- `max_count_constraint` (Number) The maximum number of Routing Zones that the Application Point can be part of.
- `name` (String) Name displayed in the Apstra web UI.
- `routing_zones_list_constraint` (String) Routing Zone constraint mode. One of: `allow`, `deny`, `none`.

Read-Only:

- `blueprint_id` (String) Not applicable in filter context. Ignore.
- `id` (String) Not applicable in filter context. Ignore.
