# This example pulls details of the interconnect domain named "DC1"
# from blueprint "007723b7-a387-4bb3-8a5e-b5e9f265de0d"

data "apstra_datacenter_interconnect_domain" "example" {
  blueprint_id = "007723b7-a387-4bb3-8a5e-b5e9f265de0d"
  name         = "DC1"
}

outupt "example" { value = data.apstra_datacenter_interconnect_domain.example }

example = {
  "blueprint_id" = "007723b7-a387-4bb3-8a5e-b5e9f265de0d"
  "esi_mac" = "02:ff:ff:ff:ff:01"
  "id" = "On_FQ5OGrUYCDGW14A"
  "name" = "DC1"
  "route_target" = "100:100"
}

