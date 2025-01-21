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
