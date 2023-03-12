# This example outputs a report of the name and count of available VNIs
# in each "in_use" VNI pool, as a map keyed by VNI pool ID.

# First collect all VNI pool IDs
data "apstra_vni_pools" "all" {}

# Grab all details for each VNI pool by looping over IDs
data "apstra_vni_pool" "each" {
  for_each = data.apstra_vni_pools.all.ids
  id = each.key
}

# Loop over pools, filter out pools other than "in_use". Calculate
# free space, output a map of "name" and "free", keyed by ID.
output "available_IDs_of_in_use_pools" {
  value = { for p in data.apstra_vni_pool.each : p.id => {
    name = p.name
    free = p.total - p.used
  } if p.status == "in_use" }
}