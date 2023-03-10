# This example uses the `apstra_configlets` data source to list IDs of
# configlets which support "junos" and "sonic" devices.
data "apstra_configlets" "two_specific_platforms" {
  supported_platforms = ["junos", "sonic"]
}

output "configlets" {
  value = data.apstra_configlets.two_specific_platforms.ids
}
