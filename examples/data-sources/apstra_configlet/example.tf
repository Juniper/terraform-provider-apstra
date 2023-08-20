# This example uses the `apstra_configlet` data source to produce a report
# indicating the number of sections (generators) which appear in configlets
# applicable to Junos platform devices.
#
# Get a list of configlet IDs for Junos platform devices.
data "apstra_configlets" "junos_only" {
  supported_platforms = ["junos"]
}

# Use the ID numbers to pull full details of those configlets.
data "apstra_configlet" "junos" {
  for_each = data.apstra_configlets.junos_only.ids
  id       = each.key
}

# Loop over all configlets. Within each configlet, filter out Junos-style
# generators. Count the generators.
output "junos_section_count" {
  value = {
    for id, configlet in data.apstra_configlet.junos :
    id => length([
      for gen in configlet.data.generators : gen if gen.config_style == "junos"
    ])
  }
}
