# The following example shows how a module might accept a logical device name
# as an input variable and use it to retrieve the agent profile ID when
# provisioning a rack type.

variable "logical_device_name" {
    type = string
    default = "AOS-7x10-Leaf"
}

data "apstra_logical_device" "selected" {
    name = var.logical_device_name
}

resource "apstra_rack_type" "my_rack" {
    name = "terraform, yo"
    fabric_connectivity_design = "l3clos"
    leaf_switches = {
        leaf_one = {
            spine_link_count = 1
            spine_link_speed = "10G"
            logical_device_id = data.apstra_logical_device.selected.id
        }
    }
    access_switches = {
        access_one = {
            count = 1
            logical_device_id = data.apstra_logical_device.selected.id
            links = {
                link_one = {
                    speed = "10G"
                    links_per_switch = 2
                    target_switch_name = "leaf_one"
                }
            }
        }
    }
}