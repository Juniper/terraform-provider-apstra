data "apstra_agent_profiles" "t" {}

output "apstra_agent_profiles" {
  value = data.apstra_agent_profiles.t
}
