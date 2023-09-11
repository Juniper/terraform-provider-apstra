# This example determines the hostname of the lowest numbered system using
# apstra_datacenter_system data source with a filter to select
# system ID 1.
#
# It then uses the returned ID to do a second lookup to get the full details
# of that system and assign the hostname to the system_one_hostname local
# variable

locals {
  blueprint_id = "abc-123"
}

data "apstra_datacenter_systems" "system_one" {
  blueprint_id = local.blueprint_id
  filter = {
    system_index = 1
  }
}

data "apstra_datacenter_system" "system_one" {
  blueprint_id = local.blueprint_id
  id           = one(data.apstra_datacenter_systems.system_one.ids)
}

locals {
  system_one_hostname = data.apstra_datacenter_system.system_one.hostname
}
