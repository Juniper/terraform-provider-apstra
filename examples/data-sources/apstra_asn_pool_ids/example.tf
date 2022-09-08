# The following example shows outputting all ASN pool IDs.

data "apstra_asn_pool_ids" "all" {}

output asn_pool_ids {
   value = data.apstra_asn_pool_ids.all.ids
}
