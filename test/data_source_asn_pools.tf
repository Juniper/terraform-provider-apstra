data "apstra_asn_pools" "t" {}

output "data_source_apstra_asn_pools" {
  value = data.apstra_asn_pools.t
}
