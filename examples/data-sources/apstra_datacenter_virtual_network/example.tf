# Returns the data Attributes associated with the specified Virtual Network
# id
data "apstra_datacenter_virtual_network" "vnet_blue" {
  blueprint_id = "b726704d-f80e-4733-9103-abd6ccd8752c"
  name = "vnet_b"
}

locals {
  vnet_b_id  = data.apstra_datacenter_virtual_network.vnet_blue.id
}