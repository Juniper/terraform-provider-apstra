# This example outputs a set of graph db node IDs representing
# non-prod leaf switches in pod 2.
data "apstra_datacenter_systems" "pod2_nonprod_leafs" {
  blueprint_id = apstra_datacenter_blueprint.example.id
  filters = [
    {
      role        = "leaf"
      system_type = "switch"
      tag_ids     = ["pod2", "dev"]
    },
    {
      role        = "leaf"
      system_type = "switch"
      tag_ids     = ["pod2", "test"]
    },
  ]
}

output "pod2_nonprod_leafs" {
  value = data.apstra_datacenter_systems.pod2_nonprod_leafs.ids
}
