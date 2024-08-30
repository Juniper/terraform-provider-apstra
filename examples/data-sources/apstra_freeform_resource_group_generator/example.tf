# This example creates a Resource Group Generator under a
# Resource Group in a Freeform Blueprint.
#
# After creating the Group Generator, the data source is invoked to look up
# the details.

# Create a resource group in a preexisting blueprint.
resource "apstra_freeform_resource_group" "fizz_grp" {
  blueprint_id = "f1b86583-9139-49ed-8a3c-0490253e006e"
  name         = "fizz_grp"
}

# Create a resource generator scoped to target all systems in the blueprint.
resource "apstra_freeform_resource_group_generator" "test_group_gen" {
  blueprint_id = "f1b86583-9139-49ed-8a3c-0490253e006e"
  group_id     = apstra_freeform_resource_group.fizz_grp.id
  name         = "test_res_gen"
  scope        = "node('system', name='target')"
}

# Invoke the resource group generator data source
data "apstra_freeform_resource_group_generator" "test_group_gen" {
  blueprint_id = "f1b86583-9139-49ed-8a3c-0490253e006e"
  id           = apstra_freeform_resource_group_generator.test_group_gen.id
}

# Output the data source so that it prints on screen
output "test_resource_group_generator_out" {
  value = data.apstra_freeform_resource_group_generator.test_group_gen
}

# The output looks like this:
# test_resource_group_generator_out = {
#   "blueprint_id" = "f1b86583-9139-49ed-8a3c-0490253e006e"
#   "id" = "wOQS9qZezzRJCgyvxwY"
#   "name" = "test_res_gen"
#   "scope" = "node('system', name='target')"
# }

