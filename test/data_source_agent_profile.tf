##The following example shows outputting all agent profile IDs.
#
#resource "apstra_agent_profile" "for_junos" {   // create an agent profile for
#  name     = "junos_profile"                    // junos boxes so we'll have
#  platform = "junos"                            // something in the output
#}
#
#resource "apstra_agent_profile" "for_nxos" {    // create an agent profile for
#  name     = "nxos_profile"                     // nxos boxes so we'll have
#  platform = "nxos"                             // something to filter out
#}
#
#data "apstra_agent_profiles" "all" {}           // fetch all agent profile IDs
#
#data "apstra_agent_profile" "all" {             // fetch the full details of
#  for_each = data.apstra_agent_profiles.all.ids // each agent profile, by ID
#  id       = each.key
#}
#
#output "junos_agent_profile_ids" {              // output IDs of junos agent profiles only
#  value = [for k, v in data.apstra_agent_profile.all : k if v.platform == "junos"]
#}
#
## The output looks like:
##
##    junos_agent_profile_ids = [
##      "7796cca6-7637-43ac-a1a1-ada6eb50ae56",
##    ]
