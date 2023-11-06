# This example collects the IDs of external gateways configured on all leaf
# switches, and external gateways configured on spine switches which share
# all route types.

// find leaf switch IDs
data "apstra_datacenter_systems" "leaf_switches" {
  blueprint_id = "007723b7-a387-4bb3-8a5e-b5e9f265de0d"
  filters = [
    { system_type = "switch", role = "leaf" },
  ]
}

// find spine switch IDs
data "apstra_datacenter_systems" "spine_switches" {
  blueprint_id = "007723b7-a387-4bb3-8a5e-b5e9f265de0d"
  filters = [
    { system_type = "switch", role = "spine" },
  ]
}

// find interesting external gateway IDs
data "apstra_datacenter_external_gateways" "leaf_peers_and_spine_peers_with_all_routes" {
  blueprint_id = "007723b7-a387-4bb3-8a5e-b5e9f265de0d"
  filters = [
    {
      local_gateway_nodes = data.apstra_datacenter_systems.leaf_switches.ids
    },
    {
      local_gateway_nodes = data.apstra_datacenter_systems.spine_switches.ids
      evpn_route_types    = "all"
    },
  ]
}
