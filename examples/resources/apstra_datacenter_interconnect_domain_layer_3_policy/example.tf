# This example configures DCI Layer 3 Policy for the "green" VRF

locals { blueprint_id = "4ef73591-d23d-4f66-9ab2-16fbf156bfac" }

# Create a DCI Domain
resource "apstra_datacenter_interconnect_domain" "dci" {
  blueprint_id = local.blueprint_id
  name = "my_dci"
  route_target = "64512:100"
}

# Discover the ID of RZ "green"
data "apstra_datacenter_routing_zone" "green" {
  blueprint_id = local.blueprint_id
  name         = "green"
}

# Create a routing policy to control the DCI routing behavior for the "green" VRF
resource "apstra_datacenter_routing_policy" "green_dci" {
  blueprint_id = local.blueprint_id
  name         = "green_dci"
}

#
resource "apstra_datacenter_interconnect_domain_layer_3_policy" "green_dci" {
  blueprint_id           = local.blueprint_id
  interconnect_domain_id = apstra_datacenter_interconnect_domain.dci.id
  routing_zone_id        = data.apstra_datacenter_routing_zone.green.id
  enabled_for_type_5     = true
  routing_policy_id      = apstra_datacenter_routing_policy.green_dci.id
  route_target           = "64513:101"
}
