resource "apstra_telemetry_service_registry_entry" "maketest" {
	application_schema = jsonencode(
		{
			properties = {
				key = {
					properties = {
						authenticated_vlan = {
							type = "string"
						}
						authorization_status = {
							type = "string"
						}
						fallback_vlan_active = {
							enum = [
								"True",
								"False",
							]
							type = "string"
						}
						port_status = {
							enum = [
								"authorized",
								"blocked",
							]
							type = "string"
						}
						supplicant_mac = {
							pattern = "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
							type    = "string"
						}
					}
					required = [
						"supplicant_mac",
						"authenticated_vlan",
						"authorization_status",
						"port_status",
						"fallback_vlan_active",
					]
					type = "object"
				}
				value = {
					description = "0 in case of blocked, 1 in case of authorized"
					type        = "integer"
				}
			}
			required = [
				"key",
				"value",
			]
			type = "object"
		}
	)
	description         = "Test Telemetry Service B"
	service_name        = "TestTelemetryServiceC"
	storage_schema_path = "aos.sdk.telemetry.schemas.iba_integer_data"
}

output "r" {
	value = apstra_telemetry_service_registry_entry.maketest
}

#Output looks something like this
#Outputs:
#
#r = {
#"application_schema" = "{\"properties\":{\"key\":{\"properties\":{\"authenticated_vlan\":{\"type\":\"string\"},\"authorization_status\":{\"type\":\"string\"},\"fallback_vlan_active\":{\"enum\":[\"True\",\"False\"],\"type\":\"string\"},\"port_status\":{\"enum\":[\"authorized\",\"blocked\"],\"type\":\"string\"},\"supplicant_mac\":{\"pattern\":\"^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$\",\"type\":\"string\"}},\"required\":[\"supplicant_mac\",\"authenticated_vlan\",\"authorization_status\",\"port_status\",\"fallback_vlan_active\"],\"type\":\"object\"},\"value\":{\"description\":\"0 in case of blocked, 1 in case of authorized\",\"type\":\"integer\"}},\"required\":[\"key\",\"value\"],\"type\":\"object\"}"
#"built_in" = false
#"description" = "Test Telemetry Service B"
#"service_name" = "TestTelemetryServiceC"
#"storage_schema_path" = "aos.sdk.telemetry.schemas.iba_integer_data"
#"version" = ""
#}
