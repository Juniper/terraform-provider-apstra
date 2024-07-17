# here we build a link block

resource "apstra_freeform_link" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  name         = "link_a_b"
  speed = "1g"
  tags         = ["a", "b"]
  endpoints = [
    {
      system_id = "-CEYpa9xZ5chndvu0OY"
      interface_name = "ge-0/0/3"
      transformation_id = 1
    },
    {
      system_id = "ySBRdHvl2KZmWKLhkIk"
      interface_name = "ge-0/0/3"
      transformation_id = 1
    }
  ]
}

data "apstra_freeform_link" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  id           = apstra_freeform_link.test.id
}

output "test_Link_out" { value = data.apstra_freeform_link.test }


//output
#test_Link_out = {
#  "aggregate_link_id" = tostring(null)
#  "blueprint_id" = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
#  "endpoints" = tomap({
#    "-CEYpa9xZ5chndvu0OY" = {
#      "interface_id" = "c459DMed3P42wapAtUY"
#      "interface_name" = "ge-0/0/3"
#      "ipv4_address" = tostring(null)
#      "ipv6_address" = tostring(null)
#      "transformation_id" = 1
#    }
#    "ySBRdHvl2KZmWKLhkIk" = {
#      "interface_id" = "1wWgi25jmyZ5NBy45dA"
#      "interface_name" = "ge-0/0/3"
#      "ipv4_address" = tostring(null)
#      "ipv6_address" = tostring(null)
#      "transformation_id" = 1
#    }
#  })
#  "id" = "SkY0hved7LajZY7WNzU"
#  "name" = "link_a_b"
#  "speed" = "10G"
#  "tags" = toset([
#    "a",
#    "b",
#  ])
#  "type" = "ethernet"
#}

