# This example outputs a list of blueprint IDs
data "apstra_blueprints" "d" {
  reference_design = "datacenter" // optional filter argument
}

output "apstra_blueprints" {
  value = data.apstra_blueprints.d
}