resource "apstra_rack_type" "example" {
  name                       = "example rack type"
  description                = "Created by Terraform"
  fabric_connectivity_design = "l3clos"
  leaf_switches = { // leaf switches are a map keyed by switch name, so
    leaf_switch = { // "leaf switch" on this line is the name used by links targeting this switch.
      logical_device_id   = "AOS-24x10-2"
      spine_link_count    = 1
      spine_link_speed    = "10G"
      redundancy_protocol = "esi"
    }
  }
  access_switches = { // access switches are a map keyed by switch name, so
    access_switch = { // "access_switch" on this line is the name used by links targeting this switch.
      logical_device_id = "AOS-24x10-2"
      count             = 1
      esi_lag_info = {
        l3_peer_link_count = 1
        l3_peer_link_speed = "10G"
      }
      links = {
        leaf_switch = {
          speed              = "10G"
          target_switch_name = "leaf_switch" // note "leaf_switch" corresponds to a map key above.
          links_per_switch   = 1
        }
      }
    }
  }
  generic_systems = {
    webserver = {
      count             = 2
      logical_device_id = "AOS-4x10-1"
      links = {
        link = {
          speed              = "10G"
          target_switch_name = "access_switch" // note "access_switch" corresponds to a map key above.
          lag_mode           = "lacp_active"
          switch_peer        = "first"
        }
      }
    }
  }
}
