# This example outputs a report of the name and count of available Integers
# in each "not_in_use" Integer pool, as a map keyed by Integer pool ID.

# First collect all Integer pool IDs
data "apstra_integer_pools" "all" {}

# Grab all details for each Integer pool by looping over values
data "apstra_integer_pool" "each" {
  for_each = data.apstra_integer_pools.all.ids
  id = each.key
}

# Loop over pools, filter out pools other than "not_in_use". Calculate
# free space, output a map of "name" and "free", keyed by ID.
output "available_IDs_of_unused_pools" {
  value = { for p in data.apstra_integer_pool.each : p.id => {
    name = p.name
    free = p.total - p.used
  } if p.status == "not_in_use" }
}
