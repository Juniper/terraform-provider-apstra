# This example returns the IDs of all Off-Box Agents responsible for Systems
# with management IP addresses in the 192.168.100.0/24 subnet.

data "apstra_agents" "agents" {
  filters = {
    management_ip    = "192.168.100.0/24"
    off_box          = true
  }
}
