# This example creates a Blueprint, and Interconnect Domain, and
# a remote gateway within that Interconnect Domain. All leaf switches
# peer with the remote gateway

# Create a Blueprint
resource "apstra_datacenter_blueprint" "example" {
  name        = "example"
  template_id = "L2_Virtual_EVPN"
}

# Discover Leaf switch IDs
data "apstra_datacenter_systems" "leafs" {
  blueprint_id = apstra_datacenter_blueprint.example.id
  filters      = [{ role = "leaf" }]
}

# Create an Interconnect Domain
resource "apstra_datacenter_interconnect_domain" "example" {
  blueprint_id = apstra_datacenter_blueprint.example.id
  name         = "example"
  route_target = "101:101"
}

# Create a Remote Gateway within that Interconnect Domain using
# all Leaf Switches as Local Gateways
resource "apstra_datacenter_interconnect_domain_gateway" "example" {
  blueprint_id           = apstra_datacenter_blueprint.example.id
  interconnect_domain_id = apstra_datacenter_interconnect_domain.example.id
  name                   = "example"
  asn                    = 1
  ip_address             = "1.1.1.1"
  local_gateway_nodes    = data.apstra_datacenter_systems.leafs.ids
}
