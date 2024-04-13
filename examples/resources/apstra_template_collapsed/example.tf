# This example creates a Collapsed Template based on the
# L3_collapsed_acs built-in rack type

resource "apstra_template_collapsed" "example" {
  name            = "example collapsed template"
  rack_type_id    = "L3_collapsed_acs"
  mesh_link_speed = "10G"
  mesh_link_count = 2
}

