# This example defines a Freeform Resource Allocation Group in a blueprint

resource "apstra_freeform_resource_group" "test" {
  blueprint_id      = "freeform_blueprint-d8c1fabf"
  name              = "test_resource_group_fizz"
  data              =  jsonencode({
    foo   = "bar"
    clown = 2
  })
}

# here we retrieve the freeform resource_group

data "apstra_freeform_resource_group" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  id = apstra_freeform_resource_group.test.id
}

# here we build an output bock to display it

output "test_resource_group_out" {value = data.apstra_freeform_resource_group.test}

//test_resource_group_out = {
//  "blueprint_id" = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
//  "data" = "{\"clown\": 2, \"foo\": \"bar\"}"
//  "generator_id" = tostring(null)
//  "id" = "98ubU5cuRj7WsT159L4"
//  "name" = "test_resource_group_fizz"
//  "parent_id" = tostring(null)
//}

