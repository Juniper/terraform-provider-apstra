# The following example shows outputting a report of all the
# Apstra Property Sets.

# List all Apstra Property Set Ids
data "apstra_property_sets" "all" {}

# Loop over Property Set IDs, creating an instance of `apstra_property_set`
# for each.
data "apstra_property_set" "each_ps" {
  for_each = toset(data.apstra_property_sets.all.ids)
    id = each.value
}

# Output the property set report
output "apstra_property_set_report" {
  value = {for k, v in data.apstra_property_set.each_ps : k => {
    name = v.name
    data = jsondecode(v.data)
    blueprints = v.blueprints
  }}
}

############################################################################
# The output object above will produce something like the following:
# apstra_property_set_report = {
#  "0a2f4768-7312-4512-b544-f17de1dc155a" = {
#    "blueprints" = toset([])
#    "data"       = {
#      "nameserver1" = "10.155.191.252"
#      "nameserver2" = "172.21.200.60"
#    },
#    "keys"       = toset([
#      "nameserver1",
#      "nameserver2",
#     ])
#    "name" = "nameservers"
#  }
#  "7d68daeb-b8f5-4512-9417-9e5812d87783" = {
#    "blueprints" = toset([
#      "11d06e08-0982-44d4-b57a-9fb04c40532c",
#      "must_bp_dc1",
#      "must_bp_dc2",
#    ])
#    "data" = {
#      "snmp_collector_01" = "10.6.1.87/32"
#      "snmp_collector_02" = "10.6.1.88/32"
#    }
#    "keys" = toset([
#      "snmp_collector_01",
#      "snmp_collector_02",
#    ])
#    "name" = "MUST_SNMP_D42"
#  }
# }
############################################################################
