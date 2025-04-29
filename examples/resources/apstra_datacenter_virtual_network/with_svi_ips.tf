# Create a virtual network with explicit SVI IPs
resource "apstra_datacenter_virtual_network" "example" {
  name                       = "vn-with-svi-ips"
  blueprint_id               = apstra_datacenter_blueprint.example.id
  type                       = "vxlan"
  routing_zone_id            = apstra_datacenter_routing_zone.example.id
  ipv4_connectivity_enabled  = true
  ipv4_subnet                = "192.0.2.0/24"
  ipv4_virtual_gateway_enabled = true
  ipv4_virtual_gateway       = "192.0.2.1"
  
  # Specify SVI IPs for leaf switches
  svi_ips {
    system_id     = "leaf1-system-id"
    ipv4_address  = "192.0.2.2/24"
    ipv4_mode     = "enabled"
  }
  
  svi_ips {
    system_id     = "leaf2-system-id"
    ipv4_address  = "192.0.2.3/24"
    ipv4_mode     = "enabled" 
  }

  # Add bindings for the leaf switches
  bindings = {
    "leaf1-system-id" = {
      access_ids = []
      vlan_id    = 100
    }
    "leaf2-system-id" = {
      access_ids = []
      vlan_id    = 100
    }
  }
}