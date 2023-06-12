# This example pulls the details of all property sets
# that have been imported into a blueprint

data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

data "apstra_datacenter_property_sets" "ps" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
}
output "o" {
  value = data.apstra_datacenter_property_sets.ps
}

#Output looks something like this
#o = {
#  "blueprint_id" = "d6c74373-45ce-4d88-9547-ac23c2ebe61e"
#  "property_sets" = toset([
#    {
#      "blueprint_id" = "d6c74373-45ce-4d88-9547-ac23c2ebe61e"
#      "data" = "{\"junos_mgmt_vrf\": \"mgmt_junos\", \"mgmt_vrf\": \"management\"}"
#      "id" = "3ae45f2e-c9ed-401b-8f00-367fb9a5e0e8"
#      "keys" = toset([
#        "junos_mgmt_vrf",
#        "mgmt_vrf",
#      ])
#      "name" = "MGMT VRF"
#      "stale" = false
#    },
#    {
#      "blueprint_id" = "d6c74373-45ce-4d88-9547-ac23c2ebe61e"
#      "data" = "{\"ntp_server\": \"172.20.12.4\"}"
#      "id" = "8e1c1316-857d-4a49-a10d-5a61063bc38c"
#      "keys" = toset([
#        "ntp_server",
#      ])
#      "name" = "NTP Server"
#      "stale" = false
#    },
#  ])
#}
