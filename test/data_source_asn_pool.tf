data "apstra_asn_pool" "t" {
  name = "Private-64512-65534"
}

output "apstra_asn_pool" {
  value = data.apstra_asn_pool.t
}
