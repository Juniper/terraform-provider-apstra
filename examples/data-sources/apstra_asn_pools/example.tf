# The following example shows outputting all ASN pool IDs.

data "apstra_asn_pools" "all" {}

output asn_pool_ids {
   value = data.apstra_asn_pools.all.ids
}
