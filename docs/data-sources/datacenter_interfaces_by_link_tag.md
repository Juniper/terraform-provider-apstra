---
page_title: "apstra_datacenter_interfaces_by_link_tag Data Source - terraform-provider-apstra"
subcategory: "Reference Design: Datacenter"
description: |-
  This data source returns the IDs of Interfaces by Link Tag.
---

# apstra_datacenter_interfaces_by_link_tag (Data Source)

This data source returns the IDs of Interfaces by Link Tag.


## Example Usage

```terraform
# Find IDs of leaf switch interfaces with links tagged
# "dev", "linux", and "backend"
data "apstra_datacenter_interfaces_by_link_tag" "x" {
  blueprint_id = "fa6782cc-c4d5-4933-ad89-e542acd6b0c1"
  system_type  = "switch" // optional
  system_role  = "leaf"   // optional
  tags         = ["dev", "linux", "backend"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID.
- `tags` (Set of String) Set of required Tags

### Optional

- `system_role` (String) Used to further specify which interface/end of the link we're looking for whenboth ends lead to the same type. For example, on a switch-to-switch link from spine to leaf, specify either `spine` or `leaf`.
- `system_type` (String) Used to specify which interface/end of the link we're looking for. Default value is `switch`.

### Read-Only

- `graph_query` (String) The graph datastore query used to perform the lookup.
- `ids` (Set of String) A set of Apstra object IDs representing selected interfaces.
