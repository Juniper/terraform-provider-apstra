---
page_title: "apstra_datacenter_virtual_network_binding_constructor Data Source - terraform-provider-apstra"
subcategory: "Reference Design: Datacenter"
description: |-
  This data source can be used to calculate the bindings data required by apstra_datacenter_virtual_network.
  Given a list of switch node IDSs, it determines whether they're leaf or access nodes, replaces individual switch IDs with ESI or MLAG redundancy group IDs, finds required parent leaf switches of all access switches.
---

# apstra_datacenter_virtual_network_binding_constructor (Data Source)

This data source can be used to calculate the `bindings` data required by `apstra_datacenter_virtual_network`.

Given a list of switch node IDSs, it determines whether they're leaf or access nodes, replaces individual switch IDs with ESI or MLAG redundancy group IDs, finds required parent leaf switches of all access switches.


## Example Usage

```terraform
# This example creates the `bindings` data required by the
# `apstra_datacenter_virtual_network` resource when making a new Virtual Network
# for the 'blueprint_xyz' available on the switches marked with an '*' in the
# following topology:
#
#                                  blueprint_xyz
#
#  +--------------+  + -------------+   +------------+            +------------+
#  |              |  |              |   |          * |            |            |
#  |  esi-leaf-a  |  |  esi-leaf-b  |   |   leaf-c   |            |   leaf-d   |
#  |              |  |              |   |            |            |            |
#  +-+----------+-+  +-+----------+-+   +-----+------+            +-+--------+-+
#    |           \    /           |           |                     |        |
#    |            \  /            |           |                     |        |
#    |             \/             |           |                     |        |
#    |             /\             |           |                     |        |
#    |            /  \            |           |                     |        |
#    |           /    \           |           |                     |        |
#  +-+----------+-+  +-+----------+-+   +-----+------+   +----------+-+    +-+----------+
#  |            * |  |              |   |            |   |          * |    |          * |
#  | esi-access-a +--+ esi-access-b |   |  access-c  |   |  access-d  |    |  access-e  |
#  |              |  |              |   |            |   |            |    |            |
#  +--------------+  +--------------+   +------------+   +------------+    +------------+

data "apstra_datacenter_virtual_network_binding_constructor" "example" {
  blueprint_id = "blueprint_xyz"
  vlan_id      = 5 // optional; default behavior allows Apstra to choose
  switch_ids   = [ "esi-access-a", "leaf-c", "access-d", "access-e"]
}

# The output looks like the following:
#
#  bindings = {
#    // the ESI leaf group is included because of a downstream requirement
#    "esi-leaf-group-a-b" = {
#      // the ESI access group has been substituted in place of 'esi-access-a'
#      "access_ids" = [ "esi-access-group-a-b" ]
#      "vlan_id" = 5
#    },
#    // this leaf included directly without downstream switches
#    "leaf-c" = {
#      "access_ids" = []
#      "vlan_id" = 5
#    }
#    // this leaf included because of a downstream requirement
#    "leaf-d" = {
#      "access_ids" = [ "access-d", "access-e" ]
#      "vlan_id" = 5
#    }
#  }
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID. Used to determine the redundancy group and access/leaf relationships of each specified switch ID.
- `switch_ids` (Set of String) Set of graph db node IDs representing access and/or leaf switches for which a binding should be constructed.

### Optional

- `vlan_id` (Number) VLAN ID will be populated directly into the `bindings` output.

### Read-Only

- `bindings` (Attributes Map) A map of bindings appropriate for use in a `apstra_datacenter_virtual_network` resource. (see [below for nested schema](#nestedatt--bindings))

<a id="nestedatt--bindings"></a>
### Nested Schema for `bindings`

Read-Only:

- `access_ids` (Set of String) A set of zero or more graph db node IDs representing Access Switch `system` nodes or a `redundancy_group` nodes.
- `vlan_id` (Number) The value supplied as `vlan_id` at the root of this datasource configuration, if any. May be `null`, in which case Apstra will choose.
