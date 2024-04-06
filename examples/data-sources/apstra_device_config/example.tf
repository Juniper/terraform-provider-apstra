# The following example shows how to access a configuration
# from a device in Apstra

# learn the details of leaf1
data "apstra_datacenter_system" "leaf1" {
  blueprint_id = var.blueprint_id
  name = "leaf_1"
}

# Collect the device config info
data "apstra_device_config" "device_a" {
  system_id = data.apstra_datacenter_system.leaf1.attributes.system_id
}
