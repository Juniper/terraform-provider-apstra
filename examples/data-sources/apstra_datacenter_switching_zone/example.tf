# This example pulls the details of a switching zone using
# manually-entered blueprint and switching zone IDs.
#
# In a production scenario it's more likely that these
# values would come from a lookup by name or from the
# read-only attributes of the resources which created
# them.

data "apstra_datacenter_switching_zone" "by_id" {
  blueprint_id = "f526f3dd-d814-4a1c-bd4b-950eabf82d13"
  id           = "OHOsoQ4zk-oKrAYaPA"
}

output "apstra_datacenter_switching_zone_by_id" {
  value = data.apstra_datacenter_switching_zone.by_id
}

# The output looks like this:

# apstra_datacenter_switching_zone_by_id = {
#   "blueprint_id" = "f526f3dd-d814-4a1c-bd4b-950eabf82d13"
#   "description" = "description"
#   "id" = "OHOsoQ4zk-oKrAYaPA"
#   "mac_vrf_name" = "mac_vrf_name"
#   "name" = "name"
#   "route_target" = "100:10001"
#   "service_type" = "vlan_bundle"
#   "tags" = toset([
#     "bar",
#     "baz",
#     "foo",
#   ])
# }
