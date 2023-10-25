---
page_title: "apstra_datacenter_svis_map Data Source - terraform-provider-apstra"
subcategory: ""
description: |-
  This data source returns a maps of Sets of SVI info keyed by Virtual Network ID, System ID and SVI ID.
---

# apstra_datacenter_svis_map (Data Source)

This data source returns a maps of Sets of SVI info keyed by Virtual Network ID, System ID and SVI ID.

## Example Usage

```terraform
# The apstra_datacenter_svis_map data source returns maps detailing SVIs in
# in the given blueprint. There are three maps:
#
# - the `by_id` map is a map of SVI details keyed by SVI ID
# - the `by_system` map is a map of sets of SVI details keyed by system id
# - the `by_virtual_network` map is a map of sets of SVI details keyed by
#   virtual network ID

data "apstra_datacenter_svis_map" "test" {
  blueprint_id = "de595eeb-2c32-4d1f-9f17-350b45cda35f"
}

# The data source makes the following available
# {
#   "blueprint_id" = "de595eeb-2c32-4d1f-9f17-350b45cda35f"
#   "by_id" = tomap({                         <- map of SVI details keyed by SVI ID
#     "4fic8g4UlJbZ0kiua9A" = {               <- key (the SVI ID)
#       "id" = "4fic8g4UlJbZ0kiua9A"
#       "ipv4_addr" = tostring(null)
#       "ipv4_mode" = "enabled"
#       "ipv6_addr" = tostring(null)
#       "ipv6_mode" = "disabled"
#       "name" = ""
#       "system_id" = "nCpG01EIbNzmhkr7AAQ"
#       "virtual_network_id" = "wN9cLDwUxXq-VlUBqyM"
#     }
#     "JCOZaWIJysxwlLsTmh4" = {               <- key (the SVI ID)
#       "id" = "JCOZaWIJysxwlLsTmh4"
#       "ipv4_addr" = tostring(null)
#       "ipv4_mode" = "enabled"
#       "ipv6_addr" = tostring(null)
#       "ipv6_mode" = "disabled"
#       "name" = ""
#       "system_id" = "nCpG01EIbNzmhkr7AAQ"
#       "virtual_network_id" = "1yPFPh010dCXY16cvl4"
#     }
#     "WE4VCweBjMD790ceEOE" = {               <- key (the SVI ID)
#       "id" = "WE4VCweBjMD790ceEOE"
#       "ipv4_addr" = tostring(null)
#       "ipv4_mode" = "enabled"
#       "ipv6_addr" = tostring(null)
#       "ipv6_mode" = "disabled"
#       "name" = "irb.3"
#       "system_id" = "QHAWJdWP3m8U81uGWb8"
#       "virtual_network_id" = "1yPFPh010dCXY16cvl4"
#     }
#     "veaWrSOWpM4BJWbyWYI" = {               <- key (the SVI ID)
#       "id" = "veaWrSOWpM4BJWbyWYI"
#       "ipv4_addr" = tostring(null)
#       "ipv4_mode" = "enabled"
#       "ipv6_addr" = tostring(null)
#       "ipv6_mode" = "disabled"
#       "name" = "irb.4"
#       "system_id" = "QHAWJdWP3m8U81uGWb8"
#       "virtual_network_id" = "wN9cLDwUxXq-VlUBqyM"
#     }
#   })
#   "by_system" = tomap({                     <- map of sets of SVI details keyed by system ID
#     "QHAWJdWP3m8U81uGWb8" = toset([         <- key (the System ID)
#       {
#         "id" = "WE4VCweBjMD790ceEOE"        <- this system's first SVI ID
#         "ipv4_addr" = tostring(null)
#         "ipv4_mode" = "enabled"
#         "ipv6_addr" = tostring(null)
#         "ipv6_mode" = "disabled"
#         "name" = "irb.3"
#         "system_id" = "QHAWJdWP3m8U81uGWb8"
#         "virtual_network_id" = "1yPFPh010dCXY16cvl4"
#       },
#       {
#         "id" = "veaWrSOWpM4BJWbyWYI"        <- this system's second SVI ID
#         "ipv4_addr" = tostring(null)
#         "ipv4_mode" = "enabled"
#         "ipv6_addr" = tostring(null)
#         "ipv6_mode" = "disabled"
#         "name" = "irb.4"
#         "system_id" = "QHAWJdWP3m8U81uGWb8"
#         "virtual_network_id" = "wN9cLDwUxXq-VlUBqyM"
#       },
#     ])
#     "nCpG01EIbNzmhkr7AAQ" = toset([         <- key (the System ID)
#       {
#         "id" = "4fic8g4UlJbZ0kiua9A"        <- this system's first SVI ID
#         "ipv4_addr" = tostring(null)
#         "ipv4_mode" = "enabled"
#         "ipv6_addr" = tostring(null)
#         "ipv6_mode" = "disabled"
#         "name" = ""
#         "system_id" = "nCpG01EIbNzmhkr7AAQ"
#         "virtual_network_id" = "wN9cLDwUxXq-VlUBqyM"
#       },
#       {
#         "id" = "JCOZaWIJysxwlLsTmh4"        <- this system's second SVI ID
#         "ipv4_addr" = tostring(null)
#         "ipv4_mode" = "enabled"
#         "ipv6_addr" = tostring(null)
#         "ipv6_mode" = "disabled"
#         "name" = ""
#         "system_id" = "nCpG01EIbNzmhkr7AAQ"
#         "virtual_network_id" = "1yPFPh010dCXY16cvl4"
#       },
#     ])
#   })
#   "by_virtual_network" = tomap({            <- map of sets of SVI details keyed by virtual network ID
#     "1yPFPh010dCXY16cvl4" = toset([         <- key (the virtual network ID)
#       {
#         "id" = "JCOZaWIJysxwlLsTmh4"        <- this network's second SVI ID
#         "ipv4_addr" = tostring(null)
#         "ipv4_mode" = "enabled"
#         "ipv6_addr" = tostring(null)
#         "ipv6_mode" = "disabled"
#         "name" = ""
#         "system_id" = "nCpG01EIbNzmhkr7AAQ"
#         "virtual_network_id" = "1yPFPh010dCXY16cvl4"
#       },
#       {
#         "id" = "WE4VCweBjMD790ceEOE"        <- this network's second SVI ID
#         "ipv4_addr" = tostring(null)
#         "ipv4_mode" = "enabled"
#         "ipv6_addr" = tostring(null)
#         "ipv6_mode" = "disabled"
#         "name" = "irb.3"
#         "system_id" = "QHAWJdWP3m8U81uGWb8"
#         "virtual_network_id" = "1yPFPh010dCXY16cvl4"
#       },
#     ])
#     "wN9cLDwUxXq-VlUBqyM" = toset([         <- key (the virtual network ID)
#       {
#         "id" = "4fic8g4UlJbZ0kiua9A"        <- this network's second SVI ID
#         "ipv4_addr" = tostring(null)
#         "ipv4_mode" = "enabled"
#         "ipv6_addr" = tostring(null)
#         "ipv6_mode" = "disabled"
#         "name" = ""
#         "system_id" = "nCpG01EIbNzmhkr7AAQ"
#         "virtual_network_id" = "wN9cLDwUxXq-VlUBqyM"
#       },
#       {
#         "id" = "veaWrSOWpM4BJWbyWYI"        <- this network's second SVI ID
#         "ipv4_addr" = tostring(null)
#         "ipv4_mode" = "enabled"
#         "ipv6_addr" = tostring(null)
#         "ipv6_mode" = "disabled"
#         "name" = "irb.4"
#         "system_id" = "QHAWJdWP3m8U81uGWb8"
#         "virtual_network_id" = "wN9cLDwUxXq-VlUBqyM"
#       },
#     ])
#   })
#   "graph_query" = "match(node(type='virtual_network',name='n_virtual_network').out(type='instantiated_by').node(type='vn_instance',name='n_vn_instance'),node(type='system',name='n_system',system_type='switch').out(type='hosted_vn_instances').node(type='vn_instance',name='n_vn_instance').out(type='member_interfaces').node(type='interface',name='n_interface'))"
# }
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID.

### Read-Only

- `by_id` (Map of Object) A map of sets of SVI info keyed by SVI ID. (see [below for nested schema](#nestedatt--by_id))
- `by_system` (Map of Set of Object) A map of sets of SVI info keyed by System ID.
- `by_virtual_network` (Map of Set of Object) A map of sets of SVI info keyed by Virtual Network ID.
- `graph_query` (String) The graph datastore query used to perform the lookup.

<a id="nestedatt--by_id"></a>
### Nested Schema for `by_id`

Read-Only:

- `id` (String)
- `ipv4_addr` (String)
- `ipv4_mode` (String)
- `ipv6_addr` (String)
- `ipv6_mode` (String)
- `name` (String)
- `system_id` (String)
- `virtual_network_id` (String)
