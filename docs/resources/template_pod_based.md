---
page_title: "apstra_template_pod_based Resource - terraform-provider-apstra"
subcategory: "Design"
description: |-
  This resource creates a Pod Based Template for a 5-stage Clos design
---

# apstra_template_pod_based (Resource)

This resource creates a Pod Based Template for a 5-stage Clos design


## Example Usage

```terraform
# This example creates a 5-stage template with two super spine
# planes and two types of pods

resource "apstra_template_pod_based" "example" {
  name                   = "dual-plane"
  super_spine = {
    logical_device_id = "AOS-24x10-2"
    per_plane_count   = 4
    plane_count       = 2
  }
  pod_infos = {
    pod_single = { count = 2 } # pod_single and pod_mlag are the IDs of the 3-stage (rack-based)
    pod_mlag   = { count = 2 } # templates used as pods in this 5-stage (pod-based) template
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Apstra name of the Pod Based Template.
- `pod_infos` (Attributes Map) Map of Pod Type info (count + details) keyed by Pod Based Template ID. (see [below for nested schema](#nestedatt--pod_infos))
- `super_spine` (Attributes) SuperSpine layer details (see [below for nested schema](#nestedatt--super_spine))

### Optional

- `fabric_link_addressing` (String) Fabric addressing scheme for Spine/SuperSpine links. Required for Apstra <= 4.1.0, not supported by Apstra >= 4.1.1.

### Read-Only

- `id` (String) Apstra ID of the Pod Based Template.

<a id="nestedatt--pod_infos"></a>
### Nested Schema for `pod_infos`

Required:

- `count` (Number) Number of instances of this Pod Type.

Read-Only:

- `pod_type` (Attributes) Pod Type attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--pod_infos--pod_type))

<a id="nestedatt--pod_infos--pod_type"></a>
### Nested Schema for `pod_infos.pod_type`

Read-Only:

- `asn_allocation_scheme` (String) "unique" is for 3-stage designs; "single" is for 5-stage designs.
- `fabric_link_addressing` (String) Fabric addressing scheme for Spine/Leaf links. Applies only to Apstra 4.1.0.
- `id` (String) ID of the pod inside the 5 stage template.
- `name` (String) Name of the pod inside the 5 stage template.
- `overlay_control_protocol` (String) Defines the inter-rack virtual network overlay protocol in the fabric.
- `rack_infos` (Attributes Map) Map of Rack Type info (count + details) (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos))
- `spine` (Attributes) Spine layer details (see [below for nested schema](#nestedatt--pod_infos--pod_type--spine))

<a id="nestedatt--pod_infos--pod_type--rack_infos"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos`

Required:

- `count` (Number) Number of instances of this Rack Type.

Read-Only:

- `rack_type` (Attributes) Rack Type attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type))

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type`

Read-Only:

- `access_switches` (Attributes Map) Access Switches are optional, link to Leaf Switches in the same rack (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches))
- `description` (String) Rack Type description, displayed in the Apstra web UI.
- `fabric_connectivity_design` (String) Must be one of 'l3clos', 'l3collapsed'.
- `generic_systems` (Attributes Map) Generic Systems are optional rack elements notmanaged by Apstra: Servers, routers, firewalls, etc... (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems))
- `id` (String) ID will always be `<null>` in nested contexts.
- `leaf_switches` (Attributes Map) Each Rack Type is required to have at least one Leaf Switch. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches))
- `name` (String) Rack Type name, displayed in the Apstra web UI.

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.access_switches`

Read-Only:

- `count` (Number) Number of Access Switches of this type.
- `esi_lag_info` (Attributes) Defines connectivity between ESI LAG peers when `redundancy_protocol` is set to `esi`. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--esi_lag_info))
- `links` (Attributes Map) Each Access Switch is required to have at least one Link to a Leaf Switch. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--links))
- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--logical_device))
- `logical_device_id` (String) ID will always be `<null>` in nested contexts.
- `redundancy_protocol` (String) Indicates whether the switch is a redundant pair.
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Access Switch (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--tags))

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--esi_lag_info"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.access_switches.tags`

