# This example creates IPv6 pools with subnets from an IPv6 allocation with
# configuration options for pool count, pool size, subnet count and subnet size.
#
# Configurable options in this block
locals {
  allocation       = "2001:db8::/32" # RFC3849
  pool_size        = "/48"
  subnet_size      = "/64"
  pool_count       = 5
  subnets_per_pool = 10
}

# Create an iterable ipv6 pool resource. Each iteration
# references a different list of subnet objects.
resource "apstra_ipv6_pool" "r" {
  count   = local.pool_count
  name    = "RFC3849 ${count.index + 1} of ${local.pool_count}"
  subnets = local.subnet_allocations[count.index]
}

# do not edit - these 'local' variables are used to simplify what would
# otherwise be a crazy complicated resource block.
locals {
  # parse the bit size from the text inputs above
  allocation_bits = parseint(split("/", local.allocation)[1], 10)
  pool_bits       = parseint(split("/", local.pool_size)[1], 10)
  subnet_bits     = parseint(split("/", local.subnet_size)[1], 10)

  # a little bit of subtraction gets us the 'new bits' format needed by `cidrsubnet()`
  new_bits_per_pool   = local.pool_bits - local.allocation_bits
  new_bits_per_subnet = local.subnet_bits - local.pool_bits

  # build a list of allocation blocks for the IPv6 pools
  pool_allocations = [
    for i in range(local.pool_count) : cidrsubnet(local.allocation, local.new_bits_per_pool, i)
  ]

  # build a list of subnet blocks within each pool allocation
  subnet_allocations = [
    for i in range(local.pool_count) : [
      for j in range(local.subnets_per_pool) : {
        network = cidrsubnet(local.pool_allocations[i], local.new_bits_per_subnet, j)
      }
    ]
  ]
}
