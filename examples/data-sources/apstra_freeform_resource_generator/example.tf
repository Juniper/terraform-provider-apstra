# This example creates an ASN resource Generator within a
# preexisting resource group in a Freeform Blueprint.
#
# After creating the Resource Generator, the data source is invoked to look up
# the details.
resource "apstra_freeform_resource_group"  "fizz_grp" {
  blueprint_id      = "631f8832-ae59-40ca-b4f6-9c19b411aeaf"
  name              = "fizz_grp"
}

resource "apstra_asn_pool" "rfc5398" {
  name = "RFC5398 ASN"
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

resource "apstra_freeform_alloc_group" "test" {
  blueprint_id      = "631f8832-ae59-40ca-b4f6-9c19b411aeaf"
  name              = "test_alloc_group2"
  type              = "asn"
  pool_ids          = [apstra_asn_pool.rfc5398.id]
}

resource "apstra_freeform_resource_generator" "test_res_gen" {
  blueprint_id   =  "631f8832-ae59-40ca-b4f6-9c19b411aeaf"
  name = "test_res_gen"
  type = "asn"
  scope = "node('system', name='target')"
  allocated_from = apstra_freeform_alloc_group.test.id
  container_id = apstra_freeform_resource_group.fizz_grp.id
}



# The output looks like:
#test_resource_generator_out = {
#  "allocated_from" = "rag_asn_test_alloc_group2"
#  "blueprint_id" = "631f8832-ae59-40ca-b4f6-9c19b411aeaf"
#  "container_id" = "gPJtXP7_SM31CYWDJ0g"
#  "id" = "EkrP9avh6pgqRqRCm44"
#  "name" = "test_res_gen"
#  "scope" = "node('system', name='target')"
#  "subnet_prefix_len" = tonumber(null)
#  "type" = "asn"
#}
