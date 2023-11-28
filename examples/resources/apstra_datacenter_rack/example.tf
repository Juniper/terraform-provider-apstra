# This example creates a new rack in an existing blueprint.

resource "apstra_datacenter_rack" "r" {
  blueprint_id = "187458bf-7abf-450e-b341-d8da948bef9c"
  name         = "Rack_13"
  rack_type_id = "bq-lzzk6tmwc1redcw4rqg"
}
