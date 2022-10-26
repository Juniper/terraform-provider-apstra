# The following example shows how a module might accept a logical device name as an input variable and use it to retrieve the agent profile ID when provisioning a rack type.

variable "logical_device_name" {}

data "apstra_logical_device" "selected" {
    name = var.logical_device_name
}

resource "apstra_rack_type" "my_rack" {
    todo = "all of this"
}