# This example pulls one predefined iba probe from a blueprint

data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

data "apstra_iba_predefined_probe" "p1" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "spine_superspine_interface_flapping"
}

output "o3" {
  value = data.apstra_iba_predefined_probe.p1
}


#o3 = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = <<-EOT
#  Generate a probe to determine if spine to superspine
#          interfaces are flapping.
#
#          A given interface is considered to be flapping if it transitions state
#          more than "Threshold" times over the last "Duration".  Such
#          flapping will cause an anomaly to be raised.
#
#          If more-than "Max Flapping Interfaces Percentage" percent of
#          interfaces on a given device are flapping, an anomaly will
#          be raised for that device.
#
#          Finally, the last "Anomaly History Count" anomaly state-changes
#          are stored for observation.
#
#  EOT
#  "name" = "spine_superspine_interface_flapping"
#  "schema" = "{\"type\": \"object\", \"properties\": {\"label\": {\"title\": \"Probe Label\", \"type\": \"string\", \"minLength\": 1, \"maxLength\": 120, \"default\": \"Interface Flapping (Spine to Superspine Interfaces)\"}, \"threshold\": {\"title\": \"Threshold\", \"description\": \"Sum total of number of flaps in recent-history for which an anomaly will be raised. The larger the value, the more memory the probe will consume.\", \"type\": \"integer\", \"minimum\": 1, \"default\": 5}, \"max_flapping_interfaces_percentage\": {\"title\": \"Max Flapping Interfaces Percentage\", \"description\": \"Maximum percentage of flapping interfaces on a device\", \"type\": \"integer\", \"minimum\": 0, \"maximum\": 100, \"default\": 10}, \"duration\": {\"title\": \"Duration\", \"description\": \"Duration of recent-history in which interface flapping will be considered. The longer the period, the more memory the probe will consume.\", \"type\": \"integer\", \"minimum\": 1, \"default\": 60}}}"
#}
