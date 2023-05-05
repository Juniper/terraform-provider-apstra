---
page_title: "apstra_datacenter_blueprint_system_node Data Source - terraform-provider-apstra"
subcategory: ""
description: |-
  This data source returns details of a specific system Graph DB node (identified by ID) system nodes within a Blueprint.
---

# apstra_datacenter_blueprint_system_node (Data Source)

This data source returns details of a specific *system* Graph DB node (identified by ID) *system* nodes within a Blueprint.

## Example Usage

```terraform
# This example determines the hostname of the lowest numbered system using
# apstra_datacenter_blueprint_system_nodes data source with a filter to select
# system ID 1.
#
# It then uses the returned ID to do a second lookup to get the full details
# of that system and assign the hostname to the system_one_hostname local
# variable

locals {
  blueprint_id = "abc-123"
}

data "apstra_datacenter_blueprint_system_nodes" "system_one" {
  blueprint_id = local.blueprint_id
  filters = {
    system_index = 1
  }
}

data "apstra_datacenter_blueprint_system_node" "system_one" {
  blueprint_id = local.blueprint_id
  id           = one(data.apstra_datacenter_blueprint_system_nodes.system_one.ids)
}

locals {
  system_one_hostname = data.apstra_datacenter_blueprint_system_node.system_one.hostname
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID
- `id` (String) Apstra Graph DB node `ID`

### Read-Only

- `attributes` (Attributes) Attributes of a `system` Graph DB node. (see [below for nested schema](#nestedatt--attributes))

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

Read-Only:

- `hostname` (String) Apstra Graph DB node `hostname`
- `id` (String) Apstra Graph DB node ID
- `label` (String) Apstra Graph DB node `label`
- `role` (String) Apstra Graph DB node `role`
- `system_id` (String) Apstra ID of the physical system (not to be confused with its fabric role)
- `system_type` (String) Apstra Graph DB node `system_type`
- `tag_ids` (Set of String) Apstra Graph DB tags associated with this system