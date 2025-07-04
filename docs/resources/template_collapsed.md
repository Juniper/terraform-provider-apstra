---
page_title: "apstra_template_collapsed Resource - terraform-provider-apstra"
subcategory: "Design"
description: |-
  This resource creates a Template for a spine-less (collapsed) Blueprint
---

# apstra_template_collapsed (Resource)

This resource creates a Template for a spine-less (collapsed) Blueprint


## Example Usage

```terraform
# This example creates a Collapsed Template based on the
# L3_collapsed_acs built-in rack type

resource "apstra_template_collapsed" "example" {
  name            = "example collapsed template"
  rack_type_id    = "L3_collapsed_acs"
  mesh_link_speed = "10G"
  mesh_link_count = 2
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `mesh_link_count` (Number) mesh_link_count integer
- `mesh_link_speed` (String) mesh_link_speed details
- `name` (String) Apstra name of the Collapsed Template.
- `rack_type_id` (String) rack type id

### Read-Only

- `id` (String) Apstra ID of the Collapsed Template.
- `rack_type` (Attributes) rack type layer details (see [below for nested schema](#nestedatt--rack_type))

<a id="nestedatt--rack_type"></a>
### Nested Schema for `rack_type`

Read-Only:

- `access_switches` (Attributes Map) Access Switches are optional, link to Leaf Switches in the same rack (see [below for nested schema](#nestedatt--rack_type--access_switches))
- `description` (String) Rack Type description, displayed in the Apstra web UI.
- `fabric_connectivity_design` (String) Must be one of 'l3clos', 'l3collapsed', 'rail_collapsed'.
- `generic_systems` (Attributes Map) Generic Systems are optional rack elements notmanaged by Apstra: Servers, routers, firewalls, etc... (see [below for nested schema](#nestedatt--rack_type--generic_systems))
- `id` (String) ID will always be `<null>` in nested contexts.
- `leaf_switches` (Attributes Map) Each Rack Type is required to have at least one Leaf Switch. (see [below for nested schema](#nestedatt--rack_type--leaf_switches))
- `name` (String) Rack Type name, displayed in the Apstra web UI.

<a id="nestedatt--rack_type--access_switches"></a>
### Nested Schema for `rack_type.access_switches`

Read-Only:

- `count` (Number) Number of Access Switches of this type.
- `esi_lag_info` (Attributes) Defines connectivity between ESI LAG peers when `redundancy_protocol` is set to `esi`. (see [below for nested schema](#nestedatt--rack_type--access_switches--esi_lag_info))
- `links` (Attributes Map) Each Access Switch is required to have at least one Link to a Leaf Switch. (see [below for nested schema](#nestedatt--rack_type--access_switches--links))
- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--rack_type--access_switches--logical_device))
- `logical_device_id` (String) ID will always be `<null>` in nested contexts.
- `redundancy_protocol` (String) Indicates whether the switch is a redundant pair.
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Access Switch (see [below for nested schema](#nestedatt--rack_type--access_switches--tags))

<a id="nestedatt--rack_type--access_switches--esi_lag_info"></a>
### Nested Schema for `rack_type.access_switches.esi_lag_info`

Required:

- `l3_peer_link_count` (Number) Count of L3 links between ESI peers.
- `l3_peer_link_speed` (String) Speed of L3 links between ESI peers.


<a id="nestedatt--rack_type--access_switches--links"></a>
### Nested Schema for `rack_type.access_switches.links`

Read-Only:

- `lag_mode` (String) LAG negotiation mode of the Link.
- `links_per_switch` (Number) Number of Links to each switch.
- `speed` (String) Speed of this Link.
- `switch_peer` (String) For non-lAG connections to redundant switch pairs, this field selects the target switch.
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Link (see [below for nested schema](#nestedatt--rack_type--access_switches--links--tags))
- `target_switch_name` (String) The `name` of the switch in this Rack Type to which this Link connects.

<a id="nestedatt--rack_type--access_switches--links--tags"></a>
### Nested Schema for `rack_type.access_switches.links.tags`

Required:

- `name` (String) Tag name field as seen in the web UI.

Optional:

- `description` (String) Tag description field as seen in the web UI.

Read-Only:

- `id` (String) Apstra ID of the Tag.



<a id="nestedatt--rack_type--access_switches--logical_device"></a>
### Nested Schema for `rack_type.access_switches.logical_device`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--rack_type--access_switches--logical_device--panels))

<a id="nestedatt--rack_type--access_switches--logical_device--panels"></a>
### Nested Schema for `rack_type.access_switches.logical_device.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--rack_type--access_switches--logical_device--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--rack_type--access_switches--logical_device--panels--port_groups"></a>
### Nested Schema for `rack_type.access_switches.logical_device.panels.port_groups`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--rack_type--access_switches--tags"></a>
### Nested Schema for `rack_type.access_switches.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.



