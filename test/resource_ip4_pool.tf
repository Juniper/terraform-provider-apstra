resource "apstra_ip4_pool" "pool4" {
  name = "tf_pool"
}

resource "apstra_ip4_pool_subnet" "subnet" {
  pool_id = apstra_ip4_pool.pool4.id
  cidr = "192.168.1.0/24"
}