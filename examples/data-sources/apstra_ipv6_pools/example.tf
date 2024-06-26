# The following example shows outputting a report of free space across all
# IPv6 resource pools using `apstra_ip4_pools` (to collect pool IDs) and
# `apstra_ip4_pool` to query for pool details

# List all pool IDs
data "apstra_ipv6_pools" "all" {}

# Loop over pool IDs, creating an instance of `apstra_ipv6_pool` for each.
data "apstra_ipv6_pool" "each" {
   for_each = toset(data.apstra_ipv6_pools.all.ids)
   id = each.value
}

# Output the name and free space of each pool
output "ipv6_pool_report" {
  value = {for k, v in data.apstra_ipv6_pool.each : k => {
    name = v.name
    free = v.total - v.used
  }}
}

############################################################################
# The output object above will produce something like the following:
#
#   ipv6_pool_report = {
#     "Private-fc01-a05-fab-48" = {
#       "free" = 1208925819614629200000000
#       "name" = "Private-fc01:a05:fab::/48"
#     }
#   }
#
############################################################################
