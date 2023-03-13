# This example uses the 'apstra_templates' data source to create
# a list of template IDs of all evpn-enabled pod-based templates.
#
# The IDs are used to pull full details about each of those templates.
#
# Finally a report is ouput indicating the name associated with
# each ID.

data "apstra_templates" "rack_based" {
  type = "rack_based"
  overlay_control_protocol = "evpn"
}

data "apstra_template_rack_based" "selected" {
  for_each = data.apstra_templates.rack_based.ids
  id = each.key
}

output "rack_based_template_id_to_name" {
  value = { for k, v in data.apstra_template_rack_based.selected : k => v.name}
}