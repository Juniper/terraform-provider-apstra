# This example creates a Rack Based Template using a variety
# of pre-configured Rack Types.

# When creating the template we'll loop over this map of
# Rack Type IDs and the desired count of each.
locals {
  rack_id_and_count = {
    access_switch_4x = 3
    L2_Compute       = 4
    two_leafs        = 2
  }
}

# Instantiate a template using the Rack Type IDs
# and counts defined above.
resource "apstra_template_rack_based" "r" {
  name                     = "terraform example"
  asn_allocation_scheme    = "single"
  overlay_control_protocol = "evpn"
  spine = {
    logical_device_id = "AOS-16x40-1"
    count             = 2
  }
  rack_infos = {
    for id, count in local.rack_id_and_count : id => { count = count }
  }
}