Required:

- `l3_peer_link_count` (Number) Count of L3 links between ESI peers.
- `l3_peer_link_speed` (String) Speed of L3 links between ESI peers.


<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--links"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.access_switches.tags`

Read-Only:

- `lag_mode` (String) LAG negotiation mode of the Link.
- `links_per_switch` (Number) Number of Links to each switch.
- `speed` (String) Speed of this Link.
- `switch_peer` (String) For non-lAG connections to redundant switch pairs, this field selects the target switch.
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Link (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--tags--tags))
- `target_switch_name` (String) The `name` of the switch in this Rack Type to which this Link connects.

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--tags--tags"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.access_switches.tags.target_switch_name`

Required:

- `name` (String) Tag name field as seen in the web UI.

Optional:

- `description` (String) Tag description field as seen in the web UI.

Read-Only:

- `id` (String) Apstra ID of the Tag.



<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--logical_device"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.access_switches.tags`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--tags--panels))

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--tags--panels"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.access_switches.tags.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--tags--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--tags--panels--port_groups"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.access_switches.tags.panels.rows`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--access_switches--tags"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.access_switches.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.



<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.generic_systems`

Read-Only:

- `count` (Number) Number of Generic Systems of this type.
- `links` (Attributes Map) Each Generic System is required to have at least one Link to a Leaf Switch or Access Switch. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--links))
- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--logical_device))
- `logical_device_id` (String) ID will always be `<null>` in nested contexts.
- `port_channel_id_max` (Number) Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.
- `port_channel_id_min` (Number) Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Generic System (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--tags))

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--links"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.generic_systems.tags`

Read-Only:

- `lag_mode` (String) LAG negotiation mode of the Link.
- `links_per_switch` (Number) Number of Links to each switch.
- `speed` (String) Speed of this Link.
- `switch_peer` (String) For non-lAG connections to redundant switch pairs, this field selects the target switch.
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Link (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--tags--tags))
- `target_switch_name` (String) The `name` of the switch in this Rack Type to which this Link connects.

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--tags--tags"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.generic_systems.tags.target_switch_name`

Required:

- `name` (String) Tag name field as seen in the web UI.

Optional:

- `description` (String) Tag description field as seen in the web UI.

Read-Only:

- `id` (String) Apstra ID of the Tag.



<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--logical_device"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.generic_systems.tags`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--tags--panels))

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--tags--panels"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.generic_systems.tags.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--tags--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--tags--panels--port_groups"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.generic_systems.tags.panels.rows`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--generic_systems--tags"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.generic_systems.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.



<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.leaf_switches`

Read-Only:

- `logical_device` (Attributes) Logical Device attributes cloned from the Global Catalog at creation time. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--logical_device))
- `logical_device_id` (String) ID will always be `<null>` in nested contexts.
- `mlag_info` (Attributes) Defines connectivity between MLAG peers when `redundancy_protocol` is set to `mlag`. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--mlag_info))
- `redundancy_protocol` (String) Enabling a redundancy protocol converts a single Leaf Switch into a LAG-capable switch pair. Must be one of 'esi', 'mlag'.
- `spine_link_count` (Number) Links per Spine.
- `spine_link_speed` (String) Speed of Spine-facing links, something like '10G'
- `tag_ids` (Set of String) IDs will always be `<null>` in nested contexts.
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to this Leaf Switch (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--tags))

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--logical_device"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.leaf_switches.tags`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--tags--panels))

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--tags--panels"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.leaf_switches.tags.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--tags--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--tags--panels--port_groups"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.leaf_switches.tags.panels.rows`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--mlag_info"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.leaf_switches.tags`

Required:

- `mlag_keepalive_vlan` (Number) MLAG keepalive VLAN ID.
- `peer_link_count` (Number) Number of links between MLAG devices.
- `peer_link_port_channel_id` (Number) Port channel number used for L2 Peer Link.
- `peer_link_speed` (String) Speed of links between MLAG devices.

Optional:

- `l3_peer_link_count` (Number) Number of L3 links between MLAG devices.
- `l3_peer_link_port_channel_id` (Number) Port channel number used for L3 Peer Link. Omit to allow Apstra to choose.
- `l3_peer_link_speed` (String) Speed of l3 links between MLAG devices.


