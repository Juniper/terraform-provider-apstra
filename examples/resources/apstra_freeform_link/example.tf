# This example creates a link between systems "-CEYpa9xZ5chndvu0OY" and
# "ySBRdHvl2KZmWKLhkIk" in a Freeform Blueprint

resource "apstra_freeform_link" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  name         = "link_a_b"
  tags         = ["a", "b"]
  endpoints = [
    {
      system_id         = "-CEYpa9xZ5chndvu0OY"
      interface_name    = "ge-0/0/3"
      transformation_id = 1
      tags              = ["prod", "native_1000BASE-T"]
    },
    {
      system_id         = "ySBRdHvl2KZmWKLhkIk"
      interface_name    = "ge-0/0/3"
      transformation_id = 1
      tags              = ["prod", "requires_transceiver"]
    }
  ]
}
