# This example pulls details of the external gateway named "DC2A"
# from blueprint "007723b7-a387-4bb3-8a5e-b5e9f265de0d"

data "apstra_datacenter_external_gateway" "example" {
  blueprint_id = "007723b7-a387-4bb3-8a5e-b5e9f265de0d"
  name         = "DC2A"
}
