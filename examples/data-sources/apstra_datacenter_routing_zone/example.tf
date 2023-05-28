# This example pulls the details of a routing zone using
# manually-entered blueprint and routing zone IDs.
#
# In a production scenario it's more likely that these
# values would come from a lookup by name or from the
# read-only attributes of the resources which created
# them.

data "apstra_datacenter_routing_zone" "test" {
  blueprint_id = "13b8ddb4-f230-4727-ab0f-aa829551a129"
  id           = "b03V9dIBBqeQkcV69nY"
}

output "routing_zone" {
  value = data.apstra_datacenter_routing_zone.test
}

# The output looks like this:

# routing_zone = {
#   "blueprint_id" = "13b8ddb4-f230-4727-ab0f-aa829551a129"
#   "dhcp_servers" = toset([
#     "192.168.10.100",
#     "192.168.20.100",
#   ])
#   "id" = "b03V9dIBBqeQkcV69nY"
#   "name" = "bar"
#   "routing_policy_id" = "6QVA7utPgvPDWe0-vLs"
#   "vlan_id" = 43
#   "vni" = 10043
#

