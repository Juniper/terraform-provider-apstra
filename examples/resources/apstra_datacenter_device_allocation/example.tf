# This example deploys a blueprint from a template and assigns
# interface maps to the fabric nodes.

# Instantiate the blueprint from the template
resource "apstra_datacenter_blueprint" "r" {
  name        = "example blueprint with switch allocation"
  template_id = "L2_Virtual_EVPN"
}

# Only one of "device_key" and "initial_interface_map_id" is required in
# most cases.
#
# When only the initial_interface_map_id is provided, we set the interface
# map (this step eliminates build errors), but do not assign systems by ID.
#
# When only the device_key is provided, we try to infer the
# initial_interface_map_id by first determining the device type, and then
# looking for candidate interface maps which can map the specific hardware
# to the logical device specified by the fabric role. When multiple
# candidate interface maps exist supplying interface_map_id becomes
# mandatory.
locals {
  switches = {
    spine1 = {
      #     device_key       = "<serial-number-goes-here>"
      initial_interface_map_id = "Juniper_vQFX__AOS-7x10-Spine"
    }
    spine2 = {
      #     device_key       = "<serial-number-goes-here>"
      initial_interface_map_id = "Juniper_vQFX__AOS-7x10-Spine"
    }
    l2_virtual_001_leaf1 = {
      #     device_key       = "<serial-number-goes-here>"
      initial_interface_map_id = "Juniper_vQFX__AOS-7x10-Leaf"
    }
    l2_virtual_002_leaf1 = {
      #     device_key       = "<serial-number-goes-here>"
      initial_interface_map_id = "Juniper_vQFX__AOS-7x10-Leaf"
    }
    l2_virtual_003_leaf1 = {
      #     device_key       = "<serial-number-goes-here>"
      initial_interface_map_id = "Juniper_vQFX__AOS-7x10-Leaf"
    }
    l2_virtual_004_leaf1 = {
      #     device_key       = "<serial-number-goes-here>"
      initial_interface_map_id = "Juniper_vQFX__AOS-7x10-Leaf"
    }
  }
}

# Assign switches to fabric roles
resource "apstra_datacenter_device_allocation" "r" {
  for_each                 = local.switches
  blueprint_id             = apstra_datacenter_blueprint.r.id
  initial_interface_map_id = each.value["initial_interface_map_id"]
  node_name                = each.key
  deploy_mode              = "deploy"
}
