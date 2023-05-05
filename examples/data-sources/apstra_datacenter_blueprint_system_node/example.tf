# This example determines the hostname of the lowest numbered system using
# apstra_datacenter_blueprint_system_nodes data source with a filter to select
# system ID 1.
#
# It then uses the returned ID to do a second lookup to get the full details
# of that system and assign the hostname to the system_one_hostname local
# variable

locals {
  blueprint_id = "abc-123"
}

data "apstra_datacenter_blueprint_system_nodes" "system_one" {
  blueprint_id = local.blueprint_id
  filters = {
    system_index = 1
  }
}

data "apstra_datacenter_blueprint_system_node" "system_one" {
  blueprint_id = local.blueprint_id
  id           = one(data.apstra_datacenter_blueprint_system_nodes.system_one.ids)
}

locals {
  system_one_hostname = data.apstra_datacenter_blueprint_system_node.system_one.hostname
}
