# This example uses the 'apstra_version' data source to conditionally create
# a resource which requires Apstra 6.2.0 based on the observed Apstra
# version.

data "apstra_version" "ours" {
  checks = {
    # the 'apstra_datacenter_switching_zone' resource requires Apstra 6.2.0 or later
    switching_zone_ok      = ">=6.2.0"

    # some other examples of version constraint checks
    version_50x_a          = ">=5.0.0,<5.1.0"
    version_50x_b          = "5.0"
    version_6x_but_not_612 = "6,!=6.1.2"
  }
}

resource "apstra_datacenter_switching_zone" "maybe" {
  # create 1 instance of this resource if the Apstra version is 6.2.0 or later, otherwise create 0 instances
  count        = data.apstra_version.ours.results["switching_zone_ok"] ? 1 : 0
  blueprint_id = local.blueprint_id
  name         = "maybe"
  mac_vrf_name = "maybe"
  service_type = "vlan_aware"
}

# The data source output looks like this:
# {
#   "checks" = {
#     "switching_zone_ok"      = ">=6.2.0"
#     "version_50x_a"          = ">=5.0.0,<5.1"
#     "version_50x_b"          = "5.0"
#     "version_6x_but_not_612" = "6,!=6.1.2"
#   }
#   "results" = {
#     "switching_zone_ok"      = true
#     "version_50x_a"          = false
#     "version_50x_b"          = false
#     "version_6x_but_not_612" = false
#   }
#   "version" = "6.2.0"
# }
