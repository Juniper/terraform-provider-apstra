# Importing a apstra_datacenter_external_gateway requires expressing both
# the blueprint ID and the external gateway ID in a JSON document:
#
# {
#   "blueprint_id": "007723b7-a387-4bb3-8a5e-b5e9f265de0d",
#   "external_gateway_id": "3zxDY0C8M0Y2m-xQFJQ"
# }

# Legacy import:

echo 'resource "apstra_datacenter_external_gateway" "legacy_import" {}' >> legacy_import.tf
terraform import 'apstra_datacenter_external_gateway.legacy_import' '{"blueprint_id":"007723b7-a387-4bb3-8a5e-b5e9f265de0d","external_gateway_id":"3zxDY0C8M0Y2m-xQFJQ"}'

# Terraform 1.5+ block import:

cat >> block_import.tf << EOF
import {
  to = apstra_datacenter_external_gateway.imported
  id = "{\"blueprint_id\":\"007723b7-a387-4bb3-8a5e-b5e9f265de0d\",\"external_gateway_id\":\"3zxDY0C8M0Y2m-xQFJQ\"}"
}
EOF
terraform plan -generate-config-out=generated.tf
terraform apply
