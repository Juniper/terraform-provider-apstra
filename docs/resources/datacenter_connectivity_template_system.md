---
page_title: "apstra_datacenter_connectivity_template_system Resource - terraform-provider-apstra"
subcategory: "Reference Design: Datacenter"
description: |-
  This resource creates a Connectivity Template suitable for use with Application Points of type system within a Datacenter Blueprint. system Application Points use the following Connectivity Template Primitive hierarchy:
  Custom Static Route
---

# apstra_datacenter_connectivity_template_system (Resource)

This resource creates a Connectivity Template suitable for use with Application Points of type *system* within a Datacenter Blueprint. *system* Application Points use the following Connectivity Template Primitive hierarchy:
 - Custom Static Route


## Example Usage

```terraform
# The following example creates a Connectivity Template compatible with
# "system" application points. It has two two Custom Static Route primitives.

locals { blueprint_id = "275769da-7b45-47d6-8f1c-49323d346bb3" }

resource "apstra_datacenter_routing_zone" "a" {
  blueprint_id = local.blueprint_id
  name         = "RZ_A"
}

resource "apstra_datacenter_routing_zone" "b" {
  blueprint_id = local.blueprint_id
  name         = "RZ_B"
}

resource "apstra_datacenter_connectivity_template_system" "DC_1" {
  blueprint_id = local.blueprint_id
  name         = "DC 1"
  description  = "Routes to 10.1.0.0/16 in RZ A and B"
  custom_static_routes = {
    (apstra_datacenter_routing_zone.a.name) = {
      routing_zone_id = apstra_datacenter_routing_zone.a.id
      network         = "10.1.0.0/16"
      next_hop        = "192.168.1.1"
    },
    (apstra_datacenter_routing_zone.b.name) = {
      routing_zone_id = apstra_datacenter_routing_zone.b.id
      network         = "10.1.0.0/16"
      next_hop        = "192.168.1.1"
    },
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Blueprint ID.
- `name` (String) Connectivity Template Name displayed in the web UI.

### Optional

- `custom_static_routes` (Attributes Map) Map of *Custom Static Route* Primitives in this Connectivity Template. (see [below for nested schema](#nestedatt--custom_static_routes))
- `description` (String) Connectivity Template Description displayed in the web UI.
- `tags` (Set of String) Set of Tags associated with this Connectivity Template.

### Read-Only

- `id` (String) Apstra graph node ID.

<a id="nestedatt--custom_static_routes"></a>
### Nested Schema for `custom_static_routes`

Required:

- `network` (String) Destination network in CIDR notation
- `next_hop` (String) Next-hop router address
- `routing_zone_id` (String) Routing Zone ID where this route should be installed

Read-Only:

- `id` (String) Unique identifier for this CT Primitive element
- `pipeline_id` (String) Unique identifier for this CT Primitive Element's upstream pipeline



