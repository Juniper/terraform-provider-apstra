data "apstra_ip4_pool" "t" {
  name = "Private-10.0.0.0/8"
}

output "apstra_ip4_pool" {
  value = data.apstra_ip4_pool.t
}
