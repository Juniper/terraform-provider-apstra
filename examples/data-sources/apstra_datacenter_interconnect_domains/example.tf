# This example collects the IDs of any Interconnect Gateways which are configured
# with route_target 100:100 or 200:200

// find interesting Interconnect Domain IDs
data "apstra_datacenter_interconnect_domains" "selected" {
  blueprint_id = "007723b7-a387-4bb3-8a5e-b5e9f265de0d"
  filters = [ { route_target = "100:100" }, { route_target = "200:200" } ]
}

output "interconnect_domain_ids" { value = data.apstra_datacenter_interconnect_domains.selected.ids }

// The output looks like this:
//
//   interconnect_domain_ids = toset([
//     "On_FQ5OGrUYCDGW14A",
//     "nn2Q3ar3ERM4okQ0cw",
//   ])
