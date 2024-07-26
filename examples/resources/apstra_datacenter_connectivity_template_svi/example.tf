# The following example creates a Connectivity Template compatible with
# "svi" application points. It has two BGP peering (IP Endpoint)
# primitives. Each BGP Peering (IP Endpoint) primitive has two routing
# policy primitives.
resource "apstra_datacenter_connectivity_template_svi" "example" {
  blueprint_id = "fe771218-3e83-450b-acb9-643567644013"
  name         = "example connectivity template"
  description  = "peer with juniper and microsoft"
  bgp_peering_ip_endpoints = [
    {
      name           = "juniper"
      neighbor_asn   = 14203
      keepalive_time = 1
      hold_time      = 3
      bfd_enabled    = true
      ipv4_address   = "192.0.2.11"
      routing_policies = [
        {
          name              = apstra_datacenter_routing_policy.allow_10_1_0_0-16_out.name
          routing_policy_id = apstra_datacenter_routing_policy.allow_10_1_0_0-16_out.id
        },
        {
          name              = apstra_datacenter_routing_policy.allow_10_2_0_0-16_in.name
          routing_policy_id = apstra_datacenter_routing_policy.allow_10_2_0_0-16_in.id
        }
      ]
    },
    {
      name           = "microsoft"
      neighbor_asn   = 8975
      keepalive_time = 1
      hold_time      = 3
      bfd_enabled    = true
      ipv6_address   = "3fff::1"
      routing_policies = [
        {
          name              = apstra_datacenter_routing_policy.allow_10_1_0_0-16_out.name
          routing_policy_id = apstra_datacenter_routing_policy.allow_10_1_0_0-16_out.id
        },
        {
          name              = apstra_datacenter_routing_policy.allow_10_3_0_0-16_in.name
          routing_policy_id = apstra_datacenter_routing_policy.allow_10_3_0_0-16_in.id
        }
      ]
    },
  ]
}

resource "apstra_datacenter_routing_policy" "allow_10_1_0_0-16_out" {
  blueprint_id  = "fe771218-3e83-450b-acb9-643567644013"
  name          = "10_1_0_0-16_out"
  import_policy = "extra_only"
  extra_exports = [{ prefix = "10.1.0.0/16", action = "permit" }]
}

resource "apstra_datacenter_routing_policy" "allow_10_2_0_0-16_in" {
  blueprint_id  = "fe771218-3e83-450b-acb9-643567644013"
  name          = "10_2_0_0-16_in"
  import_policy = "extra_only"
  extra_imports = [{ prefix = "10.2.0.0/16", action = "permit" }]
}

resource "apstra_datacenter_routing_policy" "allow_10_3_0_0-16_in" {
  blueprint_id  = "fe771218-3e83-450b-acb9-643567644013"
  name          = "10_3_0_0-16_in"
  import_policy = "extra_only"
  extra_imports = [{ prefix = "10.3.0.0/16", action = "permit" }]
}
