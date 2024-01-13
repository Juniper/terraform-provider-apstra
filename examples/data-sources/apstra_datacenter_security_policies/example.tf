# find the IDs of disabled security policies which would apply to traffic
# inbound from Internet to objects tagged with "dev" or "test"

# first, find the ID of the "internet" routing zone
data "apstra_datacenter_routing_zone" "internet" {
  blueprint_id = "a52fb4ff-b352-46a3-9141-820b40972133"
  name         = "internet"
}

# we use two filters: one matching objects tagged "dev" and one matching "test"
data "apstra_datacenter_security_policies" "nonprod_disabled_policies" {
  blueprint_id = "a52fb4ff-b352-46a3-9141-820b40972133"
  filters = [
    {
      enabled                     = false
      tags                        = ["dev"]
      source_application_point_id = data.apstra_datacenter_routing_zone.internet.id
    },
    {
      enabled                     = false
      tags                        = ["test"]
      source_application_point_id = data.apstra_datacenter_routing_zone.internet.id
    },
  ]
}

# the output is available by reference:
#   data.apstra_datacenter_security_policies.nonprod_disabled_policies.ids