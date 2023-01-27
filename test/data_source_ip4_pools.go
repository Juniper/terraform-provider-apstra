data "apstra_ip4_pools" "t" {}

output "apstra_ip4_pools" {
  value = data.apstra_ip4_pools
}
