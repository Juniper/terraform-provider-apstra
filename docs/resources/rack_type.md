---
page_title: "apstra_rack_type Resource - terraform-provider-apstra"
subcategory: "Design"
description: |-
  This resource creates a Rack Type in the Apstra Design tab.
---

# apstra_rack_type (Resource)

This resource creates a Rack Type in the Apstra Design tab.


## Example Usage

```terraform
resource "apstra_rack_type" "example" {
  name                       = "example rack type"
  description                = "Created by Terraform"
  fabric_connectivity_design = "l3clos"
  leaf_switches = { // leaf switches are a map keyed by switch name, so
    leaf_switch = { // "leaf switch" on this line is the name used by links targeting this switch.
      logical_device_id   = "AOS-24x10-2"
      spine_link_count    = 1
      spine_link_speed    = "10G"
      redundancy_protocol = "esi"
    }
  }
  access_switches = { // access switches are a map keyed by switch name, so
    access_switch = { // "access_switch" on this line is the name used by links targeting this switch.
      logical_device_id = "AOS-24x10-2"
      count             = 1
      esi_lag_info = {
        l3_peer_link_count = 1
        l3_peer_link_speed = "10G"
      }
      links = {
        leaf_switch = {
          speed              = "10G"
          target_switch_name = "leaf_switch" // note "leaf_switch" corresponds to a map key above.
          links_per_switch   = 1
        }
      }
    }
  }
  generic_systems = {
    webserver = {
      count             = 2
      logical_device_id = "AOS-4x10-1"
      links = {
        link = {
          speed              = "10G"
          target_switch_name = "access_switch" // note "access_switch" corresponds to a map key above.
          lag_mode           = "lacp_active"
          switch_peer        = "first"
        }
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `fabric_connectivity_design` (String) Must be one of 'l3clos', 'l3collapsed', 'rail_collapsed'.
- `leaf_switches` (Attributes Map) Each Rack Type is required to have at least one Leaf Switch. (see [below for nested schema](#nestedatt--leaf_switches))
- `name` (String) Rack Type name, displayed in the Apstra web UI.

### Optional

- `access_switches` (Attributes Map) Access Switches are optional, link to Leaf Switches in the same rack (see [below for nested schema](#nestedatt--access_switches))
- `description` (String) Rack Type description, displayed in the Apstra web UI.
- `generic_systems` (Attributes Map) Generic Systems are optional rack elements notmanaged by Apstra: Servers, routers, firewalls, etc... (see [below for nested schema](#nestedatt--generic_systems))

### Read-Only

- `id` (String) Object ID for the Rack Type, assigned by Apstra.

<a id="nestedatt--leaf_switches"></a>
### Nested Schema for `leaf_switches`

Required:

- `logical_device_id` (String) Apstra Object ID of the Logical Device used to model this Leaf Switch.

Optional:

- `mlag_info` (Attributes) Required when `redundancy_protocol` set to `mlag`, defines the connectivity between MLAG peers. (see [below for nested schema](#nestedatt--leaf_switches--mlag_info))
- `redundancy_protocol` (String) Enabling a redundancy protocol converts a single Leaf Switch into a LAG-capable switch pair. Must be one of 'esi', 'mlag'.
- `spine_link_count` (Number) Links per Spine.
- `spine_link_speed` (String) Speed of Spine-facing links, something like '10G'
- `tag_ids` (Set of String) Set of Tag IDs to be applied to this Leaf Switch

Read-Only:

- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--leaf_switches--logical_device))
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Leaf Switch (see [below for nested schema](#nestedatt--leaf_switches--tags))

<a id="nestedatt--leaf_switches--mlag_info"></a>
### Nested Schema for `leaf_switches.mlag_info`

Required:

- `mlag_keepalive_vlan` (Number) MLAG keepalive VLAN ID.
- `peer_link_count` (Number) Number of links between MLAG devices.
- `peer_link_port_channel_id` (Number) Port channel number used for L2 Peer Link.
- `peer_link_speed` (String) Speed of links between MLAG devices.

Optional:

- `l3_peer_link_count` (Number) Number of L3 links between MLAG devices.
- `l3_peer_link_port_channel_id` (Number) Port channel number used for L3 Peer Link. Omit to allow Apstra to choose.
- `l3_peer_link_speed` (String) Speed of l3 links between MLAG devices.


<a id="nestedatt--leaf_switches--logical_device"></a>
### Nested Schema for `leaf_switches.logical_device`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--leaf_switches--logical_device--panels))

<a id="nestedatt--leaf_switches--logical_device--panels"></a>
### Nested Schema for `leaf_switches.logical_device.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--leaf_switches--logical_device--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--leaf_switches--logical_device--panels--port_groups"></a>
### Nested Schema for `leaf_switches.logical_device.panels.port_groups`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--leaf_switches--tags"></a>
### Nested Schema for `leaf_switches.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.



