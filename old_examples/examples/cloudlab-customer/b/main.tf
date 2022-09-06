# Create an IPv4 address pool
resource "apstra_ip4_pool" "my_pool" {
  name = "terraform-pool"
}



















# Create an address range within the pool
#resource "apstra_ip4_pool_subnet" "my_subnet" {
#  pool_id = "xxxxxx"
#  cidr = "4.0.0.0/24"
#}