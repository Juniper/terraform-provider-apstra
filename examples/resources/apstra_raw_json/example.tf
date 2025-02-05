# This example uses the `apstra_raw_json` resource to create an IP pool.

# Don't ever do this. Use the `apstra_ipv4_pool` resource instead.

resource "apstra_raw_json" "example" {
  url = "/api/resources/ip-pools"
  payload = jsonencode(
    {
      display_name = "test pool"
      subnets = [
        {
          network = "10.0.0.0/22"
        },
      ]
    }
  )
}