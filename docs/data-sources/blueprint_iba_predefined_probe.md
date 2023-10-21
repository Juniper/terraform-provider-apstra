---
page_title: "apstra_blueprint_iba_predefined_probe Data Source - terraform-provider-apstra"
subcategory: ""
description: |-
  This data source provides details of a specific IBA Predefined Probe in a Blueprint.
---

# apstra_blueprint_iba_predefined_probe (Data Source)

This data source provides details of a specific IBA Predefined Probe in a Blueprint.

## Example Usage

```terraform
# This example pulls all the predefined iba probes from a blueprint

data "apstra_datacenter_blueprint" "b" {
  name = "test"
}

data "apstra_blueprint_iba_predefined_probes" "all" {
  blueprint_id = data.apstra_datacenter_blueprint.b.id
}


data "apstra_blueprint_iba_predefined_probe" "all" {
  for_each = data.apstra_blueprint_iba_predefined_probes.all.names
  blueprint_id = data.apstra_datacenter_blueprint.b.id
  name = each.key
}

output "o2" {
  value = data.apstra_blueprint_iba_predefined_probe.all
}

# Output looks something like this
#o2 = {
#"spine_superspine_interface_flapping" = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = <<-EOT
#    Generate a probe to determine if spine to superspine
#            interfaces are flapping.
#
#            A given interface is considered to be flapping if it transitions state
#            more than "Threshold" times over the last "Duration".  Such
#            flapping will cause an anomaly to be raised.
#
#            If more-than "Max Flapping Interfaces Percentage" percent of
#            interfaces on a given device are flapping, an anomaly will
#            be raised for that device.
#
#            Finally, the last "Anomaly History Count" anomaly state-changes
#            are stored for observation.
#
#    EOT
#  "name" = "spine_superspine_interface_flapping"
#  "schema" = "{\"type\": \"object\", \"properties\": {\"label\": {\"title\": \"Probe Label\", \"type\": \"string\", \"minLength\": 1, \"maxLength\": 120, \"default\": \"Interface Flapping (Spine to Superspine Interfaces)\"}, \"threshold\": {\"title\": \"Threshold\", \"description\": \"Sum total of number of flaps in recent-history for which an anomaly will be raised. The larger the value, the more memory the probe will consume.\", \"type\": \"integer\", \"minimum\": 1, \"default\": 5}, \"max_flapping_interfaces_percentage\": {\"title\": \"Max Flapping Interfaces Percentage\", \"description\": \"Maximum percentage of flapping interfaces on a device\", \"type\": \"integer\", \"minimum\": 0, \"maximum\": 100, \"default\": 10}, \"duration\": {\"title\": \"Duration\", \"description\": \"Duration of recent-history in which interface flapping will be considered. The longer the period, the more memory the probe will consume.\", \"type\": \"integer\", \"minimum\": 1, \"default\": 60}}}"
#}
#"traffic" = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = <<-EOT
#
#            This probe displays the all the interface counters available for the
#            system, their utilizations and utilizations aggregated on a per system
#            basis.
#
#    EOT
#  "name" = "traffic"
#  "schema" = "{\"type\": \"object\", \"properties\": {\"label\": {\"title\": \"Probe Label\", \"type\": \"string\", \"minLength\": 1, \"maxLength\": 120, \"default\": \"Device Traffic\"}, \"interface_counters_average_period\": {\"title\": \"Interface counters average period\", \"description\": \"The average period duration for interface counters\", \"type\": \"integer\", \"minimum\": 1, \"default\": 120}, \"enable_interface_counters_history\": {\"title\": \"Enable interface counters history\", \"description\": \"Maintain historical interface counters data\", \"type\": \"boolean\", \"default\": true}, \"interface_counters_history_retention_period\": {\"title\": \"Interface counters history retention period\", \"description\": \"Duration to maintain historical interface counters data\", \"type\": \"integer\", \"minimum\": 1, \"default\": 2592000}, \"enable_system_counters_history\": {\"title\": \"Enable system counters history\", \"description\": \"Maintain historical system interface counters data\", \"type\": \"boolean\", \"default\": true}, \"system_counters_history_retention_period\": {\"title\": \"System interface counters history retention period\", \"description\": \"Duration to maintain historical system interface counters data\", \"type\": \"integer\", \"minimum\": 1, \"default\": 2592000}}}"
#},
#"virtual_infra_hypervisor_redundancy_checks" = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = <<-EOT
#    Generate a probe to detect hypervisor redundancy.
#
#    EOT
#  "name" = "virtual_infra_hypervisor_redundancy_checks"
#  "schema" = "{\"type\": \"object\", \"properties\": {\"label\": {\"title\": \"Probe Label\", \"type\": \"string\", \"minLength\": 1, \"maxLength\": 120, \"default\": \"Hypervisor Redundancy Checks\"}}}"
#},
#"virtual_infra_lag_match" = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = <<-EOT
#    Generate a probe to detect inconsistent LAG configs between
#            fabric and virtual infra.
#
#            This probe calculates LAGs missing on hypervisors and AOS managed systems
#            connected to hypervisors.
#
#    EOT
#  "name" = "virtual_infra_lag_match"
#  "schema" = "{\"type\": \"object\", \"properties\": {\"label\": {\"title\": \"Probe Label\", \"type\": \"string\", \"minLength\": 1, \"maxLength\": 120, \"default\": \"Hypervisor & Fabric LAG Config Mismatch\"}}}"
#},
#"virtual_infra_missing_lldp" = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = <<-EOT
#    Generate a probe to detect virtual infra hosts that are not
#            configured for LLDP.
#
#    EOT
#  "name" = "virtual_infra_missing_lldp"
#  "schema" = "{\"type\": \"object\", \"properties\": {\"label\": {\"title\": \"Probe Label\", \"type\": \"string\", \"minLength\": 1, \"maxLength\": 120, \"default\": \"Hypervisor Missing LLDP Config\"}}}"
#},
#"virtual_infra_vlan_match" = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = <<-EOT
#    Generate a probe to calculate missing VLANs
#
#            This probe calculates VLAN(s) mismatch between AOS configured
#            virtual networks on the systems and the VLANs needed by the VMs running
#            on the hypervisors attached to the systems.
#
#    EOT
#  "name" = "virtual_infra_vlan_match"
#  "schema" = "{\"type\": \"object\", \"properties\": {\"label\": {\"title\": \"Probe Label\", \"type\": \"string\", \"minLength\": 1, \"maxLength\": 120, \"default\": \"Hypervisor & Fabric VLAN Config Mismatch\"}}}"
#},
#"vxlan_floodlist" = {
#  "blueprint_id" = "c151d0c1-fda1-495b-86e8-92d2499ac6f8"
#  "description" = <<-EOT
#
#            This probe validates the VXLAN flood list entries on every leaf
#            in the network. It collects appropriate telemetry data,
#            compares it to the set of flood list forwarding entries
#            expected to be present and alerts if expected entries are missing
#            on any device.
#
#            Route Labels
#
#            Expected: This route is expected on the device as per service
#                      defined.
#
#            Missing: This route is missing on the device when compared to
#                     the expected route set.
#
#            Unexpected: There are no expectations rendered (by AOS) for this
#                        route.
#
#    EOT
#  "name" = "vxlan_floodlist"
#  "schema" = "{\"type\": \"object\", \"properties\": {\"label\": {\"title\": \"Probe Label\", \"type\": \"string\", \"minLength\": 1, \"maxLength\": 120, \"default\": \"VXLAN Flood List Validation\"}, \"duration\": {\"title\": \"Anomaly Time Window\", \"type\": \"integer\", \"minimum\": 1, \"default\": 360}, \"threshold\": {\"title\": \"Anomaly Threshold (in %)\", \"description\": \"If routes are missing for more than or equal to percentage of Anomaly Time Window, an anomaly will be raised.\", \"type\": \"integer\", \"minimum\": 0, \"maximum\": 100, \"default\": 100}, \"collection_period\": {\"title\": \"Collection period\", \"description\": \"Telemetry collection interval.\", \"type\": \"number\", \"minimum\": 0, \"default\": 300.0}}}"
#}
#}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID. Used to identify the Blueprint that the IBA Predefined Probe belongs to.
- `name` (String) Populate this field to look up an IBA Predefined Probe.

### Read-Only

- `description` (String) Description of the IBA Predefined Probe
- `schema` (String) Schema of the IBA Predefined Probe's parameters
