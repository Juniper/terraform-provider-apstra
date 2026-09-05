# This example configures the DCI Connection Type for the "web" network

locals { blueprint_id = "96bac171-69ea-4135-92e5-455d2c3c2dae" }

# Create a DCI Domain
resource "apstra_datacenter_interconnect_domain" "dci" {
  blueprint_id = local.blueprint_id
  name = "my_dci"
  route_target = "64512:100"
}

# Discover the ID of VN "web"
data "apstra_datacenter_virtual_network" "web" {
  blueprint_id = local.blueprint_id
  name         = "web"
}

# Configure the DCI Connection Type for the "web" VN
resource "apstra_datacenter_interconnect_domain_connection_type" "web_dci" {
  blueprint_id           = local.blueprint_id
  interconnect_domain_id = apstra_datacenter_interconnect_domain.dci.id
  virtual_network_id     = data.apstra_datacenter_virtual_network.web.id
  layer_2_enabled         = true
  layer_3_enabled         = true
}
