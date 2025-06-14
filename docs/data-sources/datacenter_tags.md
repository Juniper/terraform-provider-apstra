---
page_title: "apstra_datacenter_tags Data Source - terraform-provider-apstra"
subcategory: "Reference Design: Datacenter"
description: |-
  This data source returns the IDs of Tags within the specified Blueprint.
---

# apstra_datacenter_tags (Data Source)

This data source returns the IDs of Tags within the specified Blueprint.


## Example Usage

```terraform
# This example shows how to collect the names and graph node IDs of all tags
# with description "firewall"

data "apstra_datacenter_tags" "firewall" {
  blueprint_id = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
  filters = [
    {
      description = "firewall"
    }
  ]
}

output "firewall_tags" { value = data.apstra_datacenter_tags.firewall }

# The output looks like this:
# firewall_tags = {
#   "blueprint_id" = "7427a88d-7ed4-40de-8600-9f3d57821ab6"
#   "filters" = tolist([
#     {
#       "blueprint_id" = tostring(null)
#       "description" = "firewall"
#       "name" = tostring(null)
#     },
#    ])
#    "ids" = toset([
#     "9mF7QTOjTSspgG5FPg",
#     "SyTIdNbDbm8QhBEgOw",
#   ])
#   "names" = toset([
#     "firewall-a",
#     "firewall-b",
#   ])
# }
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint to search.

### Optional

- `filters` (Attributes List) List of filters used to select only desired Tags. To match a filter, all specified attributes must match (each attribute within a filter is AND-ed together). The returned IDs represent the Tags matched by all of the filters together (filters are OR-ed together). (see [below for nested schema](#nestedatt--filters))

### Read-Only

- `ids` (Set of String) IDs of discovered `tag` Graph DB nodes.
- `names` (Set of String) Names (labels) of discovered `tag` Graph DB nodes.

<a id="nestedatt--filters"></a>
### Nested Schema for `filters`

Optional:

- `description` (String) Tag description
- `name` (String) Tag name

Read-Only:

- `blueprint_id` (String) Does not apply in filter context - ignore
