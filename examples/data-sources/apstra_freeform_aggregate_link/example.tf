# The apstra_freeform_aggregate_link data source looks up details of an
# aggregate link in a freeform blueprint by name or by ID.

data "apstra_freeform_aggregate_link" "by_id" {
  blueprint_id = "bbc9376c-fd03-49a6-ba9b-6ca5c98387f8"
  id           = "Bzbh3t9Qx1z5VsGynw"
  name         = null
}

data "apstra_freeform_aggregate_link" "by_name" {
  blueprint_id = "bbc9376c-fd03-49a6-ba9b-6ca5c98387f8"
  id           = null
  name         = "some_link_label"
}
