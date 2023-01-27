resource "apstra_asn_pool_range" "t" {
  pool_id = apstra_asn_pool.t.id
  first = 11
  last = 11
}
