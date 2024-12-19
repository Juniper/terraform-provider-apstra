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
