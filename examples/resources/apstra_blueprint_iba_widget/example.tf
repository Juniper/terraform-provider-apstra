# The following example instantiates a predefined probe and initiates a widget in Apstra

data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

resource "apstra_blueprint_iba_probe" "p_device_health" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  predefined_probe_id = "device_health"
  probe_config = jsonencode(
    {
      "max_cpu_utilization": 80,
      "max_memory_utilization": 80,
      "max_disk_utilization": 80,
      "duration": 660,
      "threshold_duration": 360,
      "history_duration": 604800
    }
  )
}

resource "apstra_blueprint_iba_widget" "w_device_health_high_cpu" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = "Devices with high CPU Utilization"
  probe_id = apstra_blueprint_iba_probe.p_device_health.id
  stage = "Systems with high CPU utilization"
  description = "made from terraform"
}

output "o"{
  value = apstra_blueprint_iba_widget.w_device_health_high_cpu
}

# Output looks something like this
#o = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = "made from terraform"
#  "id" = "39613d7f-a5be-4359-bbd8-497c56894153"
#  "name" = "Devices with high CPU Utilization"
#  "probe_id" = "0b738068-11dc-4050-aa7c-65cd8715cc6e"
#  "stage" = "Systems with high CPU utilization"
#}

