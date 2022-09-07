# The following example shows how a module might accept an agent profile's name
# as an input variable and use it to retrieve the agent profile ID when
# provisioning a managed device (a switch).

variable "agent_profile_name" {}
variable "switch_mgmt_ip" {}

data "apstra_agent_profile" "selected" {
  name = var.agent_profile_name
}

resource "apstra_managed_device" "switch" {
  agent_profile_id = data.apstra_agent_profile.selected.id
  management_ip    = var.switch_mgmt_ip
}
