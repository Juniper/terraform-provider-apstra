# This example searches for an Agent Profile matching:
# - platform: "junos"
# - has_username: true
# - has_password: true
#
# It then uses that Agent Profile to create a Managed Device resource for
# a device at a known Management IP. Because the serial number is known
# in advance, the Managed Device will be automaticall "acknowledged" in
# Apstra's "Devices -> Managed Devices" list.

# Get a list of all Agent Profile IDs
data "apstra_agent_profiles" "all" {}

# Use the Agent Profile IDs to pull details about each Agent Profile
data "apstra_agent_profile" "each" {
  for_each = data.apstra_agent_profiles.all.ids
  id       = each.key
}

# Filter the Agent Profile map to include only those with
# credentials for Junos boxes.
locals {
  agent_profiles_id_with_credentials_and_qfx_in_name = [
    for dp in data.apstra_agent_profile.each :
    dp.id if length(regexall("qfx", lower(dp.name))) >= 1 &&
    dp.platform == "junos" &&
    dp.has_username &&
    dp.has_password
  ]
}

# Create the Managed Device resource with a serial number (device key)
# so that the system will be automatically "acknowledged"
resource "apstra_managed_device" "example" {
  agent_profile_id = local.agent_profiles_id_with_credentials_and_qfx_in_name[0]
  device_key = "52540057C718"
  management_ip = "172.20.84.15"
  off_box = true
}