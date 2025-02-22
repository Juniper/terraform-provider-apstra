# The following example creates a Connectivity Template compatible with
# "interface" application points.
#
# It includes "Virtual Network (Multiple)" primitives which each represent
# a different Routing Zone and include all Virtual Networks from that
# Routing Zone.

locals { blueprint_id = "82a4dde9-eb98-4666-a010-d82f66296be4" }

# First, find the IDs of the interesting routing zones using the
# apstra_datacenter_routing_zones data source with a couple of filters:
data "apstra_datacenter_routing_zones" "selected" {
  blueprint_id = local.blueprint_id
  filters = [
    {
      name = "dev"
    },
    {
      name = "test"
    },
  ]
}

# Next, look up the details of each Routing Zone so that we have all of
# its details available in a map.
data "apstra_datacenter_routing_zone" "selected" {
  for_each     = data.apstra_datacenter_routing_zones.selected.ids
  blueprint_id = local.blueprint_id
  id           = each.value
}

# Loop over the Routing Zones. For each RZ, discover the IDs of all
# associated Virtual Networks.
data "apstra_datacenter_virtual_networks" "selected" {
  for_each     = data.apstra_datacenter_routing_zone.selected
  blueprint_id = local.blueprint_id
  filters = [
    {
      routing_zone_id = each.value.id
    }
  ]
}

# Finally, create an 'interface' type Connectivity Template with a
# "Virtual Network (Multiple)" primitive for each Routing Zone. Each
# of these primitives lists all of the Virtual Networks in
# that Routing Zone.
resource "apstra_datacenter_connectivity_template_interface" "example" {
  blueprint_id = local.blueprint_id
  name         = "Tagged handoff to all VNs from multiple RZs"
  description  = format("All VNs from the following RZs: \n - %s", join("\n - ", [for rz in data.apstra_datacenter_routing_zone.selected : rz.name]))
  virtual_network_multiples = {
    for rz in data.apstra_datacenter_routing_zone.selected : format("rz '%s' networks", rz.name) => {
      tagged_vn_ids = data.apstra_datacenter_virtual_networks.selected[rz.id].ids
    }
  }
}
