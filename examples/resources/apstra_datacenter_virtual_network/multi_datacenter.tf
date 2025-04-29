# Example: Creating identical virtual networks across different datacenter blueprints
# with non-overlapping SVI IPs

# DC1 Virtual Network
resource "apstra_datacenter_virtual_network" "dc1_vn" {
  name                       = "multi-dc-vn"
  blueprint_id               = apstra_datacenter_blueprint.dc1.id
  type                       = "vxlan"
  routing_zone_id            = apstra_datacenter_routing_zone.dc1_zone.id
  ipv4_connectivity_enabled  = true
  ipv4_subnet                = "10.0.0.0/24"
  ipv4_virtual_gateway_enabled = true
  ipv4_virtual_gateway       = "10.0.0.1"
  
  # DC1 SVI IPs - using 10.0.0.2-3 range
  svi_ips {
    system_id     = apstra_datacenter_device_allocation.dc1_leaf1.system_id
    ipv4_address  = "10.0.0.2/24"
    ipv4_mode     = "enabled"
  }
  
  svi_ips {
    system_id     = apstra_datacenter_device_allocation.dc1_leaf2.system_id
    ipv4_address  = "10.0.0.3/24"
    ipv4_mode     = "enabled"
  }

  # DC1 bindings
  bindings = {
    "${apstra_datacenter_device_allocation.dc1_leaf1.system_id}" = {
      access_ids = []
      vlan_id    = 100
    }
    "${apstra_datacenter_device_allocation.dc1_leaf2.system_id}" = {
      access_ids = []
      vlan_id    = 100
    }
  }
}

# DC2 Virtual Network - Identical VN configuration but with different SVI IPs
resource "apstra_datacenter_virtual_network" "dc2_vn" {
  name                       = "multi-dc-vn"
  blueprint_id               = apstra_datacenter_blueprint.dc2.id
  type                       = "vxlan"
  routing_zone_id            = apstra_datacenter_routing_zone.dc2_zone.id
  ipv4_connectivity_enabled  = true
  ipv4_subnet                = "10.0.0.0/24"
  ipv4_virtual_gateway_enabled = true
  ipv4_virtual_gateway       = "10.0.0.1"
  
  # DC2 SVI IPs - using 10.0.0.10-11 range to avoid overlap with DC1
  svi_ips {
    system_id     = apstra_datacenter_device_allocation.dc2_leaf1.system_id
    ipv4_address  = "10.0.0.10/24"
    ipv4_mode     = "enabled"
  }
  
  svi_ips {
    system_id     = apstra_datacenter_device_allocation.dc2_leaf2.system_id
    ipv4_address  = "10.0.0.11/24"
    ipv4_mode     = "enabled"
  }

  # DC2 bindings
  bindings = {
    "${apstra_datacenter_device_allocation.dc2_leaf1.system_id}" = {
      access_ids = []
      vlan_id    = 100
    }
    "${apstra_datacenter_device_allocation.dc2_leaf2.system_id}" = {
      access_ids = []
      vlan_id    = 100
    }
  }
}