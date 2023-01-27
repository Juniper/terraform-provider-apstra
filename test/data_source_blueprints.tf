data "apstra_blueprints" "t" {
  reference_design = "datacenter"
}

output "all_apstra_blueprints" {
  value = data.apstra_blueprints.t
}
