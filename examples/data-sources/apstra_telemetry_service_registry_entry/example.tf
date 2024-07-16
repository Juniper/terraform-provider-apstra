# The following example shows getting a list of
# Apstra Telemetry Service Registy Entries

#Get Only non-built-in Entries
data "apstra_telemetry_service_registry_entries" "not_builtin" {
  built_in = false
}

data "apstra_telemetry_service_registry_entry" "all_not_built_in" {
  for_each     = data.apstra_telemetry_service_registry_entries.not_builtin.names
  name = each.value
}

output "not_built_in_entry" {
  value = data.apstra_telemetry_service_registry_entry.all_not_built_in
}

#Output looks something like this.
#Outputs
#not_built_in_entry = {
#  "TestTelemetryServiceA" = {
#    "application_schema" = "{\"properties\": {\"key\": {\"properties\": {\"authenticated_vlan\": {\"type\": \"string\"}, \"authorization_status\": {\"type\": \"string\"}, \"fallback_vlan_active\": {\"enum\": [\"True\", \"False\"], \"type\": \"string\"}, \"port_status\": {\"enum\": [\"authorized\", \"blocked\"], \"type\": \"string\"}, \"supplicant_mac\": {\"pattern\": \"^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$\", \"type\": \"string\"}}, \"required\": [\"supplicant_mac\", \"authenticated_vlan\", \"authorization_status\", \"port_status\", \"fallback_vlan_active\"], \"type\": \"object\"}, \"value\": {\"description\": \"0 in case of blocked, 1 in case of authorized\", \"type\": \"integer\"}}, \"required\": [\"key\", \"value\"], \"type\": \"object\"}"
#    "built_in" = false
#    "description" = "Test Telemetry Service A"
#    "name" = "TestTelemetryServiceA"
#    "storage_schema_path" = "aos.sdk.telemetry.schemas.iba_integer_data"
#    "version" = ""
#  }
#  "TestTelemetryServiceC" = {
#    "application_schema" = "{\"properties\": {\"key\": {\"properties\": {\"authenticated_vlan\": {\"type\": \"string\"}, \"authorization_status\": {\"type\": \"string\"}, \"fallback_vlan_active\": {\"enum\": [\"True\", \"False\"], \"type\": \"string\"}, \"port_status\": {\"enum\": [\"authorized\", \"blocked\"], \"type\": \"string\"}, \"supplicant_mac\": {\"pattern\": \"^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$\", \"type\": \"string\"}}, \"required\": [\"supplicant_mac\", \"authenticated_vlan\", \"authorization_status\", \"port_status\", \"fallback_vlan_active\"], \"type\": \"object\"}, \"value\": {\"description\": \"0 in case of blocked, 1 in case of authorized\", \"type\": \"integer\"}}, \"required\": [\"key\", \"value\"], \"type\": \"object\"}"
#    "built_in" = false
#    "description" = "Test Telemetry Service B"
#    "name" = "TestTelemetryServiceC"
#    "storage_schema_path" = "aos.sdk.telemetry.schemas.iba_integer_data"
#    "version" = ""
#  }
#}
