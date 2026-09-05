# This example discovers the DCI Connection Type for the "web" network

locals {
  blueprint_id         = "96bac171-69ea-4135-92e5-455d2c3c2dae"
  virtual_network_name = "web"
}

# Find the "web" Virtual Network
data "apstra_datacenter_virtual_network" "web" {
  blueprint_id = local.blueprint_id
  name = local.virtual_network_name
}

# Discover all DCI Interconnect domains (we expect exactly one)
data "apstra_datacenter_interconnect_domains" "dci" {
  blueprint_id = local.blueprint_id
}


data "apstra_datacenter_interconnect_domain_connection_type" "web" {
  blueprint_id           = local.blueprint_id
  interconnect_domain_id = one(data.apstra_datacenter_interconnect_domains.dci.ids)
  virtual_network_id        = data.apstra_datacenter_virtual_network.web.id
}

output "web_vn_connection_type" { value = data.apstra_datacenter_interconnect_domain_connection_type.web }

# The output looks like this:
#
# web_vn_connection_type = {
#   "blueprint_id" = "96bac171-69ea-4135-92e5-455d2c3c2dae"
#   "interconnect_domain_id" = "znukpHyaD1AW4Vw-ug"
#   "layer_2_enabled" = false
#   "layer_3_enabled" = false
#   "translation_vni" = tonumber(null)
#   "virtual_network_id" = "pnsJ4Sy_VOjUa8cMeg"
# }