<a id="nestedatt--pod_infos--pod_type--rack_infos--rack_type--leaf_switches--tags"></a>
### Nested Schema for `pod_infos.pod_type.rack_infos.rack_type.leaf_switches.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.





<a id="nestedatt--pod_infos--pod_type--spine"></a>
### Nested Schema for `pod_infos.pod_type.spine`

Required:

- `count` (Number) Number of Spine Switches.
- `logical_device_id` (String) Apstra Object ID of the Logical Device used to model this Spine Switch.

Optional:

- `super_spine_link_count` (Number) Count of links to each super Spine switch.
- `super_spine_link_speed` (String) Speed of links to super Spine switches.
- `tag_ids` (Set of String) Set of Tag IDs to be applied to Spine Switches

Read-Only:

- `logical_device` (Attributes) Logical Device attributes as represented in the Global Catalog. (see [below for nested schema](#nestedatt--pod_infos--pod_type--spine--logical_device))
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to Spine Switches (see [below for nested schema](#nestedatt--pod_infos--pod_type--spine--tags))

<a id="nestedatt--pod_infos--pod_type--spine--logical_device"></a>
### Nested Schema for `pod_infos.pod_type.spine.tags`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--pod_infos--pod_type--spine--tags--panels))

<a id="nestedatt--pod_infos--pod_type--spine--tags--panels"></a>
### Nested Schema for `pod_infos.pod_type.spine.tags.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--pod_infos--pod_type--spine--tags--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--pod_infos--pod_type--spine--tags--panels--port_groups"></a>
### Nested Schema for `pod_infos.pod_type.spine.tags.panels.rows`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--pod_infos--pod_type--spine--tags"></a>
### Nested Schema for `pod_infos.pod_type.spine.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.





<a id="nestedatt--super_spine"></a>
### Nested Schema for `super_spine`

Required:

- `logical_device_id` (String) Apstra Object ID of the Logical Device used to model this Spine Switch.
- `per_plane_count` (Number) Number of Super Spine switches per plane.

Optional:

- `plane_count` (Number) Permits creation of multi-planar 5-stage topologies. Default: 1
- `tag_ids` (Set of String) Set of Tag IDs to be applied to SuperSpine Switches

Read-Only:

- `logical_device` (Attributes) Logical Device attributes as represented in the Global Catalog. (see [below for nested schema](#nestedatt--super_spine--logical_device))
- `tags` (Attributes Set) Set of Tags (Name + Description) applied to SuperSpine Switches (see [below for nested schema](#nestedatt--super_spine--tags))

<a id="nestedatt--super_spine--logical_device"></a>
### Nested Schema for `super_spine.logical_device`

Read-Only:

- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Logical device display name.
- `panels` (Attributes List) Details physical layout of interfaces on the device. (see [below for nested schema](#nestedatt--super_spine--logical_device--panels))

<a id="nestedatt--super_spine--logical_device--panels"></a>
### Nested Schema for `super_spine.logical_device.panels`

Read-Only:

- `columns` (Number) Physical horizontal dimension of the panel.
- `port_groups` (Attributes List) Ordered logical groupings of interfaces by speed or purpose within a panel (see [below for nested schema](#nestedatt--super_spine--logical_device--panels--port_groups))
- `rows` (Number) Physical vertical dimension of the panel.

<a id="nestedatt--super_spine--logical_device--panels--port_groups"></a>
### Nested Schema for `super_spine.logical_device.panels.rows`

Read-Only:

- `port_count` (Number) Number of ports in the group.
- `port_roles` (Set of String) One or more of: access, generic, l3_server, leaf, peer, server, spine, superspine and unused.
- `port_speed` (String) Port speed.




<a id="nestedatt--super_spine--tags"></a>
### Nested Schema for `super_spine.tags`

Read-Only:

- `description` (String) Tag description field as seen in the web UI.
- `id` (String) ID will always be `<null>` in nested contexts.
- `name` (String) Tag name field as seen in the web UI.


