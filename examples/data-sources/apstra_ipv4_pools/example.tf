# The following example shows outputting a report of free space across all IPv4
# resource pools using `apstra_ip4_pools` (to collect pool IDs) and
# `apstra_ip4_pool` to query for pool details

# List all pool IDs
data "apstra_ipv4_pools" "all" {}

# Loop over pool IDs, creating an instance of `apstra_ipv4_pool` for each.
data "apstra_ipv4_pool" "each" {
   for_each = toset(data.apstra_ipv4_pools.all.ids)
   id = each.value
}

# Output the name and free space of each pool
output "ipv4_pool_report" {
  value = {for k, v in data.apstra_ipv4_pool.each : k => {
    name = v.name
    free = v.total - v.used
  }}
}

################################################################################
# The output object above will produce something like the following:
#
#   ipv4_pool_report = {
#     "091a4c18-0911-4f78-8db3-c1033b88e08f" = {
#       "free" = 1008
#       "name" = "leaf-loopback"
#     }
#     "472b87f2-6174-448c-b3d1-b57c2b48ab58" = {
#       "free" = 1022
#       "name" = "spine-loopback"
#     }
#     "Private-10_0_0_0-8" = {
#       "free" = 16777216
#       "name" = "Private-10.0.0.0/8"
#     }
#     "Private-172_16_0_0-12" = {
#       "free" = 1048576
#       "name" = "Private-172.16.0.0/12"
#     }
#     "Private-192_168_0_0-16" = {
#       "free" = 65536
#       "name" = "Private-192.168.0.0/16"
#     }
#     "cc023b60-9941-40a5-a07c-2b15920e544f" = {
#       "free" = 992
#       "name" = "spine-leaf"
#     }
#   }
#
################################################################################
