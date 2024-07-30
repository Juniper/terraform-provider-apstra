# This example creates an ASN resource with the value 65535 within a
# preexisting resource group in a Freeform Blueprint.
#
# After creating the Resource, the data source is invoked to look up
# the details.
resource "apstra_freeform_resource" "test" {
  blueprint_id  = "631f8832-ae59-40ca-b4f6-9c19b411aeaf"
  name          = "aa"
  type          = "asn"
  group_id      = "Sj7LlWkPucErjSDx4vc"
  integer_value = "65535"
}

data "apstra_freeform_resource" "test" {
  blueprint_id = "631f8832-ae59-40ca-b4f6-9c19b411aeaf"
  id           = apstra_freeform_resource.test.id
}

output "test_resource_out" { value = data.apstra_freeform_resource.test }

# The output looks like:
# test_resource_out = {
#   "allocated_from" = tostring(null)
#   "blueprint_id" = "631f8832-ae59-40ca-b4f6-9c19b411aeaf"
#   "generator_id" = tostring(null)
#   "group_id" = "Sj7LlWkPucErjSDx4vc"
#   "id" = "OkPxk02GAGN5H8z9FWU"
#   "integer_value" = 65535
#   "ipv4_value" = tostring(null)
#   "ipv6_value" = tostring(null)
#   "name" = "aa"
#   "type" = "asn"
# }

