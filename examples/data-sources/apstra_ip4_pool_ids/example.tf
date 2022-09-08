# The following example shows outputting all IPv4 pool IDs.

data "apstra_ip4_pool_ids" "all" {}

output ip4_pool_ids {
   value = data.apstra_ip4_pool_ids.all.ids
}
