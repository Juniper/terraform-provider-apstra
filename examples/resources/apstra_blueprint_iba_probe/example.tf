# The following example instantiates predefined probes in Apstra

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

output "o"{
  value = apstra_blueprint_iba_probe.p_device_health
}

resource "apstra_blueprint_iba_probe" "p_device_traffic" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  probe_json = file("device_traffic.json")
}

output "o2"{
  value = apstra_blueprint_iba_probe.p_device_health
}



#Output Looks something like this
#o = {
#"blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#"description" = "This probe alerts if the system health parameters (CPU, memory and disk usage) exceed their specified thresholds for the specified duration."
#"id" = "0b738068-11dc-4050-aa7c-65cd8715cc6e"
#"name" = "Device System Health"
#"predefined_probe_id" = "device_health"
#"probe_config" = "{\"duration\":660,\"history_duration\":604800,\"max_cpu_utilization\":80,\"max_disk_utilization\":80,\"max_memory_utilization\":80,\"threshold_duration\":360}"
#"stages" = toset([
#"Check cpu utilization threshold",
#"Check disk utilization threshold",
#"Check memory utilization threshold",
#"Disk utilization data",
#"Disk utilization data per partition",
#"System cpu utilization data",
#"System memory utilization data",
#"Systems with high CPU utilization",
#"Systems with high disk utilization",
#"Systems with high memory utilization",
#"sustained_high_cpu_utilization",
#"sustained_high_disk_utilization",
#"sustained_high_memory_utilization",
#])
#}

