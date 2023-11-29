# This example creates a new rack based on an existing
# Rack Type (design object) in an existing blueprint.

resource "apstra_datacenter_rack" "r" {
  blueprint_id = "187458bf-7abf-450e-b341-d8da948bef9c"
  rack_name    = "Rack_13"
  rack_type_id = "bq-lzzk6tmwc1redcw4rqg"
}
