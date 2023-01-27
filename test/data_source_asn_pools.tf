data "apstra_asn_pools" "t" {}

output "apstra_asn_pools" {
  value = data.apstra_asn_pools.t
}