<a id="nestedatt--access_switches"></a>
### Nested Schema for `access_switches`

Required:

- `count` (Number) Number of Access Switches of this type.
- `links` (Attributes Map) Each Access Switch is required to have at least one Link to a Leaf Switch. (see [below for nested schema](#nestedatt--access_switches--links))
- `logical_device_id` (String) Apstra Object ID of the Logical Device used to model this Access Switch.

Optional:

- `esi_lag_info` (Attributes) Including this stanza converts the Access Switch into a redundant pair. (see [below for nested schema](#nestedatt--access_switches--esi_lag_info))
- `tag_ids` (Set of String) Set of Tag IDs to be applied to this Access Switch

Read-Only:

- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--access_switches--logical_device))
- `redundancy_protocol` (String) Indicates whether the switch is a redundant pair.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Access Switch (see [below for nested schema](#nestedatt--access_switches--tags))

<a id="nestedatt--access_switches--links"></a>
### Nested Schema for `access_switches.links`

Required:

- `speed` (String) Speed of this Link.
- `target_switch_name` (String) The `name` of the switch in this Rack Type to which this Link connects.

Optional:

- `lag_mode` (String) LAG negotiation mode of the Link.
- `links_per_switch` (Number) Number of Links to each switch.
- `switch_peer` (String) For non-lAG connections to redundant switch pairs, this field selects the target switch.
- `tag_ids` (Set of String) Set of Tag IDs to be applied to this Link

Read-Only:

- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Link (see [below for nested schema](#nestedatt--access_switches--links--tags))

<a id="nestedatt--access_switches--links--tags"></a>
### Nested Schema for `access_switches.links.tags`

Required:

- `name` (String) Tag name field as seen in the web UI.

Optional:

- `description` (String) Tag description field as seen in the web UI.

Read-Only:

- `id` (String) Apstra ID of the Tag.



<a id="nestedatt--access_switches--esi_lag_info"></a>
### Nested Schema for `access_switches.esi_lag_info`

Required:

- `l3_peer_link_count` (Number) Count of L3 links between ESI peers.
- `l3_peer_link_speed` (String) Speed of L3 links between ESI peers.


<a id="nestedatt--access_switches--logical_device"></a>
### Nested Schema for `access_switches.logical_device`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--access_switches--logical_device--panels))

<a id="nestedatt--access_switches--logical_device--panels"></a>
### Nested Schema for `access_switches.logical_device.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--access_switches--logical_device--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--access_switches--logical_device--panels--port_groups"></a>
### Nested Schema for `access_switches.logical_device.panels.port_groups`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--access_switches--tags"></a>
### Nested Schema for `access_switches.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.



<a id="nestedatt--generic_systems"></a>
### Nested Schema for `generic_systems`

Required:

- `count` (Number) Number of Generic Systems of this type.
- `links` (Attributes Map) Each Generic System is required to have at least one Link to a Leaf Switch or Access Switch. (see [below for nested schema](#nestedatt--generic_systems--links))
- `logical_device_id` (String) Apstra Object ID of the Logical Device used to model this Generic System.

Optional:

- `port_channel_id_max` (Number) Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.
- `port_channel_id_min` (Number) Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.
- `tag_ids` (Set of String) Set of Tag IDs to be applied to this Generic System

Read-Only:

- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--generic_systems--logical_device))
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Generic System (see [below for nested schema](#nestedatt--generic_systems--tags))

<a id="nestedatt--generic_systems--links"></a>
### Nested Schema for `generic_systems.links`

Required:

- `speed` (String) Speed of this Link.
- `target_switch_name` (String) The `name` of the switch in this Rack Type to which this Link connects.

Optional:

- `lag_mode` (String) LAG negotiation mode of the Link.
- `links_per_switch` (Number) Number of Links to each switch.
- `switch_peer` (String) For non-lAG connections to redundant switch pairs, this field selects the target switch.
- `tag_ids` (Set of String) Set of Tag IDs to be applied to this Link

Read-Only:

- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Link (see [below for nested schema](#nestedatt--generic_systems--links--tags))

<a id="nestedatt--generic_systems--links--tags"></a>
### Nested Schema for `generic_systems.links.tags`

Required:

- `name` (String) Tag name field as seen in the web UI.

Optional:

- `description` (String) Tag description field as seen in the web UI.

Read-Only:

- `id` (String) Apstra ID of the Tag.



<a id="nestedatt--generic_systems--logical_device"></a>
### Nested Schema for `generic_systems.logical_device`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--generic_systems--logical_device--panels))

<a id="nestedatt--generic_systems--logical_device--panels"></a>
### Nested Schema for `generic_systems.logical_device.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--generic_systems--logical_device--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--generic_systems--logical_device--panels--port_groups"></a>
### Nested Schema for `generic_systems.logical_device.panels.port_groups`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--generic_systems--tags"></a>
### Nested Schema for `generic_systems.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.



