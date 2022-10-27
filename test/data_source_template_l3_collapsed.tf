data "apstra_template_l3_collapsed" "foo" {
  id = "L3_Collapsed_ESI"
#  name = "Collapsed Fabric ESI"
}

output "l3_collapsed" {
  value = data.apstra_template_l3_collapsed.foo
}