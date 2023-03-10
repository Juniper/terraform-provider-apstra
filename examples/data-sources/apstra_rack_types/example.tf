# This example uses `apstra_rack_types` and `apstra_rack_type` to produce a
# report of Rack Types with 40 or more generic systems.
#
# First, collect all Rack Type IDs
data "apstra_rack_types" "all" {}

# Loop over Rack Type IDs, collect full details of each Rack Type
data "apstra_rack_type" "each" {
  for_each = data.apstra_rack_types.all.ids
  id       = each.key
}

# Create a map of Rack Type name to Generic System count, but only include Rack
# Types with 40 or more Generic Systems
output "racks_with_mlag_leafs" {
  value = {
    for rt in data.apstra_rack_type.each : rt.id => sum([
      for gs in rt.generic_systems : gs.count
    ]) if sum([for gs in rt.generic_systems : gs.count]) >= 40
  }
}