<a id="nestedatt--rack_type--generic_systems"></a>
### Nested Schema for `rack_type.generic_systems`

Read-Only:

- `count` (Number) Number of Generic Systems of this type.
- `links` (Attributes Map) Each Generic System is required to have at least one Link to a Leaf Switch or Access Switch. (see [below for nested schema](#nestedatt--rack_type--generic_systems--links))
- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--rack_type--generic_systems--logical_device))
- `logical_device_id` (String) ID will always be `<null>` in nested contexts.
- `port_channel_id_max` (Number) Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.
- `port_channel_id_min` (Number) Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Generic System (see [below for nested schema](#nestedatt--rack_type--generic_systems--tags))

<a id="nestedatt--rack_type--generic_systems--links"></a>
### Nested Schema for `rack_type.generic_systems.links`

Read-Only:

- `lag_mode` (String) LAG negotiation mode of the Link.
- `links_per_switch` (Number) Number of Links to each switch.
- `speed` (String) Speed of this Link.
- `switch_peer` (String) For non-lAG connections to redundant switch pairs, this field selects the target switch.
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Link (see [below for nested schema](#nestedatt--rack_type--generic_systems--links--tags))
- `target_switch_name` (String) The `name` of the switch in this Rack Type to which this Link connects.

<a id="nestedatt--rack_type--generic_systems--links--tags"></a>
### Nested Schema for `rack_type.generic_systems.links.tags`

Required:

- `name` (String) Tag name field as seen in the web UI.

Optional:

- `description` (String) Tag description field as seen in the web UI.

Read-Only:

- `id` (String) Apstra ID of the Tag.



<a id="nestedatt--rack_type--generic_systems--logical_device"></a>
### Nested Schema for `rack_type.generic_systems.logical_device`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--rack_type--generic_systems--logical_device--panels))

<a id="nestedatt--rack_type--generic_systems--logical_device--panels"></a>
### Nested Schema for `rack_type.generic_systems.logical_device.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--rack_type--generic_systems--logical_device--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--rack_type--generic_systems--logical_device--panels--port_groups"></a>
### Nested Schema for `rack_type.generic_systems.logical_device.panels.port_groups`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--rack_type--generic_systems--tags"></a>
### Nested Schema for `rack_type.generic_systems.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.



<a id="nestedatt--rack_type--leaf_switches"></a>
### Nested Schema for `rack_type.leaf_switches`

Read-Only:

- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--rack_type--leaf_switches--logical_device))
- `logical_device_id` (String) ID will always be `<null>` in nested contexts.
- `mlag_info` (Attributes) Defines connectivity between MLAG peers when `redundancy_protocol` is set to `mlag`. (see [below for nested schema](#nestedatt--rack_type--leaf_switches--mlag_info))
- `redundancy_protocol` (String) Enabling a redundancy protocol converts a single Leaf Switch into a LAG-capable switch pair. Must be one of 'esi', 'mlag'.
- `spine_link_count` (Number) Links per Spine.
- `spine_link_speed` (String) Speed of Spine-facing links, something like '10G'
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Leaf Switch (see [below for nested schema](#nestedatt--rack_type--leaf_switches--tags))

<a id="nestedatt--rack_type--leaf_switches--logical_device"></a>
### Nested Schema for `rack_type.leaf_switches.logical_device`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--rack_type--leaf_switches--logical_device--panels))

<a id="nestedatt--rack_type--leaf_switches--logical_device--panels"></a>
### Nested Schema for `rack_type.leaf_switches.logical_device.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--rack_type--leaf_switches--logical_device--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--rack_type--leaf_switches--logical_device--panels--port_groups"></a>
### Nested Schema for `rack_type.leaf_switches.logical_device.panels.port_groups`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--rack_type--leaf_switches--mlag_info"></a>
### Nested Schema for `rack_type.leaf_switches.mlag_info`

Required:

- `mlag_keepalive_vlan` (Number) MLAG keepalive VLAN ID.
- `peer_link_count` (Number) Number of links between MLAG devices.
- `peer_link_port_channel_id` (Number) Port channel number used for L2 Peer Link.
- `peer_link_speed` (String) Speed of links between MLAG devices.

Optional:

- `l3_peer_link_count` (Number) Number of L3 links between MLAG devices.
- `l3_peer_link_port_channel_id` (Number) Port channel number used for L3 Peer Link. Omit to allow Apstra to choose.
- `l3_peer_link_speed` (String) Speed of l3 links between MLAG devices.


<a id="nestedatt--rack_type--leaf_switches--tags"></a>
### Nested Schema for `rack_type.leaf_switches.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.



