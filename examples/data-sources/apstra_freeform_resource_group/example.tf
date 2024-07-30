# This example creates a Resource Allocation Group in a Freeform Blueprint

resource "apstra_freeform_resource_group" "test" {
  blueprint_id = "631f8832-ae59-40ca-b4f6-9c19b411aeaf"
  name         = "test_resource_group_fizz"
  data = jsonencode({
    foo   = "bar"
    clown = 2
  })
}

# The apstra_freeform_resource_group data source can select a resource
# group using Blueprint ID with Resource Group ID or Resource Group name.

data "apstra_freeform_resource_group" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  id           = apstra_freeform_resource_group.test.id
}

output "test_resource_group_out" { value = data.apstra_freeform_resource_group.test }

# The output looks like:
# test_resource_group_out = {
#   "blueprint_id" = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
#   "data" = "{\"clown\": 2, \"foo\": \"bar\"}"
#   "generator_id" = tostring(null)
#   "id" = "98ubU5cuRj7WsT159L4"
#   "name" = "test_resource_group_fizz"
#   "parent_id" = tostring(null)
# }

