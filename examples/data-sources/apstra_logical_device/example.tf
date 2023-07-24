# The following example shows how a module might accept a Logical Device's
# name as an input variable and then use it to retrieve the Logical Device
# ID when provisioning a Rack Type.

# module input variable has the Logical Device name
variable "logical_device_name" {
    type = string
    default = "AOS-24x10-2"
}

# Data lookup using the Logical Device name. Name collisions are possible,
# will produce an error. Apstra permits name collisions for many object
# types. It's probably best to avoid creating them.
data "apstra_logical_device" "selected" {
    name = var.logical_device_name
}

# Create a Rack Type using the Logical Device ID found in the object which
# was looked up by name.
resource "apstra_rack_type" "my_rack" {
    name = "terraform, yo"
    fabric_connectivity_design = "l3clos"
    leaf_switches = {
        leaf_one = {
            spine_link_count = 2
            spine_link_speed = "10G"
            logical_device_id = data.apstra_logical_device.selected.id
            redundancy_protocol = "esi"
        }
    }
    access_switches = {
        access_one = {
            count = 3
            logical_device_id = data.apstra_logical_device.selected.id
            links = {
                link_one = {
                    speed = "10G"
                    links_per_switch = 2
                    target_switch_name = "leaf_one"
                }
            }
            esi_lag_info = {
                l3_peer_link_count = 2
                l3_peer_link_speed = "10G"
            }
        }
    }
}
