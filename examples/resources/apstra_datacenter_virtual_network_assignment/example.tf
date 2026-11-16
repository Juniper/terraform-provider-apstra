# This resource binds a virtual network to a leaf switch, and configures the
# leaf switch address within that virtual network, in a datacenter
# blueprint.

resource "apstra_datacenter_virtual_network_assignment" "example" {
  blueprint_id       = "some-blueprint-id"
  virtual_network_id = "some-virtual-network-id"
  leaf_switch_id     = "some-leaf-switch-id"
  vlan = 100 # optional
  ipv4_mode          = "forced"       # Optional - One of: [enabled|disabled|forced]
  ipv4_address       = "192.0.2.2/24" # Optional - The anycast gateway address is usually at ".1"

  # Optional - in the case of MLAG or ESI leaf switches, these are access
  # switches attached to *either* leaf switch in the pair. Failure to list
  # desired access switches from both members of the pair will result in
  # state churn and incomplete configuration.
  access_switch_ids  = ["some-access-switch-id-1", "some-access-switch-id-2"]
}
