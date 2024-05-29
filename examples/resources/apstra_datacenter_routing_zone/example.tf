# This example creates a routing zone within an
# existing datacenter blueprint.
resource "apstra_datacenter_routing_zone" "blue" {
  name         = "vrf blue"
  blueprint_id = "<blueprint-id-goes-here>"
  vlan_id      = 5                                    # optional
  vni          = 5000                                 # optional
  dhcp_servers = ["192.168.100.10", "192.168.200.10"] # optional
  #  routing_policy_id = "<routing-policy-node-id-goes-here>" # optional
}

# Next, assign an IPv4 pool to be used by loopback interfaces of leaf
# switches participating in the Routing Zone.
resource "apstra_datacenter_resource_pool_allocation" "blue_loopbacks" {
  blueprint_id    = "<blueprint-id-goes-here>"
  routing_zone_id = apstra_datacenter_routing_zone.blue.id
  pool_ids        = ["<ipv4-pool-id-goes-here>"]
  role            = "leaf_loopback_ips"
}
