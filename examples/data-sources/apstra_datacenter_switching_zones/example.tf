# Without specifying no filter, a wide search is performed.
# All switching zones in the blueprint will match.
data "apstra_datacenter_switching_zones" "all" {
  blueprint_id = "05f9d3fc-671a-4efc-8e91-5ef87b2937d3"
}

# This example performs a very narrow search. Only one (or zero!)
# switching zones can match the resulting query.
data "apstra_datacenter_switching_zones" "rzs" {
  blueprint_id = "05f9d3fc-671a-4efc-8e91-5ef87b2937d3"
  filters = [
    { # all filter attributes are optional
      name = "customer_1"
    },
  ]
}
