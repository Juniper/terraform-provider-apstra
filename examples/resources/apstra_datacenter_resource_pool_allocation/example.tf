# This example deploys a blueprint from a template and assigns
# resource pools maps to various roles in the fabric.

# Instantiate the blueprint from the template
resource "apstra_datacenter_blueprint" "r" {
  name        = "example blueprint with switch allocation"
  template_id = "L2_Virtual_EVPN"
}

# Discover every ASN resource pool ID
data "apstra_asn_pools" "all" {}

# ASN pools and IPv4 pools will be allocated from these
# local variables using looping resources.
locals {
  asn_pools = {
    // use the first discovered ASN pool for spines
    spine_asns = slice(tolist(data.apstra_asn_pools.all.ids), 0, 1)
    // use all other discovered ASN pools for leafs
    leaf_asns = slice(tolist(data.apstra_asn_pools.all.ids), 1, length(data.apstra_asn_pools.all.ids))
  }
  ipv4_pools = {
    spine_loopback_ips  = ["Private-10_0_0_0-8"]
    leaf_loopback_ips   = ["Private-10_0_0_0-8"]
    spine_leaf_link_ips = ["Private-10_0_0_0-8"]
  }
}

# Assign ASN pools to fabric roles
resource "apstra_datacenter_resource_pool_allocation" "asn" {
  for_each     = local.asn_pools
  blueprint_id = apstra_datacenter_blueprint.r.id
  role         = each.key
  pool_ids     = each.value
}

# Assign IPv4 pools to fabric roles
resource "apstra_datacenter_resource_pool_allocation" "ipv4" {
  for_each     = local.ipv4_pools
  blueprint_id = apstra_datacenter_blueprint.r.id
  role         = each.key
  pool_ids     = each.value
}
