# This example looks up the details of a virtual netwwork assignment to a
# leaf switch in a datacenter blueprint.

data "apstra_datacenter_virtual_network_assignment" "example" {
  blueprint_id       = "9d129116-64f1-47a7-8aec-9b85a738e0d8"
  virtual_network_id = "ZQj_44XIXhUpliToIw"
  leaf_switch_id     = "eOZMiuxJyLCdpDStRA"
}

# Example output of this data source:
#
# example_output = {
#   "access_switch_ids" = toset([
#     "If510G__nIua9lC9Rg",
#     "ONasUURDa2GzVgckTw",
#   ])
#   "blueprint_id" = "9d129116-64f1-47a7-8aec-9b85a738e0d8"
#   "ipv4_address" = "10.0.0.2/24"
#   "ipv4_mode" = "forced"
#   "ipv6_address" = tostring(null)
#   "ipv6_mode" = "disabled"
#   "leaf_switch_id" = "eOZMiuxJyLCdpDStRA"
#   "virtual_network_id" = "ZQj_44XIXhUpliToIw"
#   "vlan" = 3
# }
