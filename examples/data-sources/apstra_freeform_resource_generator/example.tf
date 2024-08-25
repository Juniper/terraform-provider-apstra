# This example creates an ASN resource Generator within a
# preexisting resource group in a Freeform Blueprint.
#
# After creating the Resource Generator, the data source is invoked to look up
# the details.

# Create a resource group in a preexisting blueprint.
resource "apstra_freeform_resource_group" "fizz_grp" {
  blueprint_id = "f1b86583-9139-49ed-8a3c-0490253e006e"
  name         = "fizz_grp"
}

# Create an allocation group in a preexisting blueprint
# using a preexisting ASN pool.
resource "apstra_freeform_allocation_group" "test" {
  blueprint_id = "f1b86583-9139-49ed-8a3c-0490253e006e"
  name         = "test_allocation_group2"
  type         = "asn"
  pool_ids     = ["Private-64512-65534"]
}

# Create a resource generator scoped to target all systems in the blueprint.
resource "apstra_freeform_resource_generator" "test_res_gen" {
  blueprint_id   = "f1b86583-9139-49ed-8a3c-0490253e006e"
  name           = "test_res_gen"
  type           = "asn"
  scope          = "node('system', name='target')"
  allocated_from = apstra_freeform_allocation_group.test.id
  container_id   = apstra_freeform_resource_group.fizz_grp.id
}

# Invoke the resource generator data source
data "apstra_freeform_resource_generator" "test_res_gen" {
  blueprint_id   = "f1b86583-9139-49ed-8a3c-0490253e006e"
  id = apstra_freeform_resource_generator.test_res_gen.id
}

# Output the data source so that it prints on screen
output "test_resource_generator_out" {
  value = data.apstra_freeform_resource_generator.test_res_gen
}

# The output looks like this:
# test_resource_generator_out = {
#   "allocated_from" = "rag_asn_test_allocation_group2"
#   "blueprint_id" = "f1b86583-9139-49ed-8a3c-0490253e006e"
#   "container_id" = "NbzITYcIjPZN4ZBqS2Q"
#   "id" = "pwJ9EOiVR8z6qbj2Ou8"
#   "name" = "test_res_gen"
#   "scope" = "node('system', name='target')"
#   "subnet_prefix_len" = tonumber(null)
#   "type" = "asn"
# }
