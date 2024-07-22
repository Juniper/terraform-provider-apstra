# This example defines a Freeform Resource Allocation Group in a blueprint

resource "apstra_freeform_ra_group" "test" {
  blueprint_id      = "freeform_blueprint-d8c1fabf"
  name              = "test_ra_group_fizz"
  tags              = ["a", "b", "c"]
  data              =  jsonencode({
    foo   = "bar"
    clown = 2
  })
}

# here we retrieve the freeform ra_group

data "apstra_freeform_ra_group" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  id = apstra_freeform_ra_group.test.id
}

# here we build an output bock to display it

output "test_ra_group_out" {value = data.apstra_freeform_ra_group.test}

//test_ra_group_out = {
//  "blueprint_id" = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
//  "data" = "{\"clown\": 2, \"foo\": \"bar\"}"
//  "generator_id" = tostring(null)
//  "id" = "98ubU5cuRj7WsT159L4"
//  "name" = "test_ra_group_fizz"
//  "parent_id" = tostring(null)
//  "tags" = toset([
//    "a",
//    "b",
//    "c",
//  ])
//}

