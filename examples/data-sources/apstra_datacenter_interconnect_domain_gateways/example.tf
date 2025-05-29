# This example outputs the names of Interconnect Domain Gateways which
# use Leaf 1 as the local gateway end.

# Discover the ID of Leaf 1
data "apstra_datacenter_systems" "leaf_1" {
  blueprint_id = "eb5bf749-3610-4c23-ade1-50a6cc200abf"
  filters      = [{ label = "leaf_1" }]
}

# Discover IDs of local Interconnect Domain Gateways peering with Leaf 1
data "apstra_datacenter_interconnect_domain_gateways" "with_leaf_1" {
  blueprint_id = "eb5bf749-3610-4c23-ade1-50a6cc200abf"
  filters      = [{ local_gateway_nodes = [one(data.apstra_datacenter_systems.leaf_1.ids)] }]
}

# Discover details (we need the name) of all local Interconnect Domain
# Gatways peering with Leaf 1
data "apstra_datacenter_interconnect_domain_gateway" "with_leaf_1" {
  for_each     = data.apstra_datacenter_interconnect_domain_gateways.with_leaf_1.ids
  blueprint_id = "eb5bf749-3610-4c23-ade1-50a6cc200abf"
  id           = each.value
}

# Surface the list of Interconnect Domain Gateway names available outside
# of this module
output "interconnect_domain_gateway_names" {
  value = sort([
    for gw in data.apstra_datacenter_interconnect_domain_gateway.with_leaf_1
    :
    gw.name
  ])
}

# Output:
#
#  interconnect_domain_gateway_names = [
#    "another_example",
#    "example",
#  ]
