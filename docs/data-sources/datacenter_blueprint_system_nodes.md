---
page_title: "apstra_datacenter_blueprint_system_nodes Data Source - terraform-provider-apstra"
subcategory: ""
description: |-
  This data source returns Graph DB node IDs of system nodes within a Blueprint.
  Optional attributes filter the result list so that it only contains IDs of nodes which match the filters.
---

# apstra_datacenter_blueprint_system_nodes (Data Source)

This data source returns Graph DB node IDs of *system* nodes within a Blueprint.

Optional attributes filter the result list so that it only contains IDs of nodes which match the filters.

## Example Usage

```terraform
# This example outputs a set of graph db node IDs representing all spine
# switches with tag 'junos' and tag 'qfx'
data "apstra_datacenter_blueprint_system_nodes" "juniper_spines" {
  blueprint_id = apstra_datacenter_blueprint.example.id
  filters = {
    role        = "spine"
    system_type = "switch"
    tag_ids     = ["junos", "qfx"]
  }
}

output "qfx_spines" {
  value = data.apstra_datacenter_blueprint_system_nodes.juniper_spines.ids
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint to search.

### Optional

- `filters` (Attributes) Filters used to select only desired node IDs. (see [below for nested schema](#nestedatt--filters))

### Read-Only

- `ids` (Set of String) IDs of matching `system` Graph DB nodes.
- `query_string` (String) Graph DB query string based on the supplied filters; possibly useful for troubleshooting.

<a id="nestedatt--filters"></a>
### Nested Schema for `filters`

Optional:

- `hostname` (String) Apstra Graph DB node `hostname`
- `id` (String) Apstra Graph DB node ID
- `label` (String) Apstra Graph DB node `label`
- `role` (String) Apstra Graph DB node `role`
- `system_id` (String) Apstra ID of the physical system (not to be confused with its fabric role)
- `system_type` (String) Apstra Graph DB node `system_type`
- `tag_ids` (Set of String) Set of Tag IDs (labels) - only nodes with all tags will match this filter