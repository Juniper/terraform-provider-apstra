# This example retrieves details about a Managed System Agent
# from the Apstra API.

data "apstra_agent" "foo" {
  agent_id = "45718d62-0342-4182-a285-36cca8f04db3"
}
