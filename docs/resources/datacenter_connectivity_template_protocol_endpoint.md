---
page_title: "apstra_datacenter_connectivity_template_protocol_endpoint Resource - terraform-provider-apstra"
subcategory: "Reference Design: Datacenter"
description: |-
  This resource creates a Connectivity Template suitable for use with Application Points of type protocol_endpoint within a Datacenter Blueprint. protocol_endpoint Application Points use the following Connectivity Template Primitive hierarchy:
  Routing Policy
---

# apstra_datacenter_connectivity_template_protocol_endpoint (Resource)

This resource creates a Connectivity Template suitable for use with Application Points of type *protocol_endpoint* within a Datacenter Blueprint. *protocol_endpoint* Application Points use the following Connectivity Template Primitive hierarchy:
 - Routing Policy


## Example Usage

```terraform
# The following example creates a Connectivity Template compatible with
# "protocol_endpoint" application points. It has three routing policy
# primitives.

locals { blueprint_id = "4ee2bdce-6d37-4ae8-928d-4463278dd637" }

resource "apstra_datacenter_connectivity_template_protocol_endpoint" "example" {
  blueprint_id = local.blueprint_id
  name         = "example connectivity template"
  routing_policies = {
    (apstra_datacenter_routing_policy.allow_10_3_0_0-16_in.name) = {
      routing_policy_id = apstra_datacenter_routing_policy.allow_10_3_0_0-16_in.id
    }
    (apstra_datacenter_routing_policy.allow_10_2_0_0-16_in.name) = {
      routing_policy_id = apstra_datacenter_routing_policy.allow_10_2_0_0-16_in.id
    }
    (apstra_datacenter_routing_policy.allow_10_1_0_0-16_out.name) = {
      routing_policy_id = apstra_datacenter_routing_policy.allow_10_1_0_0-16_out.id
    },
  }
}

resource "apstra_datacenter_routing_policy" "allow_10_1_0_0-16_out" {
  blueprint_id = local.blueprint_id
  name          = "10_1_0_0-16_out"
  import_policy = "extra_only"
  extra_exports = [{ prefix = "10.1.0.0/16", action = "permit" }]
}

resource "apstra_datacenter_routing_policy" "allow_10_2_0_0-16_in" {
  blueprint_id = local.blueprint_id
  name          = "10_2_0_0-16_in"
  import_policy = "extra_only"
  extra_imports = [{ prefix = "10.2.0.0/16", action = "permit" }]
}

resource "apstra_datacenter_routing_policy" "allow_10_3_0_0-16_in" {
  blueprint_id = local.blueprint_id
  name          = "10_3_0_0-16_in"
  import_policy = "extra_only"
  extra_imports = [{ prefix = "10.3.0.0/16", action = "permit" }]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Blueprint ID.
- `name` (String) Connectivity Template Name displayed in the web UI

### Optional

- `description` (String) Connectivity Template Description displayed in the web UI
- `routing_policies` (Attributes Map) Map of Routing Policy Primitives to be used with this *Protocol Endpoint*. (see [below for nested schema](#nestedatt--routing_policies))
- `tags` (Set of String) Set of Tags associated with this Connectivity Template

### Read-Only

- `id` (String) Apstra graph node ID.

<a id="nestedatt--routing_policies"></a>
### Nested Schema for `routing_policies`

Required:

- `routing_policy_id` (String) Routing Policy ID to be applied

Read-Only:

- `id` (String) Unique identifier for this CT Primitive element
- `pipeline_id` (String) Unique identifier for this CT Primitive Element's upstream pipeline



