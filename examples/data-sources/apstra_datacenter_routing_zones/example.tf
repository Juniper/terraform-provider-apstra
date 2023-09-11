# Without specifying no filter, a wide search is performed.
# All routing zones in the blueprint will match.
data "apstra_datacenter_routing_zones" "all" {
  blueprint_id = "05f9d3fc-671a-4efc-8e91-5ef87b2937d3"
}

# This example performs a very narrow search. Only one (or zero!)
# routing zones can match the resulting query.
data "apstra_datacenter_routing_zones" "rzs" {
  blueprint_id = "05f9d3fc-671a-4efc-8e91-5ef87b2937d3"
  filter = { # all filter attributes are optional
    name              = "customer_1"
    vlan_id           = 55
    vni               = 10055
    dhcp_servers      = ["192.168.5.100", "192.168.10.100"]
    routing_policy_id = "vqsv3F93MBHUgg5e8ws"
  }
}
