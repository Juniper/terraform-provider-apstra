# This example uses the 'apstra_blueprints' and 'apstra_blueprint_deployment'
# data sources to output a report comparing the staging and active blueprints.

# First, get a list of all blueprint IDs
data "apstra_blueprints" "blueprint_ids" {
  reference_design = "datacenter"
}

# Next, use the 'one()' function to assert that only one blueprint exists, and
# pull the commit info for that blueprint.
data "apstra_blueprint_deployment" "time_voyager" {
  blueprint_id = one(data.apstra_blueprints.blueprint_ids.ids)
}

# local variables pre-stage some of the calculations
# and make the output stage more readable.
locals {
  active = data.apstra_blueprint_deployment.time_voyager.revision_active
  staged = data.apstra_blueprint_deployment.time_voyager.revision_staged
  change_count = local.staged - local.active
  changed = data.apstra_blueprint_deployment.time_voyager.has_uncommitted_changes
  commit_or_commits = local.change_count == 1 ? "commit" : "commits"
  change_msg = "active blueprint is ${local.change_count} ${local.commit_or_commits} behind the staging blueprint"
  no_change_msg = "blueprint has no uncommitted changes"
}

# Use a conditional expression to output one of the prepared messages
output "blueprint_status_report" {
  value = local.changed ? local.change_msg : local.no_change_msg
}
