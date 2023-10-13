# This example returns the IDs of all Routing Policies in the
# Datacenter 1 blueprint which
# - import 10.140/16 and 10.150/16
#   ...or...
# - import 10.240/16 and 10.250/16, and export loopbacks
#
# All attributes specified within a 'filter' block must match.
# If an routing policy is found to match all attributes within
# a filter block, its ID will be included in the computed `ids`
# attribute.
data "apstra_datacenter_blueprint" "DC1" {
  name = "Datacenter 1"
}

data "apstra_datacenter_routing_policies" "all" {
  blueprint_id = data.apstra_datacenter_blueprint.DC1.id
  filters = [
    {
      extra_imports = [
        { prefix = "10.140.0.0/16", action = "permit" },
        { prefix = "10.150.0.0/16", action = "permit" },
      ]
    },
    {
      extra_imports = [
        { prefix = "10.240.0.0/16", action = "permit" },
        { prefix = "10.250.0.0/16", action = "permit" },
      ]
      export_policy = {
        export_loopbacks              = true
      }
    },
  ]
}
