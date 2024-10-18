# Device configuraiton details can be fetched within within a blueprint
# either by system node ID (regardless of whether a switch is assigned)
# or by assigned system ID (device key / serial number / MAC address):

// look up the details of spine 1
data "apstra_datacenter_system" "example" {
  blueprint_id = local.blueprint_id
  name         = "spine1"
}

# retrieve the config details
data "apstra_blueprint_device_rendered_config" "example" {
  blueprint_id = local.blueprint_id
  node_id      = data.apstra_datacenter_system.example.id
  # system_id    = "525400E365A5" // specify either the system graph node ID or the assigned switch serial number
}

output "rendered_config" { value = data.apstra_blueprint_device_rendered_config.example }

# The output looks like this:
#
# rendered_config = {
#   "blueprint_id" = "d3336749-88c9-4922-8f61-043198664840"
#   "deployed_config" = <<-EOT
#   system {
#       host-name spine1;
#   }
#   interfaces {
#       replace: xe-0/0/0 {
#           description "facing_l2-virtual-001-leaf1:xe-0/0/0";
#           mtu 9216;
#   <<<<<<deployed_config trimmed for brevity>>>>>>
#   EOT
#   "incremental_config" = <<-EOT
#
#   [routing-options]
#   - autonomous-system 64512;
#   + autonomous-system 64511;
#
#   EOT
#   "node_id" = "GCmJI_Tl47TkOls2vg"
#   "staged_config" = <<-EOT
#   system {
#       host-name spine1;
#   }
#   interfaces {
#       replace: xe-0/0/0 {
#           description "facing_l2-virtual-001-leaf1:xe-0/0/0";
#           mtu 9216;
#   <<<<<<staged_config trimmed for breevity>>>>>>
#   EOT
#   "system_id" = tostring(null)
# }
#
