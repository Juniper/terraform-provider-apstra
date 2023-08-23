# Find IDs of leaf switch interfaces with links tagged
# "dev", "linux", and "backend"
data "apstra_datacenter_interface_ids_by_link_tag" "x" {
  blueprint_id = "fa6782cc-c4d5-4933-ad89-e542acd6b0c1"
  system_type  = "switch" // optional
  system_role  = "leaf"   // optional
  tags         = ["dev", "linux", "backend"]
}
