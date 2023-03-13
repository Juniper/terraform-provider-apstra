# This example creates an IPv4 Resource Pool with four allocations inside:
# - 3 blocks from RFC5737
# - a random chunk of 10.0.0.0/8
#
# After creating the pool, it outputs the total number of
# IP addresses in the pool.

# The random chunk of 10.0.0.0/8 needs three 8-bit values
# the three random octets.
resource "random_integer" "octets" {
  count = 3
  min   = 0
  max   = 255
}

# The random chunk of 10.0.0.0/8 needs an bit count
# between 8 and 32 inclusive.
resource "random_integer" "cidr_bits" {
  min = 8
  max = 32
}

# local.formatted assembles the random values and finds
# the correct "zero-host" CIDR boundary.
locals {
  formatted = format("%s/%d", cidrhost(format("10.%d.%d.%d/%d",
    random_integer.octets[0].result,
    random_integer.octets[1].result,
    random_integer.octets[2].result,
    random_integer.cidr_bits.result
  ), 0), random_integer.cidr_bits.result)
}

# Create the IPv4 address pool
resource "apstra_ipv4_pool" "example" {
  name = "RFC 5737 ranges, plus a random chunk of 10/8"
  subnets = [
    { network = local.formatted },
    { network = "192.0.2.0/24"},
    { network = "198.51.100.0/24"},
    { network = "203.0.113.0/24"},
  ]
}

output "example_pool_size" {
  value = format("pool '%s' is sized for %d addresses", apstra_ipv4_pool.example.name, apstra_ipv4_pool.example.total)
}