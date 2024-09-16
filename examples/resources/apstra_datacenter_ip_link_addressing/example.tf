# This example configures link numbering on a logical link between a leaf
# switch and a generic system. The link identified by "link_id" was created
# as a side-effect of attaching a Connectivity Template containing one or
# more IP Link primitives.
resource "apstra_datacenter_ip_link_addressing" "x" {
  blueprint_id = "22044be2-e7af-462d-847a-ce6d0b49000e"
  link_id      = "sz:HuZK45zx7V15D4qrjz0,vlan:22,a_001_leaf1<->a(link-000000001)[1]"

  switch_ipv4_address_type = "numbered"        # none | numbered
  switch_ipv4_address      = "192.0.2.0/31"
  switch_ipv6_address_type = "numbered"        # none | link_local | numbered
  switch_ipv6_address      = "2001:db8::1/127"

  generic_ipv4_address_type = "numbered"       # none | numbered
  generic_ipv4_address      = "192.0.2.1/31"
  generic_ipv6_address_type = "numbered"       # none | link_local | numbered
  generic_ipv6_address      = "2001:db8::2/127"
}

############################################################################

# Here is a more complete example which demonstrates all of the steps to
# assign an address to an IP Link handoff, beginning with creating a new
# Blueprint from an existing Template:

# Create a Blueprint using an existing template
resource "apstra_datacenter_blueprint" "example" {
  name        = "example blueprint"
  template_id = "L2_Virtual_EVPN"
}

# Assign an interface map to the first leaf switch
resource "apstra_datacenter_device_allocation" "leaf_1" {
  blueprint_id             = apstra_datacenter_blueprint.example.id
  node_name                = "l2_virtual_001_leaf1"
  initial_interface_map_id = "Juniper_vQFX__AOS-7x10-Leaf"
}

# Create a Generic System
resource "apstra_datacenter_generic_system" "example" {
  blueprint_id = apstra_datacenter_blueprint.example.id
  links = [
    {
      tags                          = ["L3 handoff"]
      target_switch_id              = apstra_datacenter_device_allocation.leaf_1.node_id
      target_switch_if_name         = "xe-0/0/2"
      target_switch_if_transform_id = 1
    }
  ]
}

# Create a Routing Zone
resource "apstra_datacenter_routing_zone" "example" {
  name         = "example"
  blueprint_id = apstra_datacenter_blueprint.example.id
}

# Create an IP Link connectivity template
resource "apstra_datacenter_connectivity_template_interface" "example" {
  name         = "example CT"
  blueprint_id = apstra_datacenter_blueprint.example.id
  ip_links = [
    {
      name                 = "IP Link example"
      routing_zone_id      = apstra_datacenter_routing_zone.example.id
      vlan_id              = 55
      ipv4_addressing_type = "numbered"
      ipv6_addressing_type = "none"
    },
  ]
}

# Discover the switch interface IDs used by the generic system
data "apstra_datacenter_interfaces_by_link_tag" "l3_handoff" {
  blueprint_id = apstra_datacenter_blueprint.example.id
  tags         = ["L3 handoff"]
  depends_on   = [apstra_datacenter_generic_system.example]
}

# Assign the CT to the Generic System
resource "apstra_datacenter_connectivity_templates_assignment" "example" {
  blueprint_id              = apstra_datacenter_blueprint.example.id
  connectivity_template_ids = [apstra_datacenter_connectivity_template_interface.example.id]
  application_point_id      = one(data.apstra_datacenter_interfaces_by_link_tag.l3_handoff.ids)
  fetch_ip_link_ids         = true
}

# Assign an IP address to the server link
resource "apstra_datacenter_ip_link_addressing" "example" {
  blueprint_id              = apstra_datacenter_blueprint.example.id
  link_id                   = apstra_datacenter_connectivity_templates_assignment.example.ip_link_ids[apstra_datacenter_connectivity_template_interface.example.id][55]
  switch_ipv4_address_type  = "numbered"
  switch_ipv4_address       = "192.0.2.0/31"
  generic_ipv4_address_type = "numbered"
  generic_ipv4_address      = "192.0.2.1/31"
}
