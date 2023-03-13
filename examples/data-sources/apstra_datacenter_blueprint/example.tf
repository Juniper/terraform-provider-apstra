# This example looks up a blueprint by name and uses the returned ID
# and build status fields to deploy the blueprint as needed, so long
# as it's error-free.

# Look up the blueprint by name.
data "apstra_datacenter_blueprint" "status" {
  name = "terraform commit example"
}

# Evaluate relevant fields from the returned status, determine
# whether it's safe/appropriate to attempt a deployment.
locals {
  has_changes = data.apstra_datacenter_blueprint.status.has_uncommitted_changes
  error_free = data.apstra_datacenter_blueprint.status.build_errors_count == 0
  build_when = local.has_changes && local.error_free
}

# Conditionally deploy the blueprint based on the assessment above.
resource "apstra_blueprint_deployment" "as_needed" {
  count = local.build_when ? 1 : 0
  blueprint_id = data.apstra_datacenter_blueprint.status.id
  comment      = "Deployment by Terraform {{`{{.TerraformVersion}}`}}, Apstra provider {{`{{.ProviderVersion}}`}}, User $USER."
}
