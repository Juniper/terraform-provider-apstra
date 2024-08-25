resource "apstra_freeform_resource_group"  "fizz_grp" {
  blueprint_id      = "6f1e3a4e-2920-451f-b069-c256c9fff938"
  name              = "fizz_grp"
}

resource "apstra_asn_pool" "rfc5398" {
  name = "RFC5398 ASNs"
  ranges = [
    {
      first = 64496
      last = 64511
    },
    {
      first = 65536
      last = 65551
    },
  ]
}

resource "apstra_freeform_allocation_group" "test" {
  blueprint_id      = "6f1e3a4e-2920-451f-b069-c256c9fff938"
  name              = "test_allocation_group2"
  type              = "asn"
  pool_id          = [apstra_asn_pool.rfc5398.id]
}

data "apstra_freeform_allocation_group" "test" {
  blueprint_id = "6f1e3a4e-2920-451f-b069-c256c9fff938"
  id           =   apstra_freeform_allocation_group.test.id
}

output "test_resource_out" { value = data.apstra_freeform_allocation_group.test }

# output produced =
#test_resource_out = {
#  "blueprint_id" = "6f1e3a4e-2920-451f-b069-c256c9fff938"
#  "id" = "rag_asn_test_allocation_group2"
#  "name" = "test_allocation_group2"
#  "pool_id" = toset([
#    "5d945d83-3361-4e9f-8a35-e935d549c298",
#  ])
#  "type" = "asn"
#}