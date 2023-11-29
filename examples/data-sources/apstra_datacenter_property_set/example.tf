# This example pulls one property set from a blueprint
data "apstra_datacenter_blueprint" "b" {
  name = "test"
}
data "apstra_datacenter_property_set" "d" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "NTP Server"
}
output "o" {
  value = data.apstra_datacenter_property_set.d
}

#Output looks like this
#o = {
#  "blueprint_id" = "d6c74373-45ce-4d88-9547-ac23c2ebe61e"
#  "data" = "{\"ntp_server\": \"172.20.12.4\"}"
#  "id" = "8e1c1316-857d-4a49-a10d-5a61063bc38c"
#  "keys" = toset([
#    "ntp_server",
#  ])
#  "name" = "NTP Server"
#  "stale" = false
#  "sync_required" = tobool(null)
#  "sync_with_catalog" = tobool(null)
#}
