# This example defines a freeform blueprint, then looks it up 
# both by name and by ID.

resource "apstra_freeform_blueprint" "test" {
  name = "bar"
}

data "apstra_freeform_blueprint" "by_name" {
  name = apstra_freeform_blueprint.test.name
}

data "apstra_freeform_blueprint" "by_id" {
  id = apstra_freeform_blueprint.test.id
}

output "by_name" { value = data.apstra_freeform_blueprint.by_name }
output "by_id" { value = data.apstra_freeform_blueprint.by_id }

# The output looks like:
# by_id = {
#   "id" = "37bc0ff6-b9fe-44b3-aa69-1a0ac0368421"
#   "name" = "bar"
# }
# by_name = {
#   "id" = "37bc0ff6-b9fe-44b3-aa69-1a0ac0368421"
#   "name" = "bar"
# }
