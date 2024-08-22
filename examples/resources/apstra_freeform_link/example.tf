# This example creates a link between systems with IDs "CEYpa9xZ5chndvu0OYa"
# and "ySBRdHvl2KZmWKLhkIk" in a Freeform Blueprint

resource "apstra_freeform_link" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  name         = "link_a_b"
  tags         = ["a", "b"]
  endpoints = {
    CEYpa9xZ5chndvu0OYa = {
      interface_name    = "ge-0/0/3"
      transformation_id = 1
      tags              = ["prod", "native_1000BASE-T"]
    },
    ySBRdHvl2KZmWKLhkIk = {
      interface_name    = "ge-0/0/3"
      transformation_id = 1
      tags              = ["prod", "requires_transceiver"]
    }
  }
}
