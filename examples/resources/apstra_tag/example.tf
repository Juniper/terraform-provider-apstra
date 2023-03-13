# this example creates a tags named after enterprise teams
# responsible for various data center asset types.
locals {
  device_owners = toset([
    "research",
    "security",
    "app team 1",
    "app team 2",
  ])
}

resource "apstra_tag" "example" {
  for_each    = local.device_owners
  name        = each.key
  description = format("device maintained by %q team", each.key)
}