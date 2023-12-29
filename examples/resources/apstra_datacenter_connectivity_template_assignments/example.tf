# This example assigns connectivity template ef77e286-be82-4855-baaf-269d2a0ed893
# to four different switch ports.

# identified by the ID "FkYtMBdeoJ5urBaIEi8"
#
# Data sources like these can be used to find node IDs to use in
# the `application_point_ids` attribute:
# - apstra_datacenter_svis_map
# - apstra_datacenter_interfaces_by_link_tag
# - apstra_datacenter_interfaces_by_system

resource "apstra_datacenter_connectivity_template_assignments" "a" {
  blueprint_id             = "0b1d5276-37e0-46fb-8d35-b8932015e56c"
  connectivity_template_id = "ef77e286-be82-4855-baaf-269d2a0ed893"
  application_point_ids = [
    "nC7HblArEjHdVkzrGAo",
    "nCPDQ_nYXJ5qvQI4MiE",
    "nCuWV_8yMbtL2EIrbT0",
    "nD4Ae9RkblAgvrOUjCQ",
  ]
}
