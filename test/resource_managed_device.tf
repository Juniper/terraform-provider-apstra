data "apstra_agent_profile" "qfx" {
#  name = "qfx"
  name = "qfx copy"
}
resource "apstra_managed_device" "amd" {
  management_ip = "172.20.102.13"
  device_key = "5254004927F5"
  agent_profile_id = data.apstra_agent_profile.qfx.id
  off_box = true
}