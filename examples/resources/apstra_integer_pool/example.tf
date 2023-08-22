# Simple example of an integer pool
resource "apstra_integer_pool" "example" {
  name = "example pool"
  ranges = [
    {
      first: 100
      last:199
    },
    {
      first: 300
      last:399
    },
  ]
}

# This example creates an Integer pool consisting of 5 ranges with random
# begin/end values. Because the ranges must not overlap, the randomly
# selected begin/end values are sorted (as text!) before being used
# in an integer pool range. If we randomly select the same value more than
# once, we'll wind up with fewer than 5 ranges.
locals {
  min_int        = 1000
  max_int        = 9999
  ranges_desired = 5
}

# generate 10 random values
resource "random_integer" "range_limits" {
  count = local.ranges_desired * 2
  min   = local.min_int
  max   = local.max_int
}

# unique-ify, count pairs, and sort
locals {
  unique_values    = distinct(random_integer.range_limits[*].result)
  pair_count       = floor(length(local.unique_values) / 2)
  unsorted_strings = formatlist("%08d", slice(local.unique_values, 0, local.pair_count * 2))
  sorted_strings   = sort(local.unsorted_strings)
  sorted_numbers   = [for s in local.sorted_strings : tonumber(s)]
}

# generate an integer pool with ranges equal to the number
# of begin/end pairs available.
resource "apstra_integer_pool" "five_random_ranges" {
  name = "five random ranges"
  ranges = [for i in range(local.pair_count) : {
    first = local.sorted_numbers[i * 2]
    last  = local.sorted_numbers[(i * 2) + 1]
  }]
}