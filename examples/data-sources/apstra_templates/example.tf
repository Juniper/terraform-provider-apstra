# This example uses the 'apstra_templates' data source to create
# a list of template IDs of all evpn-enabled pod-based templates.
data "apstra_templates" "all" {
  type = "pod_based"
  overlay_control_protocol = "evpn"
}

output "templates" {
  value = data.apstra_templates.all.ids
}