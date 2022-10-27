#The following example shows outputting all agent profile IDs.

data "apstra_agent_profiles" "all" {}

data "apstra_agent_profile" "all" {
  for_each = data.apstra_agent_profiles.all.ids
}

output "agent_profiles" {
  value = data.apstra_agent_profile.all
}
