# The following example shows outputting a report of free space across all
# ASN resource pools:

data "apstra_asn_pools" "all" {}

data "apstra_asn_pool" "all" {
  for_each = toset(data.apstra_asn_pools.all.ids)
  id = each.value
}

output "asn_report" {
  value = {for k, v in data.apstra_asn_pool.all : k => {
    name = v.name
    free = v.total - v.used
  }}
}

############################################################################
# The output object above will produce something like the following:
#
#   asn_report = {
#     "3ddb7a6a-4c84-458f-8632-705764c0f6ca" = {
#       "free" = 100
#       "name" = "spine"
#     }
#     "Private-4200000000-4294967294" = {
#       "free" = 94967293
#       "name" = "Private-4200000000-4294967294"
#     }
#     "Private-64512-65534" = {
#       "free" = 1020
#       "name" = "Private-64512-65534"
#     }
#     "dd0d3b45-2020-4382-9c01-c43e7d474546" = {
#       "free" = 10002
#       "name" = "leaf"
#     }
#   }
############################################################################
