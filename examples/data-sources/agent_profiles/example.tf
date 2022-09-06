#The following example shows outputting all agent profile IDs.

data "apstra_agent_profiles" "all" {}

output "agent_profiles" {
  value = data.apstra_agent_profiles.all
}
