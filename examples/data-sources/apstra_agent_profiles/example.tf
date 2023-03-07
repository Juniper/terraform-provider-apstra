# The following example grabs the ID numbers of all agent profiles, uses those
# IDs to grab the details of each agent profile, and then outputs the names of
# all Agent Profiles which lack a complete set of credentials.

data "apstra_agent_profiles" "all" {}

data "apstra_agent_profile" "all" {
  for_each = data.apstra_agent_profiles.all.ids
  id       = each.key
}

output "agent_profiles_missing_credentials" {
  value = [
    for k, v in data.apstra_agent_profile.all : v.name if !v.has_username || !v.has_password
  ]
}
