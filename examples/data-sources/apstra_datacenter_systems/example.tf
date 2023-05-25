# This example outputs a set of graph db node IDs representing all spine
# switches with tag 'junos' and tag 'qfx'
data "apstra_datacenter_systems" "juniper_spines" {
  blueprint_id = apstra_datacenter_blueprint.example.id
  filters = {
    role        = "spine"
    system_type = "switch"
    tag_ids     = ["junos", "qfx"]
  }
}

output "qfx_spines" {
  value = data.apstra_datacenter_systems.juniper_spines.ids
}
