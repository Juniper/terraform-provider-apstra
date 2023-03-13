# This example creates an ASN pool with two ranges.
resource "apstra_asn_pool" "rfc5398" {
  name = "RFC5398 ASNs"
  ranges = [
    {
      first = 64496
      last = 64511
    },
    {
      first = 65536
      last = 65551
    },
  ]
}
