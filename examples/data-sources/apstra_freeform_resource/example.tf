resource "apstra_freeform_resource" "test" {
  blueprint_id      = "freeform_blueprint-d8c1fabf"
  name              = "test_resource_fizz"
  type              = "asn"
  group_id          = apstra_freeform_resource_group.fizz_grp.id
  integer_value     = "65535"
}

data "apstra_freeform_resource" "test" {
  blueprint_id = "freeform_blueprint-d8c1fabf"
  id           = apstra_freeform_resource.test.id
}

output "test_resource_out" { value = data.apstra_freeform_resource.test }

#test_resource_out = {
#  "allocated_from" = tostring(null)
#  "blueprint_id" = "freeform_blueprint-d8c1fabf"
#  "generator_id" = tostring(null)
#  "group_id" = "ZmPz_zw9UtgHJy3_00w"
#  "id" = "sAkaPFD_iyBVTVI1Y8M"
#  "integer_value" = 65535
#  "ipv4_value" = tostring(null)
#  "ipv6_value" = tostring(null)
#  "name" = "test_resource_fizz"
#  "type" = "asn"
#}

